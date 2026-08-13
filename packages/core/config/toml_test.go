package config_test

import (
	"errors"
	"strings"
	"testing"

	"devctx/packages/core/config"
	"devctx/packages/core/editor"
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
	if globalConfig.DefaultEditor != editor.TypeVSCode {
		t.Fatalf("default editor = %q, want %q", globalConfig.DefaultEditor, editor.TypeVSCode)
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
