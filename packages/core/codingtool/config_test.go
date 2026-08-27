package codingtool_test

import (
	"reflect"
	"testing"

	codingtool "devctx/packages/core/codingtool"
)

func TestDefaultLaunchTargetUsesRegistryDefault(t *testing.T) {
	target := codingtool.DefaultLaunchTarget()

	if target.DefaultTool != codingtool.VSCodeID {
		t.Fatalf("default tool = %q, want %q", target.DefaultTool, codingtool.VSCodeID)
	}
	if !reflect.DeepEqual(target.Tools, map[codingtool.ID]codingtool.Config{codingtool.VSCodeID: {}}) {
		t.Fatalf("tool settings = %#v", target.Tools)
	}
}

func TestLaunchTargetStoresSettingsPerTool(t *testing.T) {
	target := codingtool.LaunchTarget{
		DefaultTool: codingtool.VSCodeID,
		Tools: map[codingtool.ID]codingtool.Config{
			codingtool.VSCodeID: {ExecutableOverride: "/opt/visual-studio-code/bin/code"},
			"cursor":            {Options: map[string]string{"profile": "work"}},
		},
	}

	if target.ConfigFor(codingtool.VSCodeID).ExecutableOverride != "/opt/visual-studio-code/bin/code" {
		t.Fatal("default tool configuration was not returned")
	}
	if target.ConfigFor("cursor").Options["profile"] != "work" {
		t.Fatal("non-default tool configuration was not returned")
	}
}
