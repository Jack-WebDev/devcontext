package config

import (
	"errors"
	"fmt"

	"github.com/BurntSushi/toml"

	"devctx/packages/core/editor"
)

var (
	// ErrInvalidGlobalConfig identifies structurally invalid global config data.
	ErrInvalidGlobalConfig = errors.New("invalid global configuration")

	// ErrUnsupportedSchemaVersion identifies config data written for an
	// unsupported schema version.
	ErrUnsupportedSchemaVersion = errors.New("unsupported global configuration schema version")
)

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
	defaultEditor := editor.Type(*raw.DefaultEditor)
	if defaultEditor != editor.TypeVSCode {
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
