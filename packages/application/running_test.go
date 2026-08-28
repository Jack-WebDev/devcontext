package application

import (
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	coreRunning "devctx/packages/core/running"
)

func TestRunningEnvironmentStatePreservesSafeLaunchAndRuntimeMetadata(t *testing.T) {
	pid := 4128
	state := runningEnvironmentState(coreRunning.Environment{
		ID:        "environment-1",
		Project:   coreRunning.ProjectIdentity{Path: project.Path("/work/api"), Name: "api"},
		Context:   coreRunning.ContextIdentity{ID: devcontext.MustID("company"), Name: "Company"},
		Tool:      coreRunning.ToolIdentity{ID: codingtool.ID("second-tool"), Name: "Second Tool"},
		StartedAt: time.Date(2026, 8, 28, 10, 30, 0, 0, time.FixedZone("SAST", 2*60*60)),
		Process:   coreRunning.Process{PID: &pid, State: coreRunning.ProcessStateRunning},
		Session:   coreRunning.Session{ID: "session-1", State: coreRunning.SessionStateActive},
		Launch:    coreRunning.LaunchIdentity{Source: launcher.InvocationSourceGUI, ResolutionSource: launcher.ResolutionSourceExplicit},
	})

	if state.ID != "environment-1" || state.Project.Path != "/work/api" || state.Context.Name != "Company" || state.Tool.Name != "Second Tool" {
		t.Fatalf("running environment state = %#v", state)
	}
	if state.StartedAt.Location() != time.UTC || state.Process.PID == nil || *state.Process.PID != pid || state.Session.State != "active" {
		t.Fatalf("running environment runtime state = %#v", state)
	}
	if state.Launch.Source != "gui" || state.Launch.ResolutionSource != "explicit" {
		t.Fatalf("running environment launch state = %#v", state.Launch)
	}
}
