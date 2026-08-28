package application

import (
	"sort"
	"time"

	devcontext "devctx/packages/core/context"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	coreRunning "devctx/packages/core/running"
)

func (s *Service) getRunningEnvironments() (RunningEnvironmentsState, error) {
	environments, err := s.refreshRunningEnvironments()
	if err != nil {
		return RunningEnvironmentsState{}, err
	}
	states := make([]RunningEnvironmentState, len(environments))
	for index, environment := range environments {
		states[index] = runningEnvironmentState(environment)
	}
	return RunningEnvironmentsState{Environments: states}, nil
}

func (s *Service) refreshRunningEnvironments() ([]coreRunning.Environment, error) {
	result, err := s.dependencies.RunningEnvironments.RefreshProcessStates(s.dependencies.ProcessInspector)
	if err != nil {
		return nil, err
	}
	for _, environment := range result.Stopped {
		s.recordHistoryEvent(environmentStoppedEvent(environment, s.now()))
	}
	active := make([]coreRunning.Environment, 0, len(result.Environments))
	for _, environment := range result.Environments {
		if environment.Process.State == coreRunning.ProcessStateRunning {
			active = append(active, environment)
		}
	}
	return active, nil
}

func (s *Service) homeRunningSummary() (HomeRunningSummary, error) {
	environments, err := s.refreshRunningEnvironments()
	if err != nil {
		return HomeRunningSummary{}, err
	}
	counts := map[string]HomeRunningContextCount{}
	for _, environment := range environments {
		count := counts[environment.Context.ID.String()]
		count.ContextID = environment.Context.ID.String()
		count.ContextName = environment.Context.Name
		count.Count++
		counts[count.ContextID] = count
	}
	contextCounts := make([]HomeRunningContextCount, 0, len(counts))
	for _, count := range counts {
		contextCounts = append(contextCounts, count)
	}
	sort.Slice(contextCounts, func(i, j int) bool { return contextCounts[i].ContextName < contextCounts[j].ContextName })
	return HomeRunningSummary{Count: len(environments), ContextCounts: contextCounts, IsolationProtected: len(environments) > 0}, nil
}

func (s *Service) runningEnvironmentConflict(projectPath project.Path, contextID devcontext.ID) (*RunningEnvironmentConflict, error) {
	environments, err := s.refreshRunningEnvironments()
	if err != nil {
		return nil, err
	}
	var differentContext *RunningEnvironmentConflict
	for _, environment := range environments {
		if environment.Project.Path != projectPath {
			continue
		}
		if environment.Context.ID == contextID {
			return &RunningEnvironmentConflict{Kind: "same_context", Environment: runningEnvironmentState(environment)}, nil
		}
		if differentContext == nil {
			differentContext = &RunningEnvironmentConflict{Kind: "different_context", Environment: runningEnvironmentState(environment)}
		}
	}
	return differentContext, nil
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
