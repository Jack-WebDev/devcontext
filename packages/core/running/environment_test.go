package running_test

import (
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/running"
)

func TestEnvironmentCapturesImmutableLaunchIdentityAndObservedState(t *testing.T) {
	pid := 4128
	environment := running.Environment{
		ID:        "environment-1",
		Project:   running.ProjectIdentity{Path: project.Path("/work/api"), Name: "api"},
		Context:   running.ContextIdentity{ID: devcontext.MustID("company"), Name: "Company"},
		Tool:      running.ToolIdentity{ID: codingtool.ID("second-tool"), Name: "Second Tool"},
		StartedAt: time.Date(2026, 8, 28, 10, 30, 0, 0, time.UTC),
		Process:   running.Process{PID: &pid, State: running.ProcessStateRunning},
		Session:   running.Session{ID: "session-1", State: running.SessionStateActive},
		Launch:    running.LaunchIdentity{Source: launcher.InvocationSourceGUI, ResolutionSource: launcher.ResolutionSourceExplicit},
	}

	if environment.Context.ID.String() != "company" || environment.Tool.Name != "Second Tool" {
		t.Fatalf("environment identity = %#v", environment)
	}
	if environment.Process.PID == nil || *environment.Process.PID != pid || environment.Session.State != running.SessionStateActive {
		t.Fatalf("environment runtime state = %#v", environment)
	}
}
