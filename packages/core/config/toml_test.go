package config_test

import (
	"errors"
	"strings"
	"testing"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
)

func TestDecodeGlobalConfigTOMLLoadsDocumentedExample(t *testing.T) {
	input := []byte(`
version = 1

default_editor = "vscode"

[ui]
remember_window_position = true

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`)

	globalConfig, err := config.DecodeGlobalConfigTOML(input)
	if err != nil {
		t.Fatalf("decode global config: %v", err)
	}

	if globalConfig.Version != config.CurrentSchemaVersion {
		t.Fatalf("version = %d, want %d", globalConfig.Version, config.CurrentSchemaVersion)
	}
	if globalConfig.DefaultTool != codingtool.TypeVSCode {
		t.Fatalf("default editor = %q, want %q", globalConfig.DefaultTool, codingtool.TypeVSCode)
	}
	if !globalConfig.UI.RememberWindowPosition {
		t.Fatal("remember window position = false, want true")
	}
	if !globalConfig.Safety.WarnOnContextMismatch {
		t.Fatal("warn on context mismatch = false, want true")
	}
	if !globalConfig.Safety.ConfirmUnboundProjects {
		t.Fatal("confirm unbound projects = false, want true")
	}
}

func TestDecodeGlobalConfigTOMLRejectsUnsupportedSchemaVersion(t *testing.T) {
	input := []byte(`
version = 2
default_editor = "vscode"

[ui]
remember_window_position = true

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`)

	_, err := config.DecodeGlobalConfigTOML(input)
	if !errors.Is(err, config.ErrUnsupportedSchemaVersion) {
		t.Fatalf("error = %v, want %v", err, config.ErrUnsupportedSchemaVersion)
	}
}

func TestDecodeGlobalConfigTOMLRejectsStructurallyInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMessage string
	}{
		{
			name:        "malformed toml",
			input:       `version = `,
			wantMessage: "invalid global configuration",
		},
		{
			name: "missing version",
			input: `
default_editor = "vscode"

[ui]
remember_window_position = true

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`,
			wantMessage: "missing version",
		},
		{
			name: "unsupported default editor",
			input: `
version = 1
default_editor = "unknown"

[ui]
remember_window_position = true

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`,
			wantMessage: `unsupported default_editor "unknown"`,
		},
		{
			name: "missing ui setting",
			input: `
version = 1
default_editor = "vscode"

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`,
			wantMessage: "missing ui.remember_window_position",
		},
		{
			name: "wrong value type",
			input: `
version = 1
default_editor = "vscode"

[ui]
remember_window_position = "yes"

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`,
			wantMessage: "invalid global configuration",
		},
		{
			name: "unknown field",
			input: `
version = 1
default_editor = "vscode"
unknown = true

[ui]
remember_window_position = true

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = true
`,
			wantMessage: `unsupported field "unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := config.DecodeGlobalConfigTOML([]byte(tt.input))
			if !errors.Is(err, config.ErrInvalidGlobalConfig) {
				t.Fatalf("error = %v, want %v", err, config.ErrInvalidGlobalConfig)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestDecodeGlobalConfigFileTOMLReportsActionableCorruptConfigError(t *testing.T) {
	path := "/home/alex/.devctx/config.toml"

	_, err := config.DecodeGlobalConfigFileTOML(path, []byte(`version = `))
	if err == nil {
		t.Fatal("error is nil")
	}
	if !errors.Is(err, config.ErrInvalidGlobalConfig) {
		t.Fatalf("error = %v, want %v", err, config.ErrInvalidGlobalConfig)
	}

	var fileErr *config.GlobalConfigFileError
	if !errors.As(err, &fileErr) {
		t.Fatalf("error type = %T, want *config.GlobalConfigFileError", err)
	}
	if fileErr.Path != path {
		t.Fatalf("path = %q, want %q", fileErr.Path, path)
	}
	if !strings.Contains(err.Error(), path) {
		t.Fatalf("error = %q, want path %q", err.Error(), path)
	}
	if !strings.Contains(err.Error(), "malformed or missing required settings") {
		t.Fatalf("error = %q, want likely cause", err.Error())
	}
	if !strings.Contains(err.Error(), "fix config.toml or move it aside") {
		t.Fatalf("error = %q, want corrective action", err.Error())
	}
	if strings.Contains(err.Error(), "expected value") {
		t.Fatalf("error = %q, should not expose raw parser details", err.Error())
	}
}

func TestEncodeGlobalConfigTOMLIsDeterministic(t *testing.T) {
	globalConfig := config.GlobalConfig{
		Version:     config.CurrentSchemaVersion,
		DefaultTool: codingtool.TypeVSCode,
		UI: config.UISettings{
			RememberWindowPosition: false,
		},
		Safety: config.SafetySettings{
			WarnOnContextMismatch:  true,
			ConfirmUnboundProjects: false,
		},
	}

	encoded, err := config.EncodeGlobalConfigTOML(globalConfig)
	if err != nil {
		t.Fatalf("encode global config: %v", err)
	}

	want := `version = 1

default_editor = "vscode"

[ui]
remember_window_position = false
close_after_launch = false
launch_verification = false
remember_projects = false
tray_enabled = false

[safety]
warn_on_context_mismatch = true
confirm_unbound_projects = false
`
	if string(encoded) != want {
		t.Fatalf("encoded TOML =\n%s\nwant:\n%s", encoded, want)
	}
}

func TestEncodeGlobalConfigTOMLRoundTripsThroughDecoder(t *testing.T) {
	globalConfig := config.DefaultGlobalConfig()

	encoded, err := config.EncodeGlobalConfigTOML(globalConfig)
	if err != nil {
		t.Fatalf("encode global config: %v", err)
	}

	decoded, err := config.DecodeGlobalConfigTOML(encoded)
	if err != nil {
		t.Fatalf("decode encoded global config: %v", err)
	}

	if decoded != globalConfig {
		t.Fatalf("decoded config = %#v, want %#v", decoded, globalConfig)
	}
}

func TestEncodeGlobalConfigTOMLRejectsUnsupportedValues(t *testing.T) {
	_, err := config.EncodeGlobalConfigTOML(config.GlobalConfig{
		Version:     config.SchemaVersion(99),
		DefaultTool: codingtool.TypeVSCode,
	})
	if !errors.Is(err, config.ErrUnsupportedSchemaVersion) {
		t.Fatalf("error = %v, want %v", err, config.ErrUnsupportedSchemaVersion)
	}

	_, err = config.EncodeGlobalConfigTOML(config.GlobalConfig{
		Version:     config.CurrentSchemaVersion,
		DefaultTool: codingtool.Type("unknown"),
	})
	if !errors.Is(err, config.ErrInvalidGlobalConfig) {
		t.Fatalf("error = %v, want %v", err, config.ErrInvalidGlobalConfig)
	}
}
