package launcher_test

import (
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestLaunchPlanRepresentsCompleteEditorLaunchOperation(t *testing.T) {
	context := devcontext.Context{
		ID:   devcontext.MustID("client-a"),
		Name: "Client A",
		Tool: codingtool.DefaultLaunchTarget(),
		Providers: provider.Configs{
			"claude": {Enabled: true},
			"codex":  {Enabled: true},
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}

	plan := launcher.LaunchPlan{
		ProjectPath:      project.Path("/work/client-a/api"),
		Context:          context,
		Tool:             context.Tool.ConfigFor(context.Tool.DefaultTool),
		Executable:       launcher.Executable("/usr/local/bin/code"),
		Arguments:        launcher.Arguments{"/work/client-a/api"},
		WorkingDirectory: launcher.WorkingDirectory("/work/client-a/api"),
		Environment: launcher.Environment{
			"CLAUDE_CONFIG_DIR": "/home/alex/.devctx/contexts/client-a/providers/claude",
			"CODEX_HOME":        "/home/alex/.devctx/contexts/client-a/providers/codex",
			"DEVCTX_CONTEXT":    "client-a",
		},
		Warnings: []launcher.ResolutionWarning{
			{
				Code:    "example_warning",
				Message: "example warning message",
			},
		},
		ResolutionSource: launcher.ResolutionSourceUserSelection,
	}

	want := launcher.LaunchPlan{
		ProjectPath:      project.Path("/work/client-a/api"),
		Context:          context,
		Tool:             codingtool.DefaultConfig(),
		Executable:       launcher.Executable("/usr/local/bin/code"),
		Arguments:        launcher.Arguments{"/work/client-a/api"},
		WorkingDirectory: launcher.WorkingDirectory("/work/client-a/api"),
		Environment: launcher.Environment{
			"CLAUDE_CONFIG_DIR": "/home/alex/.devctx/contexts/client-a/providers/claude",
			"CODEX_HOME":        "/home/alex/.devctx/contexts/client-a/providers/codex",
			"DEVCTX_CONTEXT":    "client-a",
		},
		Warnings: []launcher.ResolutionWarning{
			{
				Code:    "example_warning",
				Message: "example warning message",
			},
		},
		ResolutionSource: launcher.ResolutionSourceUserSelection,
	}

	if !reflect.DeepEqual(plan, want) {
		t.Fatalf("launch plan = %#v, want %#v", plan, want)
	}
}
