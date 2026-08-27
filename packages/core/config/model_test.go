package config_test

import (
	"testing"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
)

func TestGlobalConfigConstructsVersionedApplicationSettings(t *testing.T) {
	globalConfig := config.GlobalConfig{
		Version:     config.CurrentSchemaVersion,
		DefaultTool: codingtool.TypeVSCode,
		UI: config.UISettings{
			RememberWindowPosition: true,
		},
		Safety: config.SafetySettings{
			WarnOnContextMismatch:  true,
			ConfirmUnboundProjects: true,
		},
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
