package editor_test

import (
	"testing"

	"devctx/packages/core/editor"
)

func TestDefaultConfigUsesVSCodeWithoutExecutableOverride(t *testing.T) {
	config := editor.DefaultConfig()

	if config.Type != editor.TypeVSCode {
		t.Fatalf("editor type = %q, want %q", config.Type, editor.TypeVSCode)
	}
	if config.ExecutableOverride != "" {
		t.Fatalf("executable override = %q, want empty", config.ExecutableOverride)
	}
}

func TestConfigStoresCustomExecutableOverride(t *testing.T) {
	config := editor.Config{
		Type:               editor.TypeVSCode,
		ExecutableOverride: "/opt/visual-studio-code/bin/code",
	}

	if config.Type != editor.TypeVSCode {
		t.Fatalf("editor type = %q, want %q", config.Type, editor.TypeVSCode)
	}
	if config.ExecutableOverride != "/opt/visual-studio-code/bin/code" {
		t.Fatalf("executable override = %q, want %q", config.ExecutableOverride, "/opt/visual-studio-code/bin/code")
	}
}
