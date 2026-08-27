package context

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/provider"
)

var (
	// ErrInvalidContextConfig identifies structurally invalid context config data.
	ErrInvalidContextConfig = errors.New("invalid context configuration")

	// ErrContextIDMismatch identifies context files whose ID does not match the
	// directory they were loaded from.
	ErrContextIDMismatch = errors.New("context ID does not match directory")
)

type contextTOML struct {
	ID        *string                 `toml:"id"`
	Name      *string                 `toml:"name"`
	CreatedAt *time.Time              `toml:"created_at"`
	Tool      launchTargetTOML        `toml:"launch_target"`
	Providers map[string]providerTOML `toml:"providers"`
	Metadata  map[string]string       `toml:"metadata"`
}

type launchTargetTOML struct {
	DefaultTool *string                   `toml:"default_tool"`
	Tools       map[string]toolConfigTOML `toml:"tools"`
}

type toolConfigTOML struct {
	ExecutableOverride *string           `toml:"executable_override"`
	Options            map[string]string `toml:"options"`
}

type providerTOML struct {
	Enabled *bool             `toml:"enabled"`
	Options map[string]string `toml:"options"`
}

// DecodeContextTOML decodes TOML bytes into a typed context configuration.
//
// expectedID must be the ID derived from the context directory. The decoded
// file ID must match it.
func DecodeContextTOML(data []byte, expectedID ID) (Context, error) {
	if expectedID.String() == "" {
		return Context{}, fmt.Errorf("%w: missing expected context ID", ErrInvalidContextConfig)
	}

	var raw contextTOML
	metadata, err := toml.Decode(string(data), &raw)
	if err != nil {
		return Context{}, fmt.Errorf("%w: %w", ErrInvalidContextConfig, err)
	}

	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return Context{}, fmt.Errorf("%w: unsupported field %q", ErrInvalidContextConfig, undecoded[0].String())
	}

	return contextFromTOML(raw, expectedID)
}

// EncodeContextTOML encodes a context in deterministic TOML.
func EncodeContextTOML(ctx Context) ([]byte, error) {
	if err := validateContextForTOML(ctx); err != nil {
		return nil, err
	}

	var builder strings.Builder
	writeStringValue(&builder, "id", ctx.ID.String())
	writeStringValue(&builder, "name", ctx.Name)
	writeTimeValue(&builder, "created_at", ctx.CreatedAt)

	builder.WriteString("\n[launch_target]\n")
	writeStringValue(&builder, "default_tool", string(ctx.Tool.DefaultTool))
	for _, toolID := range sortedToolIDs(ctx.Tool.Tools) {
		toolConfig := ctx.Tool.Tools[toolID]
		builder.WriteString("\n[launch_target.tools.")
		builder.WriteString(tomlKeySegment(string(toolID)))
		builder.WriteString("]\n")
		if toolConfig.ExecutableOverride != "" {
			writeStringValue(&builder, "executable_override", toolConfig.ExecutableOverride)
		}
		if len(toolConfig.Options) > 0 {
			builder.WriteString("\n[launch_target.tools.")
			builder.WriteString(tomlKeySegment(string(toolID)))
			builder.WriteString(".options]\n")
			for _, key := range sortedOptionKeys(toolConfig.Options) {
				writeStringValue(&builder, key, toolConfig.Options[key])
			}
		}
	}

	for _, providerID := range sortedProviderIDs(ctx.Providers) {
		providerConfig := ctx.Providers[providerID]
		builder.WriteString("\n[providers.")
		builder.WriteString(tomlKeySegment(string(providerID)))
		builder.WriteString("]\n")
		builder.WriteString("enabled = ")
		builder.WriteString(strconv.FormatBool(providerConfig.Enabled))
		builder.WriteString("\n")

		if len(providerConfig.Options) == 0 {
			continue
		}
		builder.WriteString("\n[providers.")
		builder.WriteString(tomlKeySegment(string(providerID)))
		builder.WriteString(".options]\n")
		for _, key := range sortedOptionKeys(providerConfig.Options) {
			writeStringValue(&builder, key, providerConfig.Options[key])
		}
	}

	if len(ctx.Metadata) > 0 {
		builder.WriteString("\n[metadata]\n")
		for _, key := range sortedMetadataKeys(ctx.Metadata) {
			writeStringValue(&builder, key, ctx.Metadata[key])
		}
	}

	return []byte(builder.String()), nil
}

func contextFromTOML(raw contextTOML, expectedID ID) (Context, error) {
	if raw.ID == nil {
		return Context{}, fmt.Errorf("%w: missing id", ErrInvalidContextConfig)
	}
	id, err := NewID(*raw.ID)
	if err != nil {
		return Context{}, fmt.Errorf("%w: %w", ErrInvalidContextConfig, err)
	}
	if id != expectedID {
		return Context{}, fmt.Errorf("%w: file id %q, directory id %q", ErrContextIDMismatch, id.String(), expectedID.String())
	}

	if raw.Name == nil {
		return Context{}, fmt.Errorf("%w: missing name", ErrInvalidContextConfig)
	}
	if *raw.Name == "" {
		return Context{}, fmt.Errorf("%w: name cannot be empty", ErrInvalidContextConfig)
	}

	if raw.CreatedAt == nil {
		return Context{}, fmt.Errorf("%w: missing created_at", ErrInvalidContextConfig)
	}
	if raw.CreatedAt.IsZero() {
		return Context{}, fmt.Errorf("%w: created_at cannot be zero", ErrInvalidContextConfig)
	}

	toolConfig, err := launchTargetFromTOML(raw.Tool)
	if err != nil {
		return Context{}, err
	}
	providerConfigs, err := providerConfigsFromTOML(raw.Providers)
	if err != nil {
		return Context{}, err
	}
	metadata, err := metadataFromTOML(raw.Metadata)
	if err != nil {
		return Context{}, err
	}

	return Context{
		ID:        id,
		Name:      *raw.Name,
		Tool:      toolConfig,
		Providers: providerConfigs,
		Metadata:  metadata,
		CreatedAt: raw.CreatedAt.UTC(),
	}, nil
}

func launchTargetFromTOML(raw launchTargetTOML) (codingtool.LaunchTarget, error) {
	if raw.DefaultTool == nil {
		return codingtool.LaunchTarget{}, fmt.Errorf("%w: missing launch_target.default_tool", ErrInvalidContextConfig)
	}
	if *raw.DefaultTool == "" {
		return codingtool.LaunchTarget{}, fmt.Errorf("%w: launch_target.default_tool cannot be empty", ErrInvalidContextConfig)
	}

	configs := make(map[codingtool.ID]codingtool.Config, len(raw.Tools)+1)
	for toolID, rawConfig := range raw.Tools {
		if toolID == "" {
			return codingtool.LaunchTarget{}, fmt.Errorf("%w: tool ID cannot be empty", ErrInvalidContextConfig)
		}
		options, err := toolOptionsFromTOML(toolID, rawConfig.Options)
		if err != nil {
			return codingtool.LaunchTarget{}, err
		}
		config := codingtool.Config{Options: options}
		if rawConfig.ExecutableOverride != nil {
			config.ExecutableOverride = *rawConfig.ExecutableOverride
		}
		configs[codingtool.ID(toolID)] = config
	}
	defaultTool := codingtool.ID(*raw.DefaultTool)
	if _, ok := configs[defaultTool]; !ok {
		configs[defaultTool] = codingtool.Config{}
	}
	return codingtool.LaunchTarget{DefaultTool: defaultTool, Tools: configs}, nil
}

func toolOptionsFromTOML(toolID string, raw map[string]string) (map[string]string, error) {
	if raw == nil {
		return nil, nil
	}
	options := make(map[string]string, len(raw))
	for key, value := range raw {
		if key == "" {
			return nil, fmt.Errorf("%w: tool %q option key cannot be empty", ErrInvalidContextConfig, toolID)
		}
		options[key] = value
	}
	return options, nil
}

func providerConfigsFromTOML(raw map[string]providerTOML) (provider.Configs, error) {
	if raw == nil {
		return nil, nil
	}

	providerConfigs := make(provider.Configs, len(raw))
	for providerID, rawProvider := range raw {
		if providerID == "" {
			return nil, fmt.Errorf("%w: provider ID cannot be empty", ErrInvalidContextConfig)
		}
		if rawProvider.Enabled == nil {
			return nil, fmt.Errorf("%w: missing providers.%s.enabled", ErrInvalidContextConfig, providerID)
		}

		options, err := providerOptionsFromTOML(providerID, rawProvider.Options)
		if err != nil {
			return nil, err
		}

		providerConfigs[provider.ID(providerID)] = provider.Config{
			Enabled: *rawProvider.Enabled,
			Options: options,
		}
	}
	return providerConfigs, nil
}

func providerOptionsFromTOML(providerID string, raw map[string]string) (provider.Options, error) {
	if raw == nil {
		return nil, nil
	}

	options := make(provider.Options, len(raw))
	for key, value := range raw {
		if key == "" {
			return nil, fmt.Errorf("%w: provider %q option key cannot be empty", ErrInvalidContextConfig, providerID)
		}
		options[key] = value
	}
	return options, nil
}

func metadataFromTOML(raw map[string]string) (Metadata, error) {
	if raw == nil {
		return nil, nil
	}

	metadata := make(Metadata, len(raw))
	for key, value := range raw {
		if key == "" {
			return nil, fmt.Errorf("%w: metadata key cannot be empty", ErrInvalidContextConfig)
		}
		metadata[key] = value
	}
	return metadata, nil
}

func validateContextForTOML(ctx Context) error {
	if ctx.ID.String() == "" {
		return fmt.Errorf("%w: missing id", ErrInvalidContextConfig)
	}
	if ctx.Name == "" {
		return fmt.Errorf("%w: name cannot be empty", ErrInvalidContextConfig)
	}
	if ctx.CreatedAt.IsZero() {
		return fmt.Errorf("%w: created_at cannot be zero", ErrInvalidContextConfig)
	}
	if ctx.Tool.DefaultTool == "" {
		return fmt.Errorf("%w: launch_target.default_tool cannot be empty", ErrInvalidContextConfig)
	}
	if _, ok := ctx.Tool.Tools[ctx.Tool.DefaultTool]; !ok {
		return fmt.Errorf("%w: missing configuration for default tool %q", ErrInvalidContextConfig, ctx.Tool.DefaultTool)
	}
	for toolID, toolConfig := range ctx.Tool.Tools {
		if toolID == "" {
			return fmt.Errorf("%w: tool ID cannot be empty", ErrInvalidContextConfig)
		}
		for key := range toolConfig.Options {
			if key == "" {
				return fmt.Errorf("%w: tool %q option key cannot be empty", ErrInvalidContextConfig, toolID)
			}
		}
	}

	for providerID, providerConfig := range ctx.Providers {
		if providerID == "" {
			return fmt.Errorf("%w: provider ID cannot be empty", ErrInvalidContextConfig)
		}
		for key := range providerConfig.Options {
			if key == "" {
				return fmt.Errorf("%w: provider %q option key cannot be empty", ErrInvalidContextConfig, providerID)
			}
		}
	}
	for key := range ctx.Metadata {
		if key == "" {
			return fmt.Errorf("%w: metadata key cannot be empty", ErrInvalidContextConfig)
		}
	}

	return nil
}

func writeStringValue(builder *strings.Builder, key string, value string) {
	builder.WriteString(tomlKeySegment(key))
	builder.WriteString(" = ")
	builder.WriteString(strconv.Quote(value))
	builder.WriteString("\n")
}

func writeTimeValue(builder *strings.Builder, key string, value time.Time) {
	builder.WriteString(tomlKeySegment(key))
	builder.WriteString(" = ")
	builder.WriteString(value.UTC().Format(time.RFC3339Nano))
	builder.WriteString("\n")
}

func tomlKeySegment(key string) string {
	if isBareTOMLKey(key) {
		return key
	}
	return strconv.Quote(key)
}

func isBareTOMLKey(key string) bool {
	if key == "" {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return false
		}
	}
	return true
}

func sortedProviderIDs(configs provider.Configs) []provider.ID {
	ids := make([]provider.ID, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func sortedToolIDs(configs map[codingtool.ID]codingtool.Config) []codingtool.ID {
	ids := make([]codingtool.ID, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func sortedOptionKeys(options provider.Options) []string {
	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedMetadataKeys(metadata Metadata) []string {
	keys := make([]string, 0, len(metadata))
	for key := range metadata {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
