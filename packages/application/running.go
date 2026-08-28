package application

import (
	coreRunning "devctx/packages/core/running"
)

func runningEnvironmentState(environment coreRunning.Environment) RunningEnvironmentState {
	return RunningEnvironmentState{
		ID:        string(environment.ID),
		Project:   ProjectState{Name: environment.Project.Name, Path: string(environment.Project.Path)},
		Context:   RunningEnvironmentContextState{ID: environment.Context.ID.String(), Name: environment.Context.Name},
		Tool:      ToolOption{ID: string(environment.Tool.ID), Name: environment.Tool.Name},
		StartedAt: environment.StartedAt.UTC(),
		Process:   RunningEnvironmentProcessState{State: string(environment.Process.State), PID: copyProcessID(environment.Process.PID)},
		Session:   RunningEnvironmentSessionState{ID: environment.Session.ID, State: string(environment.Session.State)},
		Launch:    RunningEnvironmentLaunchState{Source: string(environment.Launch.Source), ResolutionSource: string(environment.Launch.ResolutionSource)},
	}
}

func copyProcessID(pid *int) *int {
	if pid == nil {
		return nil
	}
	copy := *pid
	return &copy
}
