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
	service := newApplicationFixture(t).service()
	state := service.runningEnvironmentState(coreRunning.Environment{
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
	if state.Lifecycle.State != "active" || state.Lifecycle.Focusable || state.Lifecycle.Revealable || state.Lifecycle.Stoppable {
		t.Fatalf("workspace lifecycle = %#v", state.Lifecycle)
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
	if got := applicationEventNames(logger.events); len(got) != 1 || got[0] != devlog.EventWorkspaceStopped {
		t.Fatalf("history events = %#v, want environment stopped", got)
	}
}

func TestGetRunningEnvironmentsIncludesAnActiveAdapterSession(t *testing.T) {
	fixture := newApplicationFixture(t)
	environment := testRunningEnvironment(fixture, "session-active", 0)
	environment.Process = coreRunning.Process{State: coreRunning.ProcessStateUnknown}
	environment.Session = coreRunning.Session{ID: "adapter-session", State: coreRunning.SessionStateActive}
	if _, err := coreRunning.NewRepository(fixture.runningPath).Record(environment); err != nil {
		t.Fatalf("record environment: %v", err)
	}

	state, appErr := fixture.service().GetRunningEnvironments()
	if appErr != nil {
		t.Fatalf("get running environments: %v", appErr)
	}
	if len(state.Environments) != 1 || state.Environments[0].Lifecycle.State != "active" {
		t.Fatalf("active session workspaces = %#v", state.Environments)
	}
}

func TestUnknownWorkspaceIsVisibleButDoesNotBlockSafetyChecks(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	unknown := testRunningEnvironment(fixture, "personal", 0)
	unknown.Process = coreRunning.Process{State: coreRunning.ProcessStateUnknown}
	if _, err := coreRunning.NewRepository(fixture.runningPath).Record(unknown); err != nil {
		t.Fatalf("record unknown workspace: %v", err)
	}
	service := fixture.service()

	workspaces, appErr := service.GetRunningEnvironments()
	if appErr != nil || len(workspaces.Environments) != 1 {
		t.Fatalf("workspaces = %#v, %v", workspaces, appErr)
	}
	if lifecycle := workspaces.Environments[0].Lifecycle; lifecycle.State != "unknown" || lifecycle.Focusable || lifecycle.Revealable || lifecycle.Stoppable {
		t.Fatalf("unknown workspace lifecycle = %#v", lifecycle)
	}

	dashboard, appErr := service.GetHomeDashboard(GetHomeDashboardRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get home dashboard: %v", appErr)
	}
	if dashboard.Running.Count != 0 || dashboard.Running.IsolationProtected {
		t.Fatalf("running summary = %#v", dashboard.Running)
	}
	preflight, appErr := service.PreflightLaunchProject(PreflightLaunchProjectRequest{ProjectPath: ".", ContextID: "company"})
	if appErr != nil {
		t.Fatalf("preflight launch: %v", appErr)
	}
	if preflight.RunningEnvironmentConflict != nil {
		t.Fatalf("running conflict = %#v, want none", preflight.RunningEnvironmentConflict)
	}
	if _, appErr := service.ArchiveContext(ArchiveContextRequest{ContextID: "personal"}); appErr != nil {
		t.Fatalf("archive context: %v", appErr)
	}
	if _, appErr := service.DeleteContext(DeleteContextRequest{ContextID: "personal", ConfirmDelete: true}); appErr != nil {
		t.Fatalf("delete context with unknown workspace: %v", appErr)
	}
}

func TestRunningSummaryAndPreflightConflictsUseActiveEnvironmentIdentity(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeContext(t, fixture.context("freelance", "Freelance"))
	repository := coreRunning.NewRepository(fixture.runningPath)
	for _, environment := range []coreRunning.Environment{
		testRunningEnvironment(fixture, "personal", 0),
		testRunningEnvironment(fixture, "company", 0),
	} {
		if _, err := repository.Record(environment); err != nil {
			t.Fatalf("record environment: %v", err)
		}
	}
	service := fixture.service()
	service.dependencies.ProcessInspector = applicationProcessInspector{states: map[int]bool{0: true}}

	dashboard, appErr := service.GetHomeDashboard(GetHomeDashboardRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get home dashboard: %v", appErr)
	}
	if dashboard.Running.Count != 2 || !dashboard.Running.IsolationProtected || len(dashboard.Running.ContextCounts) != 2 {
		t.Fatalf("running summary = %#v", dashboard.Running)
	}

	same, appErr := service.PreflightLaunchProject(PreflightLaunchProjectRequest{ProjectPath: ".", ContextID: "personal"})
	if appErr != nil || same.RunningEnvironmentConflict == nil || same.RunningEnvironmentConflict.Kind != "same_context" {
		t.Fatalf("same-context conflict = %#v, %v", same.RunningEnvironmentConflict, appErr)
	}
	different, appErr := service.PreflightLaunchProject(PreflightLaunchProjectRequest{ProjectPath: ".", ContextID: "freelance"})
	if appErr != nil || different.RunningEnvironmentConflict == nil || different.RunningEnvironmentConflict.Kind != "different_context" {
		t.Fatalf("different-context conflict = %#v, %v", different.RunningEnvironmentConflict, appErr)
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
