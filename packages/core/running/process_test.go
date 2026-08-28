package running_test

import (
	"path/filepath"
	"testing"
	"time"

	"devctx/packages/core/running"
)

func TestRefreshProcessStatesStopsOnlyKnownInactiveProcesses(t *testing.T) {
	repository := running.NewRepository(filepath.Join(t.TempDir(), "running.toml"))
	runningPID := 100
	stoppedPID := 200
	withRunningPID := runningEnvironment("/work/api", "company", time.Now())
	withRunningPID.Process = running.Process{PID: &runningPID, State: running.ProcessStateRunning}
	withRunningPID.Session = running.Session{ID: "active", State: running.SessionStateActive}
	withStoppedPID := runningEnvironment("/work/web", "company", time.Now())
	withStoppedPID.Process = running.Process{PID: &stoppedPID, State: running.ProcessStateRunning}
	withStoppedPID.Session = running.Session{ID: "stopped", State: running.SessionStateActive}
	withoutPID := runningEnvironment("/work/docs", "company", time.Now())

	for _, environment := range []running.Environment{withRunningPID, withStoppedPID, withoutPID} {
		if _, err := repository.Record(environment); err != nil {
			t.Fatalf("record environment: %v", err)
		}
	}

	result, err := repository.RefreshProcessStates(fakeProcessInspector{states: map[int]bool{runningPID: true, stoppedPID: false}})
	if err != nil {
		t.Fatalf("refresh process states: %v", err)
	}
	if len(result.Stopped) != 1 || result.Stopped[0].Project.Path != "/work/web" || result.Stopped[0].Session.State != running.SessionStateEnded {
		t.Fatalf("stopped environments = %#v", result.Stopped)
	}
	states := map[string]running.ProcessState{}
	for _, environment := range result.Environments {
		states[string(environment.Project.Path)] = environment.Process.State
	}
	if states["/work/api"] != running.ProcessStateRunning || states["/work/web"] != running.ProcessStateStopped || states["/work/docs"] != running.ProcessStateRunning {
		t.Fatalf("process states = %#v", states)
	}
}

type fakeProcessInspector struct{ states map[int]bool }

func (f fakeProcessInspector) IsRunning(pid int) (bool, error) { return f.states[pid], nil }
