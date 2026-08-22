package application

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

// GetLaunchState returns the GUI selector state for one project.
func (s *Service) GetLaunchState(request GetLaunchStateRequest) (LaunchState, *Error) {
	state, err := s.getLaunchState(request)
	if err != nil {
		return LaunchState{}, NewError(err)
	}
	return state, nil
}

// LaunchProject builds a launch plan for a selected context and starts the
// editor process.
func (s *Service) LaunchProject(request LaunchProjectRequest) (LaunchProjectResult, *Error) {
	result, err := s.launchProject(request)
	if err != nil {
		return LaunchProjectResult{}, NewError(err)
	}
	return result, nil
}

// PreflightLaunchProject checks launch readiness for a selected context without
// starting the editor process.
func (s *Service) PreflightLaunchProject(request PreflightLaunchProjectRequest) (PreflightLaunchProjectResult, *Error) {
	result, err := s.preflightLaunchProject(request)
	if err != nil {
		return PreflightLaunchProjectResult{}, NewError(err)
	}
	return result, nil
}

// BindProject remembers the selected context for one project.
func (s *Service) BindProject(request BindProjectRequest) (ProjectBindingState, *Error) {
	state, err := s.bindProject(request)
	if err != nil {
		return ProjectBindingState{}, NewError(err)
	}
	return state, nil
}

// UnbindProject removes any remembered context for one project.
func (s *Service) UnbindProject(request UnbindProjectRequest) (ProjectBindingState, *Error) {
	state, err := s.unbindProject(request)
	if err != nil {
		return ProjectBindingState{}, NewError(err)
	}
	return state, nil
}

// CreateContext creates one default context for first-run onboarding.
func (s *Service) CreateContext(request CreateContextRequest) (CreateContextResult, *Error) {
	result, err := s.createContext(request)
	if err != nil {
		return CreateContextResult{}, NewError(err)
	}
	return result, nil
}

func (s *Service) getLaunchState(request GetLaunchStateRequest) (LaunchState, error) {
	projectPath, err := s.validatedProjectPath(request.ProjectPath)
	if err != nil {
		return LaunchState{}, err
	}

	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return LaunchState{}, err
	}
	if len(contexts) == 0 {
		return s.firstRunLaunchState(projectPath), nil
	}

	lookup, err := s.dependencies.Projects.LookupWithContextValidation(string(projectPath), projectPath, s.dependencies.Contexts)
	if err != nil {
		return LaunchState{}, err
	}

	resolution, err := launcher.NewResolver(s.dependencies.Contexts, s.dependencies.Projects).Resolve(launcher.LaunchRequest{
		ProjectPath: projectPath,
		Interactive: true,
		Source:      launcher.InvocationSourceGUI,
	})
	if err != nil {
		return LaunchState{}, err
	}

	providerCredentialSessions, err := s.providerCredentialSessionStates()
	if err != nil {
		providerCredentialSessions = nil
	}

	return LaunchState{
		Project:                    projectState(projectPath),
		Contexts:                   s.contextStates(contexts),
		Binding:                    bindingState(lookup),
		Confidence:                 s.launchConfidenceState(resolution.Context),
		SelectedContextID:          selectedContextID(resolution),
		SelectionRequired:          resolution.SelectionRequired,
		ResolutionSource:           string(resolution.Source),
		Warnings:                   warningStates(resolution.Warnings),
		ProviderCredentialSessions: providerCredentialSessions,
	}, nil
}

func (s *Service) firstRunLaunchState(projectPath project.Path) LaunchState {
	sessions, err := s.providerCredentialSessionStates()
	if err != nil {
		sessions = nil
	}

	return LaunchState{
		Project:                    projectState(projectPath),
		Contexts:                   []ContextState{},
		Binding:                    ProjectBindingState{ProjectPath: string(projectPath)},
		SelectionRequired:          true,
		ResolutionSource:           string(launcher.ResolutionSourceUserSelection),
		FirstRun:                   true,
		ProviderCredentialSessions: sessions,
	}
}

func (s *Service) createContext(request CreateContextRequest) (CreateContextResult, error) {
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return CreateContextResult{}, err
	}

	ctx, err := devcontext.DefaultContextForIDWithProviderRegistry(contextID, s.now(), s.dependencies.ProviderRegistry)
	if err != nil {
		return CreateContextResult{}, err
	}

	contextPaths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, contextID)
	if err != nil {
		return CreateContextResult{}, err
	}
	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentialsAndPermissions(
		s.dependencies.Paths,
		contextPaths,
		ctx,
		request.ImportProviderIDs,
		s.dependencies.StoragePermissions,
	); err != nil {
		return CreateContextResult{}, err
	}

	return CreateContextResult{Context: s.contextState(ctx)}, nil
}

func (s *Service) launchProject(request LaunchProjectRequest) (LaunchProjectResult, error) {
	projectPath, err := s.canonicalProjectPath(request.ProjectPath)
	if err != nil {
		return LaunchProjectResult{}, err
	}
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return LaunchProjectResult{}, err
	}

	confirmation := launcher.ContextMismatchUnconfirmed
	if request.ConfirmContextMismatch {
		confirmation = launcher.ContextMismatchAccepted
	}

	plan, err := s.launchPlanBuilder().Build(launcher.LaunchRequest{
		ProjectPath:          projectPath,
		RequestedContext:     &contextID,
		MismatchConfirmation: confirmation,
		Interactive:          true,
		Source:               launcher.InvocationSourceGUI,
	})
	if err != nil {
		s.recordLaunchEvent(devlog.NewEvent(devlog.EventInput{
			Name:             devlog.LaunchEventNameForError(err),
			Timestamp:        s.now(),
			ProjectPath:      string(projectPath),
			ContextID:        contextID.String(),
			Err:              err,
			KnownEnvironment: s.dependencies.ParentEnvironment,
		}))
		return LaunchProjectResult{}, err
	}

	s.recordLaunchEvent(eventFromLaunchPlan(devlog.EventContextResolution, plan, nil, s.now()))
	for range plan.MissingProviderIDs {
		event := eventFromLaunchPlan(devlog.EventLaunchProviderMissing, plan, nil, s.now())
		event.ErrorCategory = devlog.ErrorCategoryProvider
		s.recordLaunchEvent(event)
	}

	if err := s.processLauncher().Launch(processRequestFromLaunchPlan(plan, s.dependencies.DetachMode)); err != nil {
		s.recordLaunchEvent(eventFromLaunchPlan(devlog.EventLaunchProcessFailure, plan, err, s.now()))
		return LaunchProjectResult{}, err
	}

	s.recordLaunchEvent(eventFromLaunchPlan(devlog.EventLaunchSucceeded, plan, nil, s.now()))

	return LaunchProjectResult{
		Project:  projectState(plan.ProjectPath),
		Context:  s.contextState(plan.Context),
		Warnings: warningStates(plan.Warnings),
	}, nil
}

func (s *Service) preflightLaunchProject(request PreflightLaunchProjectRequest) (PreflightLaunchProjectResult, error) {
	projectPath, err := s.canonicalProjectPath(request.ProjectPath)
	if err != nil {
		return PreflightLaunchProjectResult{}, err
	}
	if err := project.ValidateProjectDirectory(projectPath); err != nil {
		return PreflightLaunchProjectResult{}, err
	}
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return PreflightLaunchProjectResult{}, err
	}

	confirmation := launcher.ContextMismatchUnconfirmed
	if request.ConfirmContextMismatch {
		confirmation = launcher.ContextMismatchAccepted
	}

	resolution, err := launcher.NewResolver(s.dependencies.Contexts, s.dependencies.Projects).Resolve(launcher.LaunchRequest{
		ProjectPath:          projectPath,
		RequestedContext:     &contextID,
		MismatchConfirmation: confirmation,
		Interactive:          true,
		Source:               launcher.InvocationSourceGUI,
	})
	if err != nil {
		return PreflightLaunchProjectResult{}, err
	}
	if resolution.Context == nil || resolution.SelectionRequired {
		return PreflightLaunchProjectResult{}, launcher.ErrLaunchSelectionRequired
	}

	contextState := s.contextState(*resolution.Context)
	return PreflightLaunchProjectResult{
		Project:    projectState(projectPath),
		Context:    contextState,
		Confidence: contextState.Confidence,
		Warnings:   warningStates(resolution.Warnings),
	}, nil
}

func (s *Service) bindProject(request BindProjectRequest) (ProjectBindingState, error) {
	projectPath, err := s.canonicalProjectPath(request.ProjectPath)
	if err != nil {
		return ProjectBindingState{}, err
	}
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return ProjectBindingState{}, err
	}

	binding, err := s.dependencies.Projects.Bind(string(projectPath), projectPath, contextID, s.dependencies.Contexts, s.now())
	if err != nil {
		return ProjectBindingState{}, err
	}

	return ProjectBindingState{
		ProjectPath: string(binding.ProjectPath),
		Bound:       true,
		ContextID:   binding.ContextID.String(),
	}, nil
}

func (s *Service) unbindProject(request UnbindProjectRequest) (ProjectBindingState, error) {
	projectPath, err := s.validatedProjectPath(request.ProjectPath)
	if err != nil {
		return ProjectBindingState{}, err
	}

	result, err := s.dependencies.Projects.Unbind(string(projectPath), projectPath)
	if err != nil {
		return ProjectBindingState{}, err
	}

	return ProjectBindingState{
		ProjectPath: string(result.ProjectPath),
		Bound:       false,
	}, nil
}

func (s *Service) canonicalProjectPath(input string) (project.Path, error) {
	if input == "" {
		input = "."
	}
	return project.CanonicalizePath(s.dependencies.Paths, input, project.Path(s.dependencies.WorkingDirectory))
}

func (s *Service) validatedProjectPath(input string) (project.Path, error) {
	projectPath, err := s.canonicalProjectPath(input)
	if err != nil {
		return "", err
	}
	if err := project.ValidateProjectDirectory(projectPath); err != nil {
		return "", err
	}
	return projectPath, nil
}

func (s *Service) contextStates(contexts []devcontext.Context) []ContextState {
	states := make([]ContextState, len(contexts))
	for i, ctx := range contexts {
		states[i] = s.contextState(ctx)
	}
	return states
}

func (s *Service) contextState(ctx devcontext.Context) ContextState {
	providerEntries := s.providerStateEntries(ctx)
	return ContextState{
		ID:         ctx.ID.String(),
		Name:       ctx.Name,
		Editor:     EditorState{Type: string(ctx.Editor.Type)},
		Providers:  providerStatesFromEntries(providerEntries),
		Confidence: s.launchConfidenceStateForContext(ctx, providerEntries),
		Metadata:   cloneMetadata(ctx.Metadata),
	}
}

func (s *Service) providerCredentialSessionStates() ([]ProviderCredentialSessionState, error) {
	sessions, err := filesystem.DetectProviderCredentialSessions(s.dependencies.Paths)
	if err != nil {
		return nil, err
	}

	states := make([]ProviderCredentialSessionState, 0, len(sessions))
	for _, session := range sessions {
		integration, ok := s.dependencies.ProviderRegistry.Get(provider.ID(session.ProviderID))
		if !ok {
			continue
		}
		state := ProviderCredentialSessionState{
			ProviderID:        session.ProviderID,
			Name:              integration.DisplayName(),
			MetadataAvailable: session.MetadataAvailable,
		}
		switch session.ProviderID {
		case string(provider.CodexID):
			state.Codex = &CodexCredentialSessionState{
				Email:            session.Codex.Email,
				ChatGPTPlanType:  session.Codex.ChatGPTPlanType,
				ChatGPTAccountID: session.Codex.ChatGPTAccountID,
			}
		case string(provider.ClaudeID):
			state.Claude = &ClaudeCredentialSessionState{
				SubscriptionType: session.Claude.SubscriptionType,
				OrganizationUUID: session.Claude.OrganizationUUID,
				OrganizationName: session.Claude.OrganizationName,
			}
		default:
			continue
		}
		states = append(states, state)
	}
	return states, nil
}

type providerStateEntry struct {
	providerID provider.ID
	state      ProviderState
	status     provider.Status
}

func (s *Service) providerStateEntries(ctx devcontext.Context) []providerStateEntry {
	providers := s.dependencies.ProviderRegistry.All()
	entries := make([]providerStateEntry, 0, len(providers))
	contextPaths, pathsErr := filesystem.DeriveContextPaths(s.dependencies.Paths, ctx.ID)
	if pathsErr == nil {
		contextPaths = contextPaths.WithProviderStorageDirs(enabledProviderIDs(ctx))
	}
	for _, integration := range providers {
		config, ok := ctx.Providers[integration.ID()]
		enabled := ok && config.Enabled
		status := provider.NotConfiguredStatus("Provider is disabled for this context")
		if enabled {
			if pathsErr != nil {
				status = provider.UnavailableStatus("Provider status could not be determined")
			} else {
				var err error
				status, err = integration.Status(provider.RuntimeContext{
					ContextID: ctx.ID.String(),
					Config:    config,
					Paths: provider.ContextPaths{
						RootDir:           contextPaths.RootDir,
						StorageDir:        contextPaths.ProviderStorageDir(integration.ID()),
						VSCodeDir:         contextPaths.VSCodeDir,
						VSCodeUserDataDir: contextPaths.VSCodeUserDataDir,
					},
				})
				if err != nil {
					status = provider.UnavailableStatus("Provider status could not be determined")
				}
			}
		}

		entries = append(entries, providerStateEntry{
			providerID: integration.ID(),
			status:     status,
			state: ProviderState{
				ID:          string(integration.ID()),
				Name:        integration.DisplayName(),
				Enabled:     enabled,
				State:       providerReadinessState(status),
				Explanation: status.Explanation,
				Identity:    providerIdentityState(integration.ID(), enabled, status, contextPaths, pathsErr),
			},
		})
	}
	return entries
}

func providerStatesFromEntries(entries []providerStateEntry) []ProviderState {
	states := make([]ProviderState, len(entries))
	for i, entry := range entries {
		states[i] = entry.state
	}
	return states
}

func enabledProviderIDs(ctx devcontext.Context) []provider.ID {
	ids := make([]provider.ID, 0, len(ctx.Providers))
	for providerID, config := range ctx.Providers {
		if config.Enabled {
			ids = append(ids, providerID)
		}
	}
	sort.Slice(ids, func(i int, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func providerReadinessState(status provider.Status) ProviderReadinessState {
	switch status.State {
	case provider.StatusConfigured:
		return ProviderReadinessReady
	case provider.StatusNotConfigured:
		return ProviderReadinessNotConfigured
	case provider.StatusDirectoryMissing:
		return ProviderReadinessDirectoryMissing
	case provider.StatusUnavailable:
		return ProviderReadinessUnavailable
	default:
		return ProviderReadinessUnavailable
	}
}

func providerIdentityState(providerID provider.ID, enabled bool, status provider.Status, contextPaths filesystem.ContextPaths, pathsErr error) ProviderIdentityState {
	if !enabled {
		return ProviderIdentityState{Status: ProviderIdentityNone}
	}

	switch status.State {
	case provider.StatusConfigured:
		if pathsErr != nil {
			return unavailableProviderIdentity()
		}
		return verifiedProviderIdentity(providerID, contextPaths)
	case provider.StatusUnavailable:
		return unavailableProviderIdentity()
	default:
		return ProviderIdentityState{Status: ProviderIdentityNone}
	}
}

func verifiedProviderIdentity(providerID provider.ID, contextPaths filesystem.ContextPaths) ProviderIdentityState {
	switch providerID {
	case provider.CodexID:
		metadata, available, err := filesystem.DetectCodexContextCredentialMetadata(contextPaths)
		if err != nil || !available {
			return unavailableProviderIdentity()
		}
		return ProviderIdentityState{
			Status: ProviderIdentityVerified,
			Codex: &CodexProviderIdentityState{
				Email:            metadata.Email,
				ChatGPTPlanType:  metadata.ChatGPTPlanType,
				ChatGPTAccountID: metadata.ChatGPTAccountID,
			},
		}
	case provider.ClaudeID:
		metadata, available, err := filesystem.DetectClaudeContextCredentialMetadata(contextPaths)
		if err != nil || !available {
			return unavailableProviderIdentity()
		}
		return ProviderIdentityState{
			Status: ProviderIdentityVerified,
			Claude: &ClaudeProviderIdentityState{
				SubscriptionType: metadata.SubscriptionType,
				OrganizationUUID: metadata.OrganizationUUID,
				OrganizationName: metadata.OrganizationName,
			},
		}
	default:
		return unavailableProviderIdentity()
	}
}

func unavailableProviderIdentity() ProviderIdentityState {
	return ProviderIdentityState{
		Status:  ProviderIdentityUnavailable,
		Message: "Account identity unavailable.",
	}
}

func (s *Service) launchConfidenceState(ctx *devcontext.Context) *LaunchConfidenceState {
	if ctx == nil {
		return nil
	}
	confidence := s.launchConfidenceStateForContext(*ctx, s.providerStateEntries(*ctx))
	return &confidence
}

func (s *Service) launchConfidenceStateForContext(ctx devcontext.Context, providerEntries []providerStateEntry) LaunchConfidenceState {
	checks := make([]LaunchConfidenceCheck, 0)
	for _, entry := range providerEntries {
		if !entry.state.Enabled {
			continue
		}
		check, ok := launcher.ProviderConfidenceCheck(entry.providerID, entry.state.Name, entry.status)
		if ok {
			checks = append(checks, check)
		}
	}

	executable, editorErr := s.dependencies.Editor.DetectExecutable(ctx.Editor)
	checks = append(checks, launcher.VSCodeConfidenceCheck(executable, editorErr))

	contextPaths, pathsErr := filesystem.DeriveContextPaths(s.dependencies.Paths, ctx.ID)
	if pathsErr != nil {
		checks = append(checks, LaunchConfidenceCheck{
			Component:  LaunchConfidenceCheckIsolation,
			Severity:   LaunchConfidenceBlocked,
			Label:      "Isolation",
			Message:    "Context isolation paths could not be determined.",
			ActionHint: "Run diagnostics to inspect context storage.",
		})
	} else {
		contextPaths = contextPaths.WithProviderStorageDirs(enabledProviderIDs(ctx))
		checks = append(checks, launcher.IsolationConfidenceChecks(contextPaths)...)
	}

	return LaunchConfidenceState{
		ContextID: ctx.ID.String(),
		Status:    launchConfidenceStatus(checks),
		Checks:    checks,
	}
}

func launchConfidenceStatus(checks []LaunchConfidenceCheck) LaunchConfidenceStatus {
	status := LaunchConfidenceReady
	for _, check := range checks {
		if check.Severity == LaunchConfidenceBlocked {
			return LaunchConfidenceBlocked
		}
		if check.Severity == LaunchConfidenceNeedsAttention {
			status = LaunchConfidenceNeedsAttention
		}
	}
	return status
}

func processRequestFromLaunchPlan(plan launcher.LaunchPlan, detachMode launcher.DetachMode) launcher.ProcessRequest {
	return launcher.ProcessRequest{
		Executable:       plan.Executable,
		Arguments:        append(launcher.Arguments(nil), plan.Arguments...),
		Environment:      cloneLaunchEnvironment(plan.Environment),
		WorkingDirectory: plan.WorkingDirectory,
		DetachMode:       detachMode,
	}
}

func cloneLaunchEnvironment(environment launcher.Environment) launcher.Environment {
	cloned := make(launcher.Environment, len(environment))
	for key, value := range environment {
		cloned[key] = value
	}
	return cloned
}

func (s *Service) recordLaunchEvent(event devlog.Event) {
	_ = s.logger().Record(event)
}

func eventFromLaunchPlan(name devlog.EventName, plan launcher.LaunchPlan, err error, timestamp time.Time) devlog.Event {
	return devlog.NewEvent(devlog.EventInput{
		Name:             name,
		Timestamp:        timestamp,
		ProjectPath:      string(plan.ProjectPath),
		ContextID:        plan.Context.ID.String(),
		EditorID:         string(plan.Editor.Type),
		ResolutionSource: string(plan.ResolutionSource),
		Err:              err,
		KnownEnvironment: plan.Environment.Environ(),
	})
}

func projectState(projectPath project.Path) ProjectState {
	return ProjectState{
		Name: projectName(projectPath),
		Path: string(projectPath),
	}
}

func projectName(projectPath project.Path) string {
	name := filepath.Base(string(projectPath))
	if name == "." || name == string(os.PathSeparator) {
		return string(projectPath)
	}
	return name
}

func bindingState(lookup project.BindingLookup) ProjectBindingState {
	state := ProjectBindingState{
		ProjectPath: string(lookup.ProjectPath),
		Bound:       lookup.Bound,
		Dangling:    lookup.Dangling,
		Recovery:    lookup.Recovery,
	}
	if lookup.Bound {
		state.ContextID = lookup.Binding.ContextID.String()
	}
	if lookup.Dangling {
		state.MissingContextID = lookup.MissingContextID.String()
	}
	return state
}

func selectedContextID(resolution launcher.ResolutionResult) string {
	if resolution.Context == nil {
		return ""
	}
	return resolution.Context.ID.String()
}

func warningStates(warnings []launcher.ResolutionWarning) []ResolutionWarning {
	if len(warnings) == 0 {
		return nil
	}

	states := make([]ResolutionWarning, len(warnings))
	for i, warning := range warnings {
		states[i] = ResolutionWarning{
			Code:               string(warning.Code),
			Message:            warning.Message,
			ProjectPath:        string(warning.ProjectPath),
			BoundContextID:     warning.BoundContextID.String(),
			RequestedContextID: warning.RequestedContextID.String(),
		}
	}
	return states
}

func cloneMetadata(metadata devcontext.Metadata) map[string]string {
	if len(metadata) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}
