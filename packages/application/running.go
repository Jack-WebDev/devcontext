package application

import (
	"time"

	devlog "devctx/packages/core/logging"
	coreRunning "devctx/packages/core/running"
)

func (s *Service) getRunningEnvironments() (RunningEnvironmentsState, error) {
	result, err := s.dependencies.RunningEnvironments.RefreshProcessStates(s.dependencies.ProcessInspector)
	if err != nil {
		return RunningEnvironmentsState{}, err
	}
	for _, environment := range result.Stopped {
		s.recordHistoryEvent(environmentStoppedEvent(environment, s.now()))
	}

	states := make([]RunningEnvironmentState, 0, len(result.Environments))
	for _, environment := range result.Environments {
		if environment.Process.State == coreRunning.ProcessStateRunning {
			states = append(states, runningEnvironmentState(environment))
		}
	}
	return RunningEnvironmentsState{Environments: states}, nil
}

func environmentStoppedEvent(environment coreRunning.Environment, timestamp time.Time) devlog.Event {
	return devlog.NewEvent(devlog.EventInput{
		Name:        devlog.EventEnvironmentStopped,
		Timestamp:   timestamp,
		ProjectPath: string(environment.Project.Path),
		ContextID:   environment.Context.ID.String(),
		ToolID:      string(environment.Tool.ID),
	})
}

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
