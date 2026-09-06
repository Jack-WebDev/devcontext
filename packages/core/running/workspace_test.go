package running_test

import (
	"testing"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/running"
)

func TestEnvironmentLifecycleSeparatesObservationFromAdapterCapabilities(t *testing.T) {
	capabilities := codingtool.WorkspaceCapabilities{Focusable: true, Revealable: true, Stoppable: true}
	tests := []struct {
		name    string
		process running.ProcessState
		session running.SessionState
		want    running.WorkspaceState
	}{
		{name: "active process", process: running.ProcessStateRunning, want: running.WorkspaceStateActive},
		{name: "active session", session: running.SessionStateActive, want: running.WorkspaceStateActive},
		{name: "stopped process", process: running.ProcessStateStopped, session: running.SessionStateActive, want: running.WorkspaceStateStopped},
		{name: "ended session", process: running.ProcessStateUnknown, session: running.SessionStateEnded, want: running.WorkspaceStateStopped},
		{name: "unobserved", process: running.ProcessStateUnknown, session: running.SessionStateUnknown, want: running.WorkspaceStateUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lifecycle := (running.Environment{Process: running.Process{State: tt.process}, Session: running.Session{State: tt.session}}).Lifecycle(capabilities)
			if lifecycle.State != tt.want || !lifecycle.Focusable || !lifecycle.Revealable || !lifecycle.Stoppable {
				t.Fatalf("lifecycle = %#v", lifecycle)
			}
		})
	}
}
