package application

import (
	"fmt"
	"sort"
	"time"

	codingtool "devctx/packages/core/codingtool"
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
		states[index] = s.runningEnvironmentState(environment)
	}
	return RunningEnvironmentsState{Environments: states}, nil
}

func (s *Service) revealWorkspace(request WorkspaceActionRequest) (WorkspaceRevealResult, error) {
	environment, err := s.activeWorkspace(request.WorkspaceID)
	if err != nil {
		return WorkspaceRevealResult{}, err
	}
	if !s.dependencies.ToolRegistry.WorkspaceCapabilities(environment.Tool.ID).Revealable {
		return WorkspaceRevealResult{}, fmt.Errorf("workspace reveal is not available")
	}
	targets, err := s.dependencies.ToolRegistry.RevealTargets(environment.Tool.ID, workspaceReference(environment))
	if err != nil {
		return WorkspaceRevealResult{}, err
	}
	result := WorkspaceRevealResult{Targets: make([]WorkspaceRevealTarget, len(targets))}
	for i, target := range targets {
		result.Targets[i] = WorkspaceRevealTarget{ID: target.ID, Label: target.Label}
	}
	if len(targets) != 1 && request.TargetID == "" {
		return result, nil
	}
	targetID := request.TargetID
	if targetID == "" {
		targetID = targets[0].ID
	}
	for _, target := range targets {
		if target.ID == targetID {
			return result, s.dependencies.ToolRegistry.RevealWorkspace(environment.Tool.ID, workspaceReference(environment), targetID)
		}
	}
	return WorkspaceRevealResult{}, fmt.Errorf("workspace reveal target does not exist")
}

func (s *Service) stopWorkspace(request WorkspaceActionRequest) error {
	environment, err := s.activeWorkspace(request.WorkspaceID)
	if err != nil {
		return err
	}
	if !s.dependencies.ToolRegistry.WorkspaceCapabilities(environment.Tool.ID).Stoppable {
		return fmt.Errorf("workspace stop is not available")
	}
	if err := s.dependencies.ToolRegistry.StopWorkspace(environment.Tool.ID, workspaceReference(environment)); err != nil {
		return err
	}
	stopped, err := s.dependencies.RunningEnvironments.MarkStopped(environment.ID)
	if err != nil {
		return err
	}
	s.recordHistoryEvent(environmentStoppedEvent(stopped, s.now()))
	return nil
}

func (s *Service) activeWorkspace(id string) (coreRunning.Environment, error) {
	if id == "" {
		return coreRunning.Environment{}, fmt.Errorf("workspace ID is required")
	}
	environments, err := s.refreshRunningEnvironments()
	if err != nil {
		return coreRunning.Environment{}, err
	}
	for _, environment := range environments {
		if string(environment.ID) == id {
			return environment, nil
		}
	}
	return coreRunning.Environment{}, fmt.Errorf("active workspace %q does not exist", id)
}

func workspaceReference(environment coreRunning.Environment) codingtool.WorkspaceReference {
	return codingtool.WorkspaceReference{ID: string(environment.ID), ProjectPath: string(environment.Project.Path), SessionID: environment.Session.ID, ProcessID: copyProcessID(environment.Process.PID)}
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
		if environment.Lifecycle(codingtool.WorkspaceCapabilities{}).State == coreRunning.WorkspaceStateActive {
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
			return &RunningEnvironmentConflict{Kind: "same_context", Environment: s.runningEnvironmentState(environment)}, nil
		}
		if differentContext == nil {
			differentContext = &RunningEnvironmentConflict{Kind: "different_context", Environment: s.runningEnvironmentState(environment)}
		}
	}
	return differentContext, nil
}

func environmentStoppedEvent(environment coreRunning.Environment, timestamp time.Time) devlog.Event {
	return devlog.NewEvent(devlog.EventInput{
		Name:        devlog.EventWorkspaceStopped,
		Timestamp:   timestamp,
		ProjectPath: string(environment.Project.Path),
		ContextID:   environment.Context.ID.String(),
		ToolID:      string(environment.Tool.ID),
	})
}

func (s *Service) runningEnvironmentState(environment coreRunning.Environment) RunningEnvironmentState {
	lifecycle := environment.Lifecycle(s.dependencies.ToolRegistry.WorkspaceCapabilities(environment.Tool.ID))
	return RunningEnvironmentState{
		ID:        string(environment.ID),
		Project:   ProjectState{Name: environment.Project.Name, Path: string(environment.Project.Path)},
		Context:   RunningEnvironmentContextState{ID: environment.Context.ID.String(), Name: environment.Context.Name},
		Tool:      ToolOption{ID: string(environment.Tool.ID), Name: environment.Tool.Name},
		StartedAt: environment.StartedAt.UTC(),
		Process:   RunningEnvironmentProcessState{State: string(environment.Process.State), PID: copyProcessID(environment.Process.PID)},
		Session:   RunningEnvironmentSessionState{ID: environment.Session.ID, State: string(environment.Session.State)},
		Launch:    RunningEnvironmentLaunchState{Source: string(environment.Launch.Source), ResolutionSource: string(environment.Launch.ResolutionSource)},
		Lifecycle: WorkspaceLifecycleState{
			State:      string(lifecycle.State),
			Focusable:  lifecycle.Focusable,
			Revealable: lifecycle.Revealable,
			Stoppable:  lifecycle.Stoppable,
		},
	}
}

func copyProcessID(pid *int) *int {
	if pid == nil {
		return nil
	}
	copy := *pid
	return &copy
}
