package running_test

import (
	"path/filepath"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/running"
)

func TestRepositoryRecordUpdatesMatchingProjectAndContext(t *testing.T) {
	repository := running.NewRepository(filepath.Join(t.TempDir(), "running.toml"))
	first, err := repository.Record(runningEnvironment("/work/api", "company", time.Date(2026, 8, 28, 10, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("record first environment: %v", err)
	}
	if first.ID == "" {
		t.Fatal("recorded environment ID is empty")
	}

	updated, err := repository.Record(runningEnvironment("/work/api", "company", time.Date(2026, 8, 28, 10, 5, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("update environment: %v", err)
	}
	if updated.ID != first.ID || !updated.StartedAt.After(first.StartedAt) {
		t.Fatalf("updated environment = %#v, want retained ID and new started time", updated)
	}

	otherContext, err := repository.Record(runningEnvironment("/work/api", "personal", time.Date(2026, 8, 28, 10, 10, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("record different-context environment: %v", err)
	}
	if otherContext.ID == first.ID {
		t.Fatal("different context reused environment ID")
	}

	environments, err := repository.List()
	if err != nil {
		t.Fatalf("list environments: %v", err)
	}
	if len(environments) != 2 {
		t.Fatalf("environments = %#v, want two", environments)
	}
}

func TestRepositoryMarkStoppedChangesOnlySelectedWorkspace(t *testing.T) {
	repository := running.NewRepository(filepath.Join(t.TempDir(), "running.toml"))
	first, err := repository.Record(runningEnvironment("/work/api", "company", time.Now()))
	if err != nil {
		t.Fatalf("record first: %v", err)
	}
	second, err := repository.Record(runningEnvironment("/work/web", "personal", time.Now()))
	if err != nil {
		t.Fatalf("record second: %v", err)
	}
	stopped, err := repository.MarkStopped(first.ID)
	if err != nil {
		t.Fatalf("mark stopped: %v", err)
	}
	if stopped.ID != first.ID || stopped.Process.State != running.ProcessStateStopped || stopped.Session.State != running.SessionStateEnded {
		t.Fatalf("stopped workspace = %#v", stopped)
	}
	all, err := repository.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, environment := range all {
		if environment.ID == second.ID && environment.Process.State != running.ProcessStateRunning {
			t.Fatalf("other workspace changed: %#v", environment)
		}
	}
}

func runningEnvironment(projectPath, contextID string, startedAt time.Time) running.Environment {
	return running.Environment{
		Project:   running.ProjectIdentity{Path: project.Path(projectPath), Name: "api"},
		Context:   running.ContextIdentity{ID: devcontext.MustID(contextID), Name: contextID},
		Tool:      running.ToolIdentity{ID: codingtool.ID("second-tool"), Name: "Second Tool"},
		StartedAt: startedAt,
		Process:   running.Process{State: running.ProcessStateRunning},
		Session:   running.Session{State: running.SessionStateUnknown},
		Launch:    running.LaunchIdentity{Source: launcher.InvocationSourceGUI, ResolutionSource: launcher.ResolutionSourceExplicit},
	}
}
