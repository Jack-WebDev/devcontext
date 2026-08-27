package context_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
)

func TestContextTOMLRoundTripPreservesLaunchTarget(t *testing.T) {
	ctx := devcontext.Context{
		ID:   devcontext.MustID("client-a"),
		Name: "Client A",
		Tool: codingtool.LaunchTarget{
			DefaultTool: "cursor",
			Tools: map[codingtool.ID]codingtool.Config{
				"cursor": {ExecutableOverride: "/opt/cursor", Options: map[string]string{"profile": "client-a"}},
				"vscode": {Options: map[string]string{"extensions": "stable"}},
			},
		},
		CreatedAt: time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC),
	}

	data, err := devcontext.EncodeContextTOML(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := "id = \"client-a\"\nname = \"Client A\"\ncreated_at = 2026-08-13T11:00:00Z\n\n[launch_target]\ndefault_tool = \"cursor\"\n\n[launch_target.tools.cursor]\nexecutable_override = \"/opt/cursor\"\n\n[launch_target.tools.cursor.options]\nprofile = \"client-a\"\n\n[launch_target.tools.vscode]\n\n[launch_target.tools.vscode.options]\nextensions = \"stable\"\n"
	if string(data) != want {
		t.Fatalf("TOML =\n%s\nwant:\n%s", data, want)
	}

	decoded, err := devcontext.DecodeContextTOML(data, ctx.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, ctx) {
		t.Fatalf("decoded = %#v, want %#v", decoded, ctx)
	}
}

func TestDecodeContextTOMLRejectsLegacyEditorConfig(t *testing.T) {
	_, err := devcontext.DecodeContextTOML([]byte("id = \"personal\"\nname = \"Personal\"\ncreated_at = 2026-08-13T10:30:00Z\n\n[editor]\ntype = \"vscode\"\n"), devcontext.MustID("personal"))
	if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
		t.Fatalf("error = %v, want invalid context config", err)
	}
}

func TestContextTOMLRequiresConfiguredDefaultTool(t *testing.T) {
	ctx := devcontext.Context{ID: devcontext.MustID("personal"), Name: "Personal", Tool: codingtool.LaunchTarget{DefaultTool: "vscode"}, CreatedAt: time.Now()}
	_, err := devcontext.EncodeContextTOML(ctx)
	if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
		t.Fatalf("error = %v, want invalid context config", err)
	}
}
