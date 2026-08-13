package context

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"devctx/packages/core/editor"
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
	Editor    editorTOML              `toml:"editor"`
	Providers map[string]providerTOML `toml:"providers"`
	Metadata  map[string]string       `toml:"metadata"`
}

type editorTOML struct {
	Type               *string `toml:"type"`
	ExecutableOverride *string `toml:"executable_override"`
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

	builder.WriteString("\n[editor]\n")
	writeStringValue(&builder, "type", string(ctx.Editor.Type))
	if ctx.Editor.ExecutableOverride != "" {
		writeStringValue(&builder, "executable_override", ctx.Editor.ExecutableOverride)
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

	editorConfig, err := editorConfigFromTOML(raw.Editor)
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
		Editor:    editorConfig,
		Providers: providerConfigs,
		Metadata:  metadata,
		CreatedAt: raw.CreatedAt.UTC(),
	}, nil
}

func editorConfigFromTOML(raw editorTOML) (editor.Config, error) {
	if raw.Type == nil {
		return editor.Config{}, fmt.Errorf("%w: missing editor.type", ErrInvalidContextConfig)
	}
	if *raw.Type == "" {
		return editor.Config{}, fmt.Errorf("%w: editor.type cannot be empty", ErrInvalidContextConfig)
	}

	editorConfig := editor.Config{
		Type: editor.Type(*raw.Type),
	}
	if raw.ExecutableOverride != nil {
		editorConfig.ExecutableOverride = *raw.ExecutableOverride
	}
	return editorConfig, nil
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
	if ctx.Editor.Type == "" {
		return fmt.Errorf("%w: editor.type cannot be empty", ErrInvalidContextConfig)
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
