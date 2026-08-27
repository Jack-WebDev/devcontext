package codingtool_test

import (
	"testing"

	codingtool "devctx/packages/core/codingtool"
)

func TestDefaultConfigUsesVSCodeWithoutExecutableOverride(t *testing.T) {
	config := codingtool.DefaultConfig()

	if config.Type != codingtool.TypeVSCode {
		t.Fatalf("editor type = %q, want %q", config.Type, codingtool.TypeVSCode)
	}
	if config.ExecutableOverride != "" {
		t.Fatalf("executable override = %q, want empty", config.ExecutableOverride)
	}
}

func TestConfigStoresCustomExecutableOverride(t *testing.T) {
	config := codingtool.Config{
		Type:               codingtool.TypeVSCode,
		ExecutableOverride: "/opt/visual-studio-code/bin/code",
	}

	if config.Type != codingtool.TypeVSCode {
		t.Fatalf("editor type = %q, want %q", config.Type, codingtool.TypeVSCode)
	}
	if config.ExecutableOverride != "/opt/visual-studio-code/bin/code" {
		t.Fatalf("executable override = %q, want %q", config.ExecutableOverride, "/opt/visual-studio-code/bin/code")
	}
}
