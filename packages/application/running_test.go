package application

import (
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
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

func TestGetRunningEnvironmentsRefreshesProcessStateAndRecordsStoppedEvent(t *testing.T) {
	fixture := newApplicationFixture(t)
	logger := &applicationRecordingLogger{}
	fixture.logger = logger
	activePID := 100
	stoppedPID := 200
	repository := coreRunning.NewRepository(fixture.runningPath)
	for _, environment := range []coreRunning.Environment{
		testRunningEnvironment(fixture, "active", activePID),
		testRunningEnvironment(fixture, "stopped", stoppedPID),
	} {
		if _, err := repository.Record(environment); err != nil {
			t.Fatalf("record environment: %v", err)
		}
	}
	service := fixture.service()
	service.dependencies.ProcessInspector = applicationProcessInspector{states: map[int]bool{activePID: true, stoppedPID: false}}

	state, appErr := service.GetRunningEnvironments()
	if appErr != nil {
		t.Fatalf("get running environments: %v", appErr)
	}
	if len(state.Environments) != 1 || state.Environments[0].Context.ID != "active" || state.Environments[0].Process.State != "running" {
		t.Fatalf("active running environments = %#v", state.Environments)
	}
	if got := applicationEventNames(logger.events); len(got) != 1 || got[0] != devlog.EventEnvironmentStopped {
		t.Fatalf("history events = %#v, want environment stopped", got)
	}
}

func testRunningEnvironment(fixture applicationFixture, contextID string, pid int) coreRunning.Environment {
	return coreRunning.Environment{
		Project:   coreRunning.ProjectIdentity{Path: project.Path(fixture.projectDir), Name: "current"},
		Context:   coreRunning.ContextIdentity{ID: devcontext.MustID(contextID), Name: contextID},
		Tool:      coreRunning.ToolIdentity{ID: fixture.editor.ID(), Name: "Fake Tool"},
		StartedAt: fixture.now,
		Process:   coreRunning.Process{PID: &pid, State: coreRunning.ProcessStateRunning},
		Session:   coreRunning.Session{State: coreRunning.SessionStateUnknown},
		Launch:    coreRunning.LaunchIdentity{Source: launcher.InvocationSourceGUI, ResolutionSource: launcher.ResolutionSourceExplicit},
	}
}

type applicationProcessInspector struct{ states map[int]bool }

func (i applicationProcessInspector) IsRunning(pid int) (bool, error) { return i.states[pid], nil }
