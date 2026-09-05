package application

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
	coreRunning "devctx/packages/core/running"
)

// GetLaunchState returns the GUI selector state for one project.
func (s *Service) GetLaunchState(request GetLaunchStateRequest) (LaunchState, *Error) {
	state, err := s.getLaunchState(request)
	if err != nil {
		return LaunchState{}, NewError(err)
	}
	return state, nil
}

// ValidateProjectDirectory returns a presentation-safe project identity only
// when the selected folder is a readable directory.
func (s *Service) ValidateProjectDirectory(request ValidateProjectDirectoryRequest) (ProjectState, *Error) {
	projectPath, err := s.validatedProjectPath(request.ProjectPath)
	if err != nil {
		return ProjectState{}, NewError(err)
	}
	return projectState(projectPath), nil
}

func (s *Service) GetSettings() (SettingsState, *Error) {
	settings, err := s.getSettings()
	if err != nil {
		return SettingsState{}, NewError(err)
	}
	return settings, nil
}

func (s *Service) UpdateSettings(request UpdateSettingsRequest) (SettingsState, *Error) {
	settings, err := s.updateSettings(request)
	if err != nil {
		return SettingsState{}, NewError(err)
	}
	return settings, nil
}

func (s *Service) GetTrayState() (TrayState, *Error) {
	state, err := s.getTrayState()
	if err != nil {
		return TrayState{}, NewError(err)
	}
	return state, nil
}

func (s *Service) getTrayState() (TrayState, error) {
	settings, err := s.getSettings()
	if err != nil {
		return TrayState{}, err
	}
	running, err := s.getRunningEnvironments()
	if err != nil {
		return TrayState{}, err
	}
	recents, err := s.getRecentProjects()
	if err != nil {
		return TrayState{}, err
	}
	state := TrayState{Enabled: settings.TrayEnabled, Indicator: "normal", Environments: make([]TrayEnvironmentItem, 0, len(running.Environments)), RecentProjects: make([]TrayRecentProjectItem, 0, len(recents))}
	for _, environment := range running.Environments {
		state.Environments = append(state.Environments, TrayEnvironmentItem{ID: environment.ID, ProjectName: environment.Project.Name, ContextName: environment.Context.Name, ToolName: environment.Tool.Name})
	}
	for _, recent := range recents {
		state.RecentProjects = append(state.RecentProjects, TrayRecentProjectItem{ProjectName: recent.Project.Name, ContextName: recent.ContextName})
	}
	return state, nil
}

func (s *Service) getSettings() (SettingsState, error) {
	globalConfig, err := config.ReadGlobalConfigFile(s.dependencies.ConfigPath)
	if err != nil {
		return SettingsState{}, err
	}
	return settingsState(globalConfig), nil
}

func (s *Service) updateSettings(request UpdateSettingsRequest) (SettingsState, error) {
	globalConfig, err := config.ReadGlobalConfigFile(s.dependencies.ConfigPath)
	if err != nil {
		return SettingsState{}, err
	}
	globalConfig.UI.CloseAfterLaunch = request.CloseAfterLaunch
	globalConfig.UI.LaunchVerification = request.LaunchVerification
	globalConfig.UI.RememberProjects = request.RememberProjects
	globalConfig.UI.TrayEnabled = request.TrayEnabled
	if err := config.WriteGlobalConfigFileWithPermissions(s.dependencies.ConfigPath, globalConfig, s.dependencies.StoragePermissions); err != nil {
		return SettingsState{}, err
	}
	return settingsState(globalConfig), nil
}

func settingsState(globalConfig config.GlobalConfig) SettingsState {
	return SettingsState{CloseAfterLaunch: globalConfig.UI.CloseAfterLaunch, LaunchVerification: globalConfig.UI.LaunchVerification, RememberProjects: globalConfig.UI.RememberProjects, TrayEnabled: globalConfig.UI.TrayEnabled}
}

// GetHomeDashboard returns the backend-owned summary for the Home screen.
func (s *Service) GetHomeDashboard(request GetHomeDashboardRequest) (HomeDashboardState, *Error) {
	dashboard, err := s.getHomeDashboard(request)
	if err != nil {
		return HomeDashboardState{}, NewError(err)
	}
	return dashboard, nil
}

// GetRecentProjects returns successful project launches without reading log
// files. It is independent of project bindings and the currently selected
// project.
func (s *Service) GetRecentProjects() (RecentProjectsState, *Error) {
	projects, err := s.getRecentProjects()
	if err != nil {
		return RecentProjectsState{}, NewError(err)
	}
	return RecentProjectsState{Projects: projects}, nil
}

// GetContexts returns all configured development identities with their
// backend-derived readiness, project usage, and recent launch summaries.
func (s *Service) GetContexts() (ContextListState, *Error) {
	contexts, err := s.getContexts()
	if err != nil {
		return ContextListState{}, NewError(err)
	}
	return ContextListState{Contexts: contexts}, nil
}

// GetContextDetails returns one context's presentation-safe details.
func (s *Service) GetContextDetails(request GetContextDetailsRequest) (ContextDetailsState, *Error) {
	details, err := s.getContextDetails(request)
	if err != nil {
		return ContextDetailsState{}, NewError(err)
	}
	return details, nil
}

// GetTrustCenter returns factual local protection, project mapping, and
// coding-tool integration-boundary data for the Trust Center.
func (s *Service) GetTrustCenter() (TrustCenterState, *Error) {
	state, err := s.getTrustCenter()
	if err != nil {
		return TrustCenterState{}, NewError(err)
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

// GetContextTemplates returns safe defaults for the create-context flow.
func (s *Service) GetContextTemplates() ContextTemplatesState {
	templates := devcontext.Templates()
	states := make([]ContextTemplateState, len(templates))
	for index, template := range templates {
		states[index] = ContextTemplateState{
			ID: template.ID, Name: template.Name, Description: template.Description,
			Icon: template.Icon, Accent: template.Accent,
		}
	}
	return ContextTemplatesState{Templates: states}
}

// DuplicateContext creates a new isolated context with the source context's
// metadata, provider enablement, and coding-tool settings. Provider
// credentials are intentionally excluded.
func (s *Service) DuplicateContext(request DuplicateContextRequest) (DuplicateContextResult, *Error) {
	result, err := s.duplicateContext(request)
	if err != nil {
		return DuplicateContextResult{}, NewError(err)
	}
	return result, nil
}

// ExportContextMetadata returns a versioned portable configuration for one
// context. Credentials and context-owned integration storage are never read.
func (s *Service) ExportContextMetadata(request ExportContextMetadataRequest) (ContextMetadataExport, *Error) {
	result, err := s.exportContextMetadata(request)
	if err != nil {
		return ContextMetadataExport{}, NewError(err)
	}
	return result, nil
}

// ImportContextMetadata creates a new isolated context from a versioned safe
// metadata export. It does not import credentials or provider storage.
func (s *Service) ImportContextMetadata(request ImportContextMetadataRequest) (ImportContextMetadataResult, *Error) {
	result, err := s.importContextMetadata(request)
	if err != nil {
		return ImportContextMetadataResult{}, NewError(err)
	}
	return result, nil
}

// GetProjects returns known project summaries without requiring log scraping.
func (s *Service) GetProjects() (ProjectsState, *Error) {
	projects, err := s.getProjects()
	if err != nil {
		return ProjectsState{}, NewError(err)
	}
	return ProjectsState{Projects: projects}, nil
}

// GetDiagnostics returns backend-owned diagnostics for one configured context.
func (s *Service) GetDiagnostics(request GetDiagnosticsRequest) (DiagnosticsState, *Error) {
	diagnostics, err := s.getDiagnostics(request)
	if err != nil {
		return DiagnosticsState{}, NewError(err)
	}
	return diagnostics, nil
}

// GetRepairActions returns backend-owned repair actions and previews for one
// configured context.
func (s *Service) GetRepairActions(request GetRepairActionsRequest) (RepairActionsState, *Error) {
	actions, err := s.getRepairActions(request)
	if err != nil {
		return RepairActionsState{}, NewError(err)
	}
	return actions, nil
}

// RunRepairAction executes one backend-advertised repair action.
func (s *Service) RunRepairAction(request RunRepairActionRequest) (RunRepairActionResult, *Error) {
	result, err := s.runRepairAction(request)
	if err != nil {
		return RunRepairActionResult{}, NewError(err)
	}
	return result, nil
}

// GetHistory returns local user-facing activity records.
func (s *Service) GetHistory() (HistoryState, *Error) {
	history, err := s.getHistory()
	if err != nil {
		return HistoryState{}, NewError(err)
	}
	return history, nil
}

// GetRunningEnvironments returns active coding-tool environments after
// refreshing process state where a PID is available.
func (s *Service) GetRunningEnvironments() (RunningEnvironmentsState, *Error) {
	state, err := s.getRunningEnvironments()
	if err != nil {
		return RunningEnvironmentsState{}, NewError(err)
	}
	return state, nil
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

func (s *Service) getHomeDashboard(request GetHomeDashboardRequest) (HomeDashboardState, error) {
	launchState, err := s.getLaunchState(GetLaunchStateRequest{ProjectPath: request.ProjectPath})
	if err != nil {
		return HomeDashboardState{}, err
	}

	dashboard := HomeDashboardState{
		Project:        launchState.Project,
		RecentProjects: []RecentProjectState{},
		Running:        HomeRunningSummary{},
		Activity:       HomeActivitySummary{},
	}
	running, err := s.homeRunningSummary()
	if err != nil {
		return HomeDashboardState{}, err
	}
	dashboard.Running = running
	for _, context := range launchState.Contexts {
		if context.ID != launchState.SelectedContextID {
			continue
		}
		dashboard.CurrentContext = &HomeCurrentContextState{
			ID:         context.ID,
			Name:       context.Name,
			Tool:       context.Tool,
			Confidence: context.Confidence,
		}
		break
	}
	return dashboard, nil
}

func (s *Service) getRecentProjects() ([]RecentProjectState, error) {
	recents, err := s.dependencies.RecentProjects.List()
	if err != nil {
		return nil, err
	}

	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return nil, err
	}
	contextNames := make(map[devcontext.ID]string, len(contexts))
	for _, configuredContext := range contexts {
		contextNames[configuredContext.ID] = configuredContext.Name
	}

	projects := make([]RecentProjectState, 0, len(recents))
	for _, recent := range recents {
		projects = append(projects, RecentProjectState{
			Project:        projectState(recent.ProjectPath),
			ContextID:      recent.ContextID.String(),
			ContextName:    contextNames[recent.ContextID],
			LastLaunchedAt: recent.LastLaunchedAt.UTC(),
		})
	}
	sort.SliceStable(projects, func(i, j int) bool {
		return projects[i].LastLaunchedAt.After(projects[j].LastLaunchedAt)
	})
	return projects, nil
}

func (s *Service) getContexts() ([]ContextListItem, error) {
	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return nil, err
	}
	usage, err := s.contextUsage()
	if err != nil {
		return nil, err
	}

	items := make([]ContextListItem, 0, len(contexts))
	for _, configuredContext := range contexts {
		state := s.contextState(configuredContext)
		item := ContextListItem{
			Context:          state,
			EnabledProviders: enabledProviderStates(state.Providers),
			ProjectCount:     usage.projectCounts[configuredContext.ID],
		}
		if lastUsed, found := usage.lastUsedAt[configuredContext.ID]; found {
			item.LastUsedAt = &lastUsed
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) getContextDetails(request GetContextDetailsRequest) (ContextDetailsState, error) {
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return ContextDetailsState{}, err
	}
	configuredContext, err := s.dependencies.Contexts.Get(contextID)
	if err != nil {
		return ContextDetailsState{}, err
	}
	paths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, contextID)
	if err != nil {
		return ContextDetailsState{}, err
	}
	usage, err := s.contextUsage()
	if err != nil {
		return ContextDetailsState{}, err
	}

	state := s.contextState(configuredContext)
	details := ContextDetailsState{
		Context:          state,
		Location:         paths.RootDir,
		CreatedAt:        configuredContext.CreatedAt.UTC(),
		ProjectCount:     usage.projectCounts[contextID],
		EnabledProviders: enabledProviderStates(state.Providers),
	}
	if lastUsed, found := usage.lastUsedAt[contextID]; found {
		details.LastUsedAt = &lastUsed
	}
	return details, nil
}

type contextUsage struct {
	projectCounts map[devcontext.ID]int
	lastUsedAt    map[devcontext.ID]time.Time
}

func (s *Service) contextUsage() (contextUsage, error) {
	bindings, err := s.dependencies.Projects.List()
	if err != nil {
		return contextUsage{}, err
	}
	recents, err := s.dependencies.RecentProjects.List()
	if err != nil {
		return contextUsage{}, err
	}

	usage := contextUsage{
		projectCounts: make(map[devcontext.ID]int),
		lastUsedAt:    make(map[devcontext.ID]time.Time),
	}
	for _, binding := range bindings {
		usage.projectCounts[binding.ContextID]++
	}
	for _, recent := range recents {
		if previous, found := usage.lastUsedAt[recent.ContextID]; !found || recent.LastLaunchedAt.After(previous) {
			usage.lastUsedAt[recent.ContextID] = recent.LastLaunchedAt.UTC()
		}
	}
	return usage, nil
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
	contextID, err := s.contextIDForCreateRequest(request)
	if err != nil {
		return CreateContextResult{}, err
	}

	ctx, err := s.contextFromCreateRequest(contextID, request)
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
	s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{
		Name:      devlog.EventContextCreated,
		Timestamp: s.now(),
		ContextID: ctx.ID.String(),
		ToolID:    string(ctx.Tool.DefaultTool),
	}))
	recordedProviders := make(map[provider.ID]bool, len(request.ImportProviderIDs))
	for _, rawProviderID := range request.ImportProviderIDs {
		providerID := provider.ID(rawProviderID)
		if recordedProviders[providerID] {
			continue
		}
		if _, enabled := ctx.Providers[providerID]; !enabled {
			continue
		}
		recordedProviders[providerID] = true
		s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{
			Name:      devlog.EventProviderConnected,
			Timestamp: s.now(),
			ContextID: ctx.ID.String(),
			ToolID:    string(ctx.Tool.DefaultTool),
		}))
	}

	return CreateContextResult{Context: s.contextState(ctx)}, nil
}

func (s *Service) contextIDForCreateRequest(request CreateContextRequest) (devcontext.ID, error) {
	if rawID := strings.TrimSpace(request.ContextID); rawID != "" {
		return devcontext.NewID(rawID)
	}

	name := strings.TrimSpace(request.Name)
	if name == "" && request.TemplateID != "" {
		if template, ok := devcontext.TemplateByID(request.TemplateID); ok {
			name = template.Name
		}
	}
	base := contextIDSlug(name)
	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return devcontext.ID{}, fmt.Errorf("list contexts while generating context ID: %w", err)
	}
	existing := make(map[string]struct{}, len(contexts))
	for _, ctx := range contexts {
		existing[ctx.ID.String()] = struct{}{}
	}
	for suffix := 1; ; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		if _, found := existing[candidate]; found {
			continue
		}
		return devcontext.NewID(candidate)
	}
}

// contextIDSlug converts a display name to the restricted identifier format
// used by persisted context directories. It intentionally keeps IDs internal.
func contextIDSlug(name string) string {
	var builder strings.Builder
	separatorPending := false
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			if separatorPending && builder.Len() > 0 {
				builder.WriteByte('-')
			}
			builder.WriteRune(r)
			separatorPending = false
		default:
			separatorPending = builder.Len() > 0
		}
	}
	if builder.Len() == 0 {
		return "context"
	}
	return builder.String()
}

func (s *Service) contextFromCreateRequest(contextID devcontext.ID, request CreateContextRequest) (devcontext.Context, error) {
	if request.TemplateID != "" {
		template, ok := devcontext.TemplateByID(request.TemplateID)
		if !ok {
			return devcontext.Context{}, fmt.Errorf("unknown context template %q", request.TemplateID)
		}
		if strings.TrimSpace(request.Name) == "" {
			request.Name = template.Name
		}
		if strings.TrimSpace(request.Description) == "" {
			request.Description = template.Description
		}
		if strings.TrimSpace(request.Icon) == "" {
			request.Icon = template.Icon
		}
		if strings.TrimSpace(request.Accent) == "" {
			request.Accent = template.Accent
		}
	}
	if strings.TrimSpace(request.Name) == "" {
		return devcontext.DefaultContextForIDWithRegistries(contextID, s.now(), s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry)
	}
	if err := s.applyDevelopmentToolSelections(&request); err != nil {
		return devcontext.Context{}, err
	}
	toolID := codingtool.ID(request.ToolID)
	if toolID == "" {
		toolID = s.dependencies.ToolRegistry.DefaultID()
	}
	if _, ok := s.dependencies.ToolRegistry.Get(toolID); !ok {
		return devcontext.Context{}, fmt.Errorf("unknown coding tool %q", toolID)
	}
	providers := make(provider.Configs, len(request.EnabledProviderIDs))
	for _, rawID := range request.EnabledProviderIDs {
		providerID := provider.ID(rawID)
		if _, ok := s.dependencies.ProviderRegistry.Get(providerID); !ok {
			return devcontext.Context{}, fmt.Errorf("unknown provider %q", providerID)
		}
		providers[providerID] = provider.Config{Enabled: true}
	}
	metadata := devcontext.Metadata{}
	for key, value := range map[string]string{"purpose": request.Purpose, "description": request.Description, "icon": request.Icon, "accent": request.Accent} {
		if value = strings.TrimSpace(value); value != "" {
			metadata[key] = value
		}
	}
	return devcontext.Context{ID: contextID, Name: strings.TrimSpace(request.Name), Tool: codingtool.LaunchTarget{DefaultTool: toolID, Tools: map[codingtool.ID]codingtool.Config{toolID: {}}}, Providers: providers, Metadata: metadata, CreatedAt: s.now().UTC()}, nil
}

// applyDevelopmentToolSelections translates the creation flow's generic
// registry IDs into the persisted tool and provider configuration. Registry
// IDs must not be shared across the two integration types: accepting an
// ambiguous ID could silently configure the wrong integration.
func (s *Service) applyDevelopmentToolSelections(request *CreateContextRequest) error {
	if len(request.EnabledDevelopmentToolIDs) == 0 {
		return nil
	}

	providerIDs := make([]string, 0, len(request.EnabledDevelopmentToolIDs))
	seen := make(map[string]struct{}, len(request.EnabledDevelopmentToolIDs))
	for _, rawID := range request.EnabledDevelopmentToolIDs {
		id := strings.TrimSpace(rawID)
		if id == "" {
			return fmt.Errorf("development tool selection has an empty ID")
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		_, isTool := s.dependencies.ToolRegistry.Get(codingtool.ID(id))
		_, isProvider := s.dependencies.ProviderRegistry.Get(provider.ID(id))
		switch {
		case isTool && isProvider:
			return fmt.Errorf("development tool ID %q is ambiguous", id)
		case isTool:
			if request.ToolID != "" && request.ToolID != id {
				return fmt.Errorf("select only one coding tool")
			}
			request.ToolID = id
		case isProvider:
			providerIDs = append(providerIDs, id)
		default:
			return fmt.Errorf("unknown development tool %q", id)
		}
	}
	request.EnabledProviderIDs = providerIDs
	return nil
}

func (s *Service) duplicateContext(request DuplicateContextRequest) (DuplicateContextResult, error) {
	sourceID, err := devcontext.NewID(request.SourceContextID)
	if err != nil {
		return DuplicateContextResult{}, err
	}
	targetID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return DuplicateContextResult{}, err
	}
	source, err := s.dependencies.Contexts.Get(sourceID)
	if err != nil {
		return DuplicateContextResult{}, err
	}
	name := strings.TrimSpace(request.Name)
	if name == "" {
		name = source.Name + " copy"
	}
	copy := devcontext.Context{
		ID: targetID, Name: name, Tool: cloneLaunchTarget(source.Tool),
		Providers: cloneProviderConfigs(source.Providers), Metadata: cloneMetadata(source.Metadata), CreatedAt: s.now().UTC(),
	}
	paths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, targetID)
	if err != nil {
		return DuplicateContextResult{}, err
	}
	if err := filesystem.CreateContextDirectoryTreeWithRegistriesCredentialsAndPermissions(
		s.dependencies.Paths, paths, copy, s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry,
		nil, s.dependencies.StoragePermissions,
	); err != nil {
		return DuplicateContextResult{}, err
	}
	s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{Name: devlog.EventContextCreated, Timestamp: s.now(), ContextID: copy.ID.String(), ToolID: string(copy.Tool.DefaultTool)}))
	return DuplicateContextResult{Context: s.contextState(copy)}, nil
}

func (s *Service) exportContextMetadata(request ExportContextMetadataRequest) (ContextMetadataExport, error) {
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return ContextMetadataExport{}, err
	}
	ctx, err := s.dependencies.Contexts.Get(contextID)
	if err != nil {
		return ContextMetadataExport{}, err
	}

	providers := make([]ContextTransferProvider, 0, len(ctx.Providers))
	for _, providerID := range sortedProviderConfigIDs(ctx.Providers) {
		config := ctx.Providers[providerID]
		providers = append(providers, ContextTransferProvider{ID: string(providerID), Enabled: config.Enabled, Options: cloneStringMap(config.Options)})
	}
	tools := make([]ContextTransferTool, 0, len(ctx.Tool.Tools))
	for _, toolID := range sortedToolConfigIDs(ctx.Tool.Tools) {
		config := ctx.Tool.Tools[toolID]
		tools = append(tools, ContextTransferTool{ID: string(toolID), Options: cloneStringMap(config.Options)})
	}

	return ContextMetadataExport{
		Version: ContextTransferVersion,
		Context: ContextTransferMetadata{
			Name: ctx.Name, Metadata: cloneStringMap(ctx.Metadata), Providers: providers,
			LaunchTarget: ContextTransferLaunchTarget{DefaultTool: string(ctx.Tool.DefaultTool), Tools: tools},
		},
	}, nil
}

func (s *Service) importContextMetadata(request ImportContextMetadataRequest) (ImportContextMetadataResult, error) {
	contextID, err := devcontext.NewID(request.ContextID)
	if err != nil {
		return ImportContextMetadataResult{}, err
	}
	ctx, err := s.contextFromMetadataExport(contextID, request.Export)
	if err != nil {
		return ImportContextMetadataResult{}, err
	}
	paths, err := filesystem.DeriveContextPaths(s.dependencies.Paths, contextID)
	if err != nil {
		return ImportContextMetadataResult{}, err
	}
	if err := filesystem.CreateContextDirectoryTreeWithRegistriesAndPermissions(
		paths, ctx, s.dependencies.ProviderRegistry, s.dependencies.ToolRegistry, s.dependencies.StoragePermissions,
	); err != nil {
		return ImportContextMetadataResult{}, err
	}
	s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{Name: devlog.EventContextCreated, Timestamp: s.now(), ContextID: ctx.ID.String(), ToolID: string(ctx.Tool.DefaultTool)}))
	return ImportContextMetadataResult{Context: s.contextState(ctx)}, nil
}

func (s *Service) contextFromMetadataExport(contextID devcontext.ID, exported ContextMetadataExport) (devcontext.Context, error) {
	if exported.Version != ContextTransferVersion {
		return devcontext.Context{}, fmt.Errorf("%w: unsupported context metadata export version %d", devcontext.ErrInvalidContextConfig, exported.Version)
	}
	if strings.TrimSpace(exported.Context.Name) == "" {
		return devcontext.Context{}, fmt.Errorf("%w: imported context name cannot be empty", devcontext.ErrInvalidContextConfig)
	}
	defaultTool := codingtool.ID(exported.Context.LaunchTarget.DefaultTool)
	if _, ok := s.dependencies.ToolRegistry.Get(defaultTool); !ok {
		return devcontext.Context{}, fmt.Errorf("%w: unknown coding tool %q", devcontext.ErrInvalidContextConfig, defaultTool)
	}

	tools := make(map[codingtool.ID]codingtool.Config, len(exported.Context.LaunchTarget.Tools)+1)
	for _, exportedTool := range exported.Context.LaunchTarget.Tools {
		toolID := codingtool.ID(exportedTool.ID)
		if toolID == "" {
			return devcontext.Context{}, fmt.Errorf("%w: imported coding tool ID cannot be empty", devcontext.ErrInvalidContextConfig)
		}
		if _, ok := s.dependencies.ToolRegistry.Get(toolID); !ok {
			return devcontext.Context{}, fmt.Errorf("%w: unknown coding tool %q", devcontext.ErrInvalidContextConfig, toolID)
		}
		if _, exists := tools[toolID]; exists {
			return devcontext.Context{}, fmt.Errorf("%w: duplicate imported coding tool %q", devcontext.ErrInvalidContextConfig, toolID)
		}
		if err := validateNonEmptyOptionKeys("tool", exportedTool.ID, exportedTool.Options); err != nil {
			return devcontext.Context{}, err
		}
		tools[toolID] = codingtool.Config{Options: cloneStringMap(exportedTool.Options)}
	}
	if _, ok := tools[defaultTool]; !ok {
		tools[defaultTool] = codingtool.Config{}
	}

	providers := make(provider.Configs, len(exported.Context.Providers))
	for _, exportedProvider := range exported.Context.Providers {
		providerID := provider.ID(exportedProvider.ID)
		if providerID == "" {
			return devcontext.Context{}, fmt.Errorf("%w: imported provider ID cannot be empty", devcontext.ErrInvalidContextConfig)
		}
		if _, ok := s.dependencies.ProviderRegistry.Get(providerID); !ok {
			return devcontext.Context{}, fmt.Errorf("%w: unknown provider %q", devcontext.ErrInvalidContextConfig, providerID)
		}
		if _, exists := providers[providerID]; exists {
			return devcontext.Context{}, fmt.Errorf("%w: duplicate imported provider %q", devcontext.ErrInvalidContextConfig, providerID)
		}
		if err := validateNonEmptyOptionKeys("provider", exportedProvider.ID, exportedProvider.Options); err != nil {
			return devcontext.Context{}, err
		}
		providers[providerID] = provider.Config{Enabled: exportedProvider.Enabled, Options: cloneStringMap(exportedProvider.Options)}
	}
	if err := validateNonEmptyOptionKeys("metadata", "", exported.Context.Metadata); err != nil {
		return devcontext.Context{}, err
	}

	return devcontext.Context{
		ID: contextID, Name: strings.TrimSpace(exported.Context.Name),
		Tool:      codingtool.LaunchTarget{DefaultTool: defaultTool, Tools: tools},
		Providers: providers, Metadata: devcontext.Metadata(cloneStringMap(exported.Context.Metadata)), CreatedAt: s.now().UTC(),
	}, nil
}

func validateNonEmptyOptionKeys(kind, id string, values map[string]string) error {
	for key := range values {
		if key == "" {
			if id == "" {
				return fmt.Errorf("%w: imported %s key cannot be empty", devcontext.ErrInvalidContextConfig, kind)
			}
			return fmt.Errorf("%w: imported %s %q option key cannot be empty", devcontext.ErrInvalidContextConfig, kind, id)
		}
	}
	return nil
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func sortedProviderConfigIDs(configs provider.Configs) []provider.ID {
	ids := make([]provider.ID, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func sortedToolConfigIDs(configs map[codingtool.ID]codingtool.Config) []codingtool.ID {
	ids := make([]codingtool.ID, 0, len(configs))
	for id := range configs {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func cloneLaunchTarget(source codingtool.LaunchTarget) codingtool.LaunchTarget {
	tools := make(map[codingtool.ID]codingtool.Config, len(source.Tools))
	for id, config := range source.Tools {
		options := make(map[string]string, len(config.Options))
		for key, value := range config.Options {
			options[key] = value
		}
		tools[id] = codingtool.Config{ExecutableOverride: config.ExecutableOverride, Options: options}
	}
	return codingtool.LaunchTarget{DefaultTool: source.DefaultTool, Tools: tools}
}

func cloneProviderConfigs(source provider.Configs) provider.Configs {
	configs := make(provider.Configs, len(source))
	for id, config := range source {
		options := make(map[string]string, len(config.Options))
		for key, value := range config.Options {
			options[key] = value
		}
		configs[id] = provider.Config{Enabled: config.Enabled, Options: options}
	}
	return configs
}

func (s *Service) getProjects() ([]ProjectListItem, error) {
	bindings, err := s.dependencies.Projects.List()
	if err != nil {
		return nil, err
	}
	recents, err := s.dependencies.RecentProjects.List()
	if err != nil {
		return nil, err
	}
	contexts, err := s.dependencies.Contexts.List()
	if err != nil {
		return nil, err
	}
	names := map[devcontext.ID]string{}
	for _, ctx := range contexts {
		names[ctx.ID] = ctx.Name
	}
	items := map[project.Path]ProjectListItem{}
	for _, binding := range bindings {
		items[binding.ProjectPath] = ProjectListItem{Project: projectState(binding.ProjectPath), ContextID: binding.ContextID.String(), ContextName: names[binding.ContextID]}
	}
	for _, recent := range recents {
		item := items[recent.ProjectPath]
		if item.Project.Path == "" {
			item.Project = projectState(recent.ProjectPath)
		}
		timestamp := recent.LastLaunchedAt.UTC()
		item.LastLaunchedAt = &timestamp
		items[recent.ProjectPath] = item
	}
	result := make([]ProjectListItem, 0, len(items))
	for _, item := range items {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].LastLaunchedAt == nil {
			return false
		}
		if result[j].LastLaunchedAt == nil {
			return true
		}
		return result[i].LastLaunchedAt.After(*result[j].LastLaunchedAt)
	})
	return result, nil
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
	if err := s.exportCodingToolStatus(plan); err != nil {
		return LaunchProjectResult{}, err
	}

	if err := s.processLauncher().Launch(processRequestFromLaunchPlan(plan, s.dependencies.DetachMode)); err != nil {
		s.recordLaunchEvent(eventFromLaunchPlan(devlog.EventLaunchProcessFailure, plan, err, s.now()))
		return LaunchProjectResult{}, newLaunchFailureError(err, plan.Executable, plan.Environment, s.now())
	}

	s.recordLaunchEvent(eventFromLaunchPlan(devlog.EventLaunchSucceeded, plan, nil, s.now()))
	_ = s.dependencies.RecentProjects.Record(plan.ProjectPath, plan.Context.ID, s.now())
	_, _ = s.dependencies.RunningEnvironments.Record(runningEnvironmentFromLaunchPlan(plan, s.now()))

	return LaunchProjectResult{
		Project:  projectState(plan.ProjectPath),
		Context:  s.contextState(plan.Context),
		Warnings: warningStates(plan.Warnings),
	}, nil
}

func runningEnvironmentFromLaunchPlan(plan launcher.LaunchPlan, startedAt time.Time) coreRunning.Environment {
	return coreRunning.Environment{
		Project:   coreRunning.ProjectIdentity{Path: plan.ProjectPath, Name: projectName(plan.ProjectPath)},
		Context:   coreRunning.ContextIdentity{ID: plan.Context.ID, Name: plan.Context.Name},
		Tool:      coreRunning.ToolIdentity{ID: plan.Tool.ID, Name: plan.Tool.DisplayName},
		StartedAt: startedAt.UTC(),
		Process:   coreRunning.Process{State: coreRunning.ProcessStateRunning},
		Session:   coreRunning.Session{State: coreRunning.SessionStateUnknown},
		Launch: coreRunning.LaunchIdentity{
			Source:           launcher.InvocationSourceGUI,
			ResolutionSource: plan.ResolutionSource,
		},
	}
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
	runningConflict, err := s.runningEnvironmentConflict(projectPath, contextID)
	if err != nil {
		return PreflightLaunchProjectResult{}, err
	}
	warnings := warningStates(resolution.Warnings)
	return PreflightLaunchProjectResult{
		Project:                    projectState(projectPath),
		Context:                    contextState,
		Confidence:                 contextState.Confidence,
		Groups:                     preflightGroups(contextState, warnings, runningConflict),
		VerificationSteps:          s.launchVerificationSteps(contextState, projectState(projectPath)),
		Warnings:                   warnings,
		RunningEnvironmentConflict: runningConflict,
	}, nil
}

// preflightGroups maps implementation-level readiness signals to the five
// product areas rendered by the launcher. The blocking policy is deliberately
// owned here so every frontend uses the same safe continuation rule.
func preflightGroups(context ContextState, warnings []ResolutionWarning, conflict *RunningEnvironmentConflict) []PreflightGroup {
	return []PreflightGroup{
		newPreflightGroup(
			PreflightGroupProject,
			"Project",
			projectPreflightChecks(warnings),
		),
		newPreflightGroup(
			PreflightGroupContext,
			"Context",
			contextPreflightChecks(context),
		),
		newPreflightGroup(
			PreflightGroupIsolation,
			"Isolation",
			confidencePreflightChecks(
				"isolation",
				checksForComponent(
					context.Confidence.Checks,
					LaunchConfidenceCheckIsolation,
				),
			),
		),
		newPreflightGroup(PreflightGroupTools, "Tools", toolPreflightChecks(context)),
		newPreflightGroup(
			PreflightGroupWorkspace,
			"Workspace",
			workspacePreflightChecks(conflict),
		),
	}
}

func projectPreflightChecks(warnings []ResolutionWarning) []PreflightCheck {
	checks := []PreflightCheck{
		newPreflightCheck("project_directory", "Project folder", LaunchConfidenceReady, "Project folder is ready.", ""),
	}
	for index, warning := range warnings {
		checks = append(checks, newPreflightCheck(
			fmt.Sprintf("project_%s_%d", warning.Code, index),
			"Project binding",
			LaunchConfidenceNeedsAttention,
			warning.Message,
			"Review the context selected for this project.",
		))
	}
	return checks
}

func contextPreflightChecks(context ContextState) []PreflightCheck {
	checks := []PreflightCheck{
		newPreflightCheck("context_selection", "Selected context", LaunchConfidenceReady, context.Name+" is ready to use.", ""),
	}
	for index, check := range checksForComponent(context.Confidence.Checks, LaunchConfidenceCheckIdentity) {
		checks = append(checks, confidencePreflightCheck(fmt.Sprintf("identity_%d", index), check))
	}
	return checks
}

func toolPreflightChecks(context ContextState) []PreflightCheck {
	checks := confidencePreflightChecks("provider", checksForComponent(context.Confidence.Checks, LaunchConfidenceCheckProvider))
	return append(checks, confidencePreflightChecks("tool", checksForComponent(context.Confidence.Checks, LaunchConfidenceCheckTool))...)
}

func workspacePreflightChecks(conflict *RunningEnvironmentConflict) []PreflightCheck {
	if conflict == nil {
		return []PreflightCheck{
			newPreflightCheck("workspace_availability", "Active workspaces", LaunchConfidenceReady, "No active workspace is using this project.", ""),
		}
	}

	message := "This project already has an active workspace."
	if conflict.Kind == "different_context" {
		message = "This project is already open in another context."
	}
	return []PreflightCheck{
		newPreflightCheck("workspace_conflict", "Active workspace", LaunchConfidenceNeedsAttention, message, "Review the existing workspace before continuing."),
	}
}

func confidencePreflightChecks(prefix string, confidenceChecks []LaunchConfidenceCheck) []PreflightCheck {
	checks := make([]PreflightCheck, 0, len(confidenceChecks))
	for index, check := range confidenceChecks {
		checks = append(checks, confidencePreflightCheck(fmt.Sprintf("%s_%d", prefix, index), check))
	}
	return checks
}

func confidencePreflightCheck(id string, check LaunchConfidenceCheck) PreflightCheck {
	return newPreflightCheck(id, check.Label, check.Severity, check.Message, check.ActionHint)
}

func newPreflightCheck(id string, label string, status LaunchConfidenceStatus, message string, actionHint string) PreflightCheck {
	return PreflightCheck{
		ID:         id,
		Label:      label,
		Status:     status,
		Blocking:   preflightStatusBlocksLaunch(status),
		Message:    message,
		ActionHint: actionHint,
	}
}

// preflightStatusBlocksLaunch is the single severity policy for preflight
// evidence. Blocked checks require remediation; needs-attention checks can
// continue only through an explicit UI decision in a later phase.
func preflightStatusBlocksLaunch(status LaunchConfidenceStatus) bool {
	return status == LaunchConfidenceBlocked
}

func newPreflightGroup(id PreflightGroupID, label string, checks []PreflightCheck) PreflightGroup {
	status := LaunchConfidenceReady
	message := label + " is ready."
	blocking := false
	for _, check := range checks {
		if check.Status == LaunchConfidenceBlocked {
			status = LaunchConfidenceBlocked
			message = check.Message
			blocking = true
			break
		}
		if check.Status == LaunchConfidenceNeedsAttention && status == LaunchConfidenceReady {
			status = LaunchConfidenceNeedsAttention
			message = check.Message
		}
	}
	return PreflightGroup{
		ID:       id,
		Label:    label,
		Status:   status,
		Blocking: blocking,
		Message:  message,
		Checks:   checks,
	}
}

// launchVerificationSteps describes work that starts only after a preflight
// succeeds. Preflight proves readiness; it does not build the launch plan,
// write tool status, or start a process, so these stages must remain pending.
func (s *Service) launchVerificationSteps(context ContextState, project ProjectState) []LaunchVerificationStep {
	steps := []LaunchVerificationStep{
		{
			ID:      "prepare_environment",
			Label:   "Prepare isolated environment",
			Status:  LaunchVerificationStepPending,
			Message: "Building the isolated launch environment.",
		},
	}

	if codingToolConsumesStatusData(s.dependencies.ToolRegistry, context.Tool.ID) {
		steps = append(steps, LaunchVerificationStep{
			ID:      "write_tool_status",
			Label:   "Apply " + context.Tool.Name + " settings",
			Status:  LaunchVerificationStepPending,
			Message: "Writing safe context settings for " + context.Tool.Name + ".",
		})
	}

	return append(steps,
		LaunchVerificationStep{
			ID:      "start_tool",
			Label:   "Start " + context.Tool.Name,
			Status:  LaunchVerificationStepPending,
			Message: "Starting " + context.Tool.Name + ".",
		},
		LaunchVerificationStep{
			ID:      "open_project",
			Label:   "Open " + project.Name,
			Status:  LaunchVerificationStepPending,
			Message: "Opening " + project.Name + " in " + context.Tool.Name + ".",
		},
	)
}

func checksForComponent(checks []LaunchConfidenceCheck, component LaunchConfidenceCheckComponent) []LaunchConfidenceCheck {
	result := make([]LaunchConfidenceCheck, 0)
	for _, check := range checks {
		if check.Component == component {
			result = append(result, check)
		}
	}
	return result
}

func codingToolConsumesStatusData(registry codingtool.Registry, toolID string) bool {
	registered, ok := registry.Lookup(codingtool.ID(toolID))
	if !ok {
		return false
	}
	_, ok = registered.Integration.(codingtool.StatusDataConsumer)
	return ok
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
	s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{
		Name:        devlog.EventProjectBindingChanged,
		Timestamp:   s.now(),
		ProjectPath: string(binding.ProjectPath),
		ContextID:   binding.ContextID.String(),
	}))

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
	if result.Removed {
		s.recordHistoryEvent(devlog.NewEvent(devlog.EventInput{
			Name:        devlog.EventProjectBindingChanged,
			Timestamp:   s.now(),
			ProjectPath: string(result.ProjectPath),
			ContextID:   result.Binding.ContextID.String(),
		}))
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
		Purpose:        ctx.Metadata["purpose"],
		Description:    ctx.Metadata["description"],
		Tool:           toolState(ctx.Tool.DefaultTool, confidence),
		AvailableTools: toolOptions(s.dependencies.ToolRegistry),
		Providers:      providerStatesFromEntries(providerEntries),
		DevelopmentTools: developmentToolIntegrations(
			ctx,
			providerEntries,
			confidence,
			s.dependencies.ToolRegistry,
			s.dependencies.ProviderRegistry,
		),
		Confidence: confidence,
		Metadata:   cloneMetadata(ctx.Metadata),
	}
}

func enabledProviderStates(providers []ProviderState) []ProviderState {
	enabled := make([]ProviderState, 0, len(providers))
	for _, provider := range providers {
		if provider.Enabled {
			enabled = append(enabled, provider)
		}
	}
	return enabled
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
		identity := providerIdentityState(integration, enabled, status, runtime, pathsErr)
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
				SetupAction: providerSetupAction(integration, enabled, status, identity, runtime),
				Identity:    identity,
			},
		})
	}
	return entries
}

func providerSetupAction(integration provider.Provider, enabled bool, status provider.Status, identity ProviderIdentityState, runtime provider.RuntimeContext) *ProviderSetupAction {
	if !enabled {
		return nil
	}

	switch status.State {
	case provider.StatusNotConfigured, provider.StatusDirectoryMissing:
		return &ProviderSetupAction{
			State:   ProviderSetupOpenAndConfigure,
			Label:   "Open and configure",
			Message: providerSetupMessage(integration, runtime),
		}
	case provider.StatusConfigured:
		if identity.Status == ProviderIdentityVerified {
			return &ProviderSetupAction{
				State:   ProviderSetupVerified,
				Label:   "Verified",
				Message: integration.DisplayName() + " account identity is verified for this context.",
			}
		}
		return &ProviderSetupAction{
			State:   ProviderSetupWaitingForSignIn,
			Label:   "Waiting for sign-in",
			Message: "Waiting for " + integration.DisplayName() + " sign-in verification.",
		}
	default:
		return nil
	}
}

func providerSetupMessage(integration provider.Provider, runtime provider.RuntimeContext) string {
	if guidanceProvider, ok := integration.(provider.SetupGuidanceProvider); ok {
		if message := guidanceProvider.SetupGuidance(runtime).Message; message != "" {
			return message
		}
	}
	return integration.DisplayName() + " needs to be configured for this context."
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
	if identityCheck, ok := launcher.AccountIdentityMismatchConfidenceCheck(identityEvidence(providerEntries)); ok {
		checks = append(checks, identityCheck)
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

func identityEvidence(entries []providerStateEntry) []launcher.AccountIdentityEvidence {
	evidence := make([]launcher.AccountIdentityEvidence, 0, len(entries))
	for _, entry := range entries {
		if !entry.state.Enabled || entry.state.Identity.Status != ProviderIdentityVerified {
			continue
		}
		fields := make([]launcher.AccountIdentityField, 0, len(entry.state.Identity.Fields))
		for _, field := range entry.state.Identity.Fields {
			fields = append(fields, launcher.AccountIdentityField{Label: field.Label, Value: field.Value})
		}
		evidence = append(evidence, launcher.AccountIdentityEvidence{ProviderID: entry.state.ID, Fields: fields})
	}
	return evidence
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

func developmentToolIntegrations(
	ctx devcontext.Context,
	providerEntries []providerStateEntry,
	confidence LaunchConfidenceState,
	toolRegistry codingtool.Registry,
	providerRegistry provider.Registry,
) []DevelopmentToolIntegration {
	tools := toolRegistry.All()
	integrations := make([]DevelopmentToolIntegration, 0, len(tools)+len(providerEntries))
	for _, tool := range tools {
		selected := tool.Integration.ID() == ctx.Tool.DefaultTool
		status := DevelopmentToolAvailable
		message := "Available to add to this context."
		var recoveryHint string
		if selected {
			state := toolState(tool.Integration.ID(), confidence)
			status = developmentToolStatusForTool(state.Status)
			message = state.Message
			recoveryHint = state.ActionHint
		}
		integrations = append(integrations, DevelopmentToolIntegration{
			ID:           string(tool.Integration.ID()),
			Name:         tool.DisplayName,
			Category:     developmentToolCategory(tool.Category),
			Status:       status,
			Message:      message,
			RecoveryHint: recoveryHint,
			Enabled:      selected,
		})
	}

	for _, entry := range providerEntries {
		status, message, recoveryHint := developmentToolStatusForProvider(entry.state)
		integrations = append(integrations, DevelopmentToolIntegration{
			ID:           entry.state.ID,
			Name:         entry.state.Name,
			Category:     developmentToolCategory(providerRegistry.Category(entry.providerID)),
			Status:       status,
			Message:      message,
			RecoveryHint: recoveryHint,
			Enabled:      entry.state.Enabled,
		})
	}
	return integrations
}

func developmentToolCategory(value string) DevelopmentToolCategory {
	switch DevelopmentToolCategory(value) {
	case DevelopmentToolCategoryCoding,
		DevelopmentToolCategoryAI,
		DevelopmentToolCategoryVersionControl,
		DevelopmentToolCategorySourceHosting,
		DevelopmentToolCategoryCloudRegistries:
		return DevelopmentToolCategory(value)
	default:
		return DevelopmentToolCategoryOther
	}
}

func developmentToolStatusForTool(status LaunchConfidenceStatus) DevelopmentToolStatus {
	switch status {
	case LaunchConfidenceReady:
		return DevelopmentToolAvailable
	case LaunchConfidenceNeedsAttention, LaunchConfidenceBlocked:
		return DevelopmentToolUnavailable
	default:
		return DevelopmentToolError
	}
}

func developmentToolStatusForProvider(state ProviderState) (DevelopmentToolStatus, string, string) {
	if !state.Enabled {
		return DevelopmentToolAvailable, "Available to add to this context.", ""
	}
	if state.SetupAction != nil {
		switch state.SetupAction.State {
		case ProviderSetupWaitingForSignIn:
			return DevelopmentToolNeedsSignIn, state.SetupAction.Message, state.SetupAction.Label
		case ProviderSetupVerified:
			return DevelopmentToolConnected, state.SetupAction.Message, ""
		case ProviderSetupOpenAndConfigure:
			return DevelopmentToolNotConfigured, state.SetupAction.Message, state.SetupAction.Label
		}
	}
	switch state.State {
	case ProviderReadinessReady:
		return DevelopmentToolConnected, state.Explanation, ""
	case ProviderReadinessNotConfigured:
		return DevelopmentToolNotConfigured, state.Explanation, state.ActionHint
	case ProviderReadinessDirectoryMissing:
		return DevelopmentToolError, state.Explanation, state.ActionHint
	case ProviderReadinessUnavailable:
		return DevelopmentToolUnavailable, state.Explanation, state.ActionHint
	default:
		return DevelopmentToolError, "Integration status could not be determined.", ""
	}
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

func (s *Service) recordHistoryEvent(event devlog.Event) {
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
