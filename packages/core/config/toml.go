package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"

	codingtool "devctx/packages/core/codingtool"
)

var (
	// ErrInvalidGlobalConfig identifies structurally invalid global config data.
	ErrInvalidGlobalConfig = errors.New("invalid global configuration")

	// ErrUnsupportedSchemaVersion identifies config data written for an
	// unsupported schema version.
	ErrUnsupportedSchemaVersion = errors.New("unsupported global configuration schema version")
)

const globalConfigRecoveryGuidance = "fix config.toml or move it aside so Dev Context can create a new default configuration"

// GlobalConfigFileError describes an invalid global configuration file without
// exposing its contents.
type GlobalConfigFileError struct {
	Path     string
	Cause    string
	Recovery string
	Err      error
}

func (e *GlobalConfigFileError) Error() string {
	return fmt.Sprintf("global configuration %q is invalid: %s; %s", e.Path, e.Cause, e.Recovery)
}

func (e *GlobalConfigFileError) Unwrap() error {
	return e.Err
}

type globalConfigTOML struct {
	Version       *int               `toml:"version"`
	DefaultEditor *string            `toml:"default_editor"`
	UI            uiSettingsTOML     `toml:"ui"`
	Safety        safetySettingsTOML `toml:"safety"`
}

type uiSettingsTOML struct {
	RememberWindowPosition *bool `toml:"remember_window_position"`
}

type safetySettingsTOML struct {
	WarnOnContextMismatch  *bool `toml:"warn_on_context_mismatch"`
	ConfirmUnboundProjects *bool `toml:"confirm_unbound_projects"`
}

// DecodeGlobalConfigTOML decodes TOML bytes into a typed global configuration.
func DecodeGlobalConfigTOML(data []byte) (GlobalConfig, error) {
	var raw globalConfigTOML
	metadata, err := toml.Decode(string(data), &raw)
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("%w: %w", ErrInvalidGlobalConfig, err)
	}

	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return GlobalConfig{}, fmt.Errorf("%w: unsupported field %q", ErrInvalidGlobalConfig, undecoded[0].String())
	}

	return globalConfigFromTOML(raw)
}

// DecodeGlobalConfigFileTOML decodes TOML bytes and wraps failures with
// path-specific recovery guidance.
func DecodeGlobalConfigFileTOML(path string, data []byte) (GlobalConfig, error) {
	globalConfig, err := DecodeGlobalConfigTOML(data)
	if err != nil {
		return GlobalConfig{}, newGlobalConfigFileError(path, err)
	}
	return globalConfig, nil
}

// EncodeGlobalConfigTOML encodes global configuration in deterministic TOML.
func EncodeGlobalConfigTOML(globalConfig GlobalConfig) ([]byte, error) {
	if err := validateGlobalConfig(globalConfig); err != nil {
		return nil, err
	}

	var builder strings.Builder
	builder.WriteString("version = ")
	builder.WriteString(strconv.Itoa(int(globalConfig.Version)))
	builder.WriteString("\n\n")
	builder.WriteString("default_editor = ")
	builder.WriteString(strconv.Quote(string(globalConfig.DefaultEditor)))
	builder.WriteString("\n\n")
	builder.WriteString("[ui]\n")
	builder.WriteString("remember_window_position = ")
	builder.WriteString(strconv.FormatBool(globalConfig.UI.RememberWindowPosition))
	builder.WriteString("\n\n")
	builder.WriteString("[safety]\n")
	builder.WriteString("warn_on_context_mismatch = ")
	builder.WriteString(strconv.FormatBool(globalConfig.Safety.WarnOnContextMismatch))
	builder.WriteString("\n")
	builder.WriteString("confirm_unbound_projects = ")
	builder.WriteString(strconv.FormatBool(globalConfig.Safety.ConfirmUnboundProjects))
	builder.WriteString("\n")

	return []byte(builder.String()), nil
}

func globalConfigFromTOML(raw globalConfigTOML) (GlobalConfig, error) {
	if raw.Version == nil {
		return GlobalConfig{}, fmt.Errorf("%w: missing version", ErrInvalidGlobalConfig)
	}
	version := SchemaVersion(*raw.Version)
	if version != CurrentSchemaVersion {
		return GlobalConfig{}, fmt.Errorf("%w: %d", ErrUnsupportedSchemaVersion, version)
	}

	if raw.DefaultEditor == nil {
		return GlobalConfig{}, fmt.Errorf("%w: missing default_editor", ErrInvalidGlobalConfig)
	}
	defaultEditor := codingtool.Type(*raw.DefaultEditor)
	if defaultEditor != codingtool.TypeVSCode {
		return GlobalConfig{}, fmt.Errorf("%w: unsupported default_editor %q", ErrInvalidGlobalConfig, defaultEditor)
	}

	if raw.UI.RememberWindowPosition == nil {
		return GlobalConfig{}, fmt.Errorf("%w: missing ui.remember_window_position", ErrInvalidGlobalConfig)
	}
	if raw.Safety.WarnOnContextMismatch == nil {
		return GlobalConfig{}, fmt.Errorf("%w: missing safety.warn_on_context_mismatch", ErrInvalidGlobalConfig)
	}
	if raw.Safety.ConfirmUnboundProjects == nil {
		return GlobalConfig{}, fmt.Errorf("%w: missing safety.confirm_unbound_projects", ErrInvalidGlobalConfig)
	}

	return GlobalConfig{
		Version:       version,
		DefaultEditor: defaultEditor,
		UI: UISettings{
			RememberWindowPosition: *raw.UI.RememberWindowPosition,
		},
		Safety: SafetySettings{
			WarnOnContextMismatch:  *raw.Safety.WarnOnContextMismatch,
			ConfirmUnboundProjects: *raw.Safety.ConfirmUnboundProjects,
		},
	}, nil
}

func validateGlobalConfig(globalConfig GlobalConfig) error {
	if globalConfig.Version != CurrentSchemaVersion {
		return fmt.Errorf("%w: %d", ErrUnsupportedSchemaVersion, globalConfig.Version)
	}
	if globalConfig.DefaultEditor != codingtool.TypeVSCode {
		return fmt.Errorf("%w: unsupported default_editor %q", ErrInvalidGlobalConfig, globalConfig.DefaultEditor)
	}
	return nil
}

func newGlobalConfigFileError(path string, err error) error {
	cause := "the file is not valid config.toml"
	switch {
	case errors.Is(err, ErrUnsupportedSchemaVersion):
		cause = "the schema version is not supported by this version of Dev Context"
	case errors.Is(err, ErrInvalidGlobalConfig):
		cause = "the file is malformed or missing required settings"
	}

	return &GlobalConfigFileError{
		Path:     path,
		Cause:    cause,
		Recovery: globalConfigRecoveryGuidance,
		Err:      err,
	}
}
