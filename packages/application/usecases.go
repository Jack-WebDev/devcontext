package application

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	codingtool "devctx/packages/core/codingtool"
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

	ctx, err := devcontext.DefaultContextForIDWithRegistries(contextID, s.now(), s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry)
	if err != nil {
		return CreateContextResult{}, err
	}

	contextPaths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, contextID)
	if err != nil {
		return CreateContextResult{}, err
	}
	if err := filesystem.CreateContextDirectoryTreeWithRegistriesCredentialsAndPermissions(
		s.dependencies.Paths,
		contextPaths,
		ctx,
		s.dependencies.ProviderRegistry,
		s.dependencies.ToolRegistry,
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
	confidence := s.launchConfidenceStateForContext(ctx, providerEntries)
	return ContextState{
		ID:             ctx.ID.String(),
		Name:           ctx.Name,
		Description:    ctx.Metadata["description"],
		Tool:           toolState(ctx.Tool.DefaultTool, confidence),
		AvailableTools: toolOptions(s.dependencies.ToolRegistry),
		Providers:      providerStatesFromEntries(providerEntries),
		Confidence:     confidence,
		Metadata:       cloneMetadata(ctx.Metadata),
	}
}

func (s *Service) providerCredentialSessionStates() ([]ProviderCredentialSessionState, error) {
	homeDir, err := s.dependencies.Paths.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve provider credential source directory: %w", err)
	}

	states := make([]ProviderCredentialSessionState, 0)
	for _, integration := range s.dependencies.ProviderRegistry.All() {
		detector, ok := integration.(provider.GlobalCredentialDetector)
		if !ok {
			continue
		}
		session, found, err := detector.DetectGlobalCredentialSession(provider.GlobalCredentialContext{UserHomeDir: homeDir})
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		state := ProviderCredentialSessionState{
			ProviderID:        string(integration.ID()),
			Name:              integration.DisplayName(),
			MetadataAvailable: session.MetadataAvailable,
			Fields:            providerMetadataFields(session.Fields),
		}
		states = append(states, state)
	}
	return states, nil
}

type providerStateEntry struct {
	providerID provider.ID
	provider   provider.Provider
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
						RootDir:    contextPaths.RootDir,
						StorageDir: contextPaths.ProviderStorageDir(integration.ID()),
					},
				})
				if err != nil {
					status = provider.UnavailableStatus("Provider status could not be determined")
				}
			}
		}

		runtime := providerRuntimeContext(ctx, config, contextPaths, integration.ID())
		entries = append(entries, providerStateEntry{
			providerID: integration.ID(),
			provider:   integration,
			status:     status,
			state: ProviderState{
				ID:          string(integration.ID()),
				Name:        integration.DisplayName(),
				Enabled:     enabled,
				State:       providerReadinessState(status),
				Explanation: status.Explanation,
				ActionHint:  providerActionHint(integration, runtime, status),
				Identity:    providerIdentityState(integration, enabled, status, runtime, pathsErr),
			},
		})
	}
	return entries
}

func providerActionHint(integration provider.Provider, runtime provider.RuntimeContext, status provider.Status) string {
	if status.State != provider.StatusNotConfigured {
		return ""
	}
	if guidanceProvider, ok := integration.(provider.SetupGuidanceProvider); ok {
		if guidance := guidanceProvider.SetupGuidance(runtime); guidance.ActionHint != "" {
			return guidance.ActionHint
		}
	}
	return "Open " + integration.DisplayName() + " and complete its setup for this context."
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

func providerIdentityState(integration provider.Provider, enabled bool, status provider.Status, runtime provider.RuntimeContext, pathsErr error) ProviderIdentityState {
	if !enabled {
		return ProviderIdentityState{Status: ProviderIdentityNone}
	}

	switch status.State {
	case provider.StatusConfigured:
		if pathsErr != nil {
			return unavailableProviderIdentity()
		}
		return verifiedProviderIdentity(integration, runtime)
	case provider.StatusUnavailable:
		return unavailableProviderIdentity()
	default:
		return ProviderIdentityState{Status: ProviderIdentityNone}
	}
}

func verifiedProviderIdentity(integration provider.Provider, runtime provider.RuntimeContext) ProviderIdentityState {
	detector, ok := integration.(provider.ContextIdentityDetector)
	if !ok {
		return unavailableProviderIdentity()
	}
	identity, available, err := detector.DetectContextIdentity(runtime)
	if err != nil || !available {
		return unavailableProviderIdentity()
	}

	return ProviderIdentityState{
		Status: ProviderIdentityVerified,
		Fields: providerMetadataFields(identity.Fields),
	}
}

func providerRuntimeContext(ctx devcontext.Context, config provider.Config, paths filesystem.ContextPaths, providerID provider.ID) provider.RuntimeContext {
	return provider.RuntimeContext{
		ContextID: ctx.ID.String(),
		Config:    config,
		Paths: provider.ContextPaths{
			RootDir:    paths.RootDir,
			StorageDir: paths.ProviderStorageDir(providerID),
		},
	}
}

func providerMetadataFields(fields []provider.MetadataField) []ProviderMetadataField {
	result := make([]ProviderMetadataField, 0, len(fields))
	for _, field := range fields {
		if field.Label != "" && field.Value != "" {
			result = append(result, ProviderMetadataField{Label: field.Label, Value: field.Value})
		}
	}
	return result
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

	toolID := ctx.Tool.DefaultTool
	toolConfig := ctx.Tool.ConfigFor(toolID)
	registeredTool, registered := s.dependencies.ToolRegistry.Lookup(toolID)
	toolName := string(toolID)
	var executable codingtool.Executable
	var toolErr error
	if !registered {
		toolErr = fmt.Errorf("selected coding tool %q is not registered", toolID)
	} else {
		toolName = registeredTool.DisplayName
		executable, toolErr = registeredTool.Integration.DetectExecutable(toolConfig)
	}
	checks = append(checks, launcher.ToolConfidenceCheck(toolID, toolName, executable, toolErr))

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
		checks = append(checks, launcher.IsolationConfidenceChecks(contextPaths, enabledProviderIntegrations(providerEntries), toolID, toolName)...)
	}

	return LaunchConfidenceState{
		ContextID: ctx.ID.String(),
		Status:    launchConfidenceStatus(checks),
		Checks:    checks,
	}
}

func toolState(toolID codingtool.ID, confidence LaunchConfidenceState) ToolState {
	for _, check := range confidence.Checks {
		if check.Component == LaunchConfidenceCheckTool && check.ToolID == string(toolID) {
			return ToolState{
				ID:         check.ToolID,
				Name:       check.Label,
				Status:     check.Severity,
				Message:    check.Message,
				ActionHint: check.ActionHint,
			}
		}
	}
	return ToolState{
		ID:      string(toolID),
		Name:    string(toolID),
		Status:  LaunchConfidenceBlocked,
		Message: "Coding tool readiness could not be determined.",
	}
}

func toolOptions(registry codingtool.Registry) []ToolOption {
	tools := registry.All()
	options := make([]ToolOption, len(tools))
	for i, tool := range tools {
		options[i] = ToolOption{ID: string(tool.Integration.ID()), Name: tool.DisplayName}
	}
	return options
}

func enabledProviderIntegrations(entries []providerStateEntry) []provider.Provider {
	providers := make([]provider.Provider, 0, len(entries))
	for _, entry := range entries {
		if entry.state.Enabled {
			providers = append(providers, entry.provider)
		}
	}
	return providers
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
		Tool:             plan.Tool,
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
		ToolID:           string(plan.Tool.ID),
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
