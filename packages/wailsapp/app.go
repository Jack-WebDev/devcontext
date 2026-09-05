package wailsapp

import (
	"context"

	"devctx/packages/application"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type service interface {
	ValidateProjectDirectory(application.ValidateProjectDirectoryRequest) (application.ProjectState, *application.Error)
	GetLaunchState(application.GetLaunchStateRequest) (application.LaunchState, *application.Error)
	GetHomeDashboard(application.GetHomeDashboardRequest) (application.HomeDashboardState, *application.Error)
	GetRecentProjects() (application.RecentProjectsState, *application.Error)
	GetContexts() (application.ContextListState, *application.Error)
	GetContextDetails(application.GetContextDetailsRequest) (application.ContextDetailsState, *application.Error)
	UpdateContextDetails(application.UpdateContextDetailsRequest) (application.ContextState, *application.Error)
	UpdateContextAppearance(application.UpdateContextAppearanceRequest) (application.ContextState, *application.Error)
	GetTrustCenter() (application.TrustCenterState, *application.Error)
	PreflightLaunchProject(application.PreflightLaunchProjectRequest) (application.PreflightLaunchProjectResult, *application.Error)
	LaunchProject(application.LaunchProjectRequest) (application.LaunchProjectResult, *application.Error)
	BindProject(application.BindProjectRequest) (application.ProjectBindingState, *application.Error)
	UnbindProject(application.UnbindProjectRequest) (application.ProjectBindingState, *application.Error)
	CreateContext(application.CreateContextRequest) (application.CreateContextResult, *application.Error)
	GetContextTemplates() application.ContextTemplatesState
	DuplicateContext(application.DuplicateContextRequest) (application.DuplicateContextResult, *application.Error)
	ExportContextMetadata(application.ExportContextMetadataRequest) (application.ContextMetadataExport, *application.Error)
	ImportContextMetadata(application.ImportContextMetadataRequest) (application.ImportContextMetadataResult, *application.Error)
	GetProjects() (application.ProjectsState, *application.Error)
	GetDiagnostics(application.GetDiagnosticsRequest) (application.DiagnosticsState, *application.Error)
	GetRepairActions(application.GetRepairActionsRequest) (application.RepairActionsState, *application.Error)
	RunRepairAction(application.RunRepairActionRequest) (application.RunRepairActionResult, *application.Error)
	GetHistory() (application.HistoryState, *application.Error)
	GetRunningEnvironments() (application.RunningEnvironmentsState, *application.Error)
	GetSettings() (application.SettingsState, *application.Error)
	UpdateSettings(application.UpdateSettingsRequest) (application.SettingsState, *application.Error)
	GetTrayState() (application.TrayState, *application.Error)
}

func (a *App) GetSettings() any {
	settings, err := a.service.GetSettings()
	if err != nil {
		return err
	}
	return settings
}
func (a *App) UpdateSettings(request application.UpdateSettingsRequest) any {
	settings, err := a.service.UpdateSettings(request)
	if err != nil {
		return err
	}
	return settings
}
func (a *App) GetTrayState() any {
	state, err := a.service.GetTrayState()
	if err != nil {
		return err
	}
	return state
}

func (a *App) GetProjects() any {
	projects, err := a.service.GetProjects()
	if err != nil {
		return err
	}
	return projects
}

// GetDiagnostics returns structured, presentation-safe diagnostics.
func (a *App) GetDiagnostics(request application.GetDiagnosticsRequest) any {
	diagnostics, err := a.service.GetDiagnostics(request)
	if err != nil {
		return err
	}
	return diagnostics
}

// GetRepairActions returns repair options and destructive-impact previews.
func (a *App) GetRepairActions(request application.GetRepairActionsRequest) any {
	actions, err := a.service.GetRepairActions(request)
	if err != nil {
		return err
	}
	return actions
}

// RunRepairAction executes one repair action after backend confirmation checks.
func (a *App) RunRepairAction(request application.RunRepairActionRequest) any {
	result, err := a.service.RunRepairAction(request)
	if err != nil {
		return err
	}
	return result
}

// GetHistory returns local user-facing activity records.
func (a *App) GetHistory() any {
	history, err := a.service.GetHistory()
	if err != nil {
		return err
	}
	return history
}

// GetRunningEnvironments returns active coding-tool environments.
func (a *App) GetRunningEnvironments() any {
	environments, err := a.service.GetRunningEnvironments()
	if err != nil {
		return err
	}
	return environments
}

// GetHomeDashboard returns the Home screen state for the requested project.
func (a *App) GetHomeDashboard(request application.GetHomeDashboardRequest) any {
	dashboard, err := a.service.GetHomeDashboard(request)
	if err != nil {
		return err
	}
	return dashboard
}

// GetRecentProjects returns recent successful launches for presentation.
func (a *App) GetRecentProjects() any {
	projects, err := a.service.GetRecentProjects()
	if err != nil {
		return err
	}
	return projects
}

// GetContexts returns configured context summaries for presentation.
func (a *App) GetContexts() any {
	contexts, err := a.service.GetContexts()
	if err != nil {
		return err
	}
	return contexts
}

// GetContextDetails returns one configured context's presentation-safe details.
func (a *App) GetContextDetails(request application.GetContextDetailsRequest) any {
	details, err := a.service.GetContextDetails(request)
	if err != nil {
		return err
	}
	return details
}

// UpdateContextDetails updates only a context's display identity.
func (a *App) UpdateContextDetails(request application.UpdateContextDetailsRequest) any {
	context, err := a.service.UpdateContextDetails(request)
	if err != nil {
		return err
	}
	return context
}

// UpdateContextAppearance updates only a context's icon and accent metadata.
func (a *App) UpdateContextAppearance(request application.UpdateContextAppearanceRequest) any {
	context, err := a.service.UpdateContextAppearance(request)
	if err != nil {
		return err
	}
	return context
}

// GetTrustCenter returns factual local protection and integration-boundary
// data for the Trust Center.
func (a *App) GetTrustCenter() any {
	state, err := a.service.GetTrustCenter()
	if err != nil {
		return err
	}
	return state
}

// App is the Wails-bound application surface.
type App struct {
	ctx     context.Context
	service service
	mode    ApplicationMode
}

// New creates the application surface bound by Wails.
func New(service service, mode ApplicationMode) *App {
	return &App{service: service, mode: mode}
}

// Startup stores the process context supplied by Wails.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// GetApplicationMode returns the startup intent selected by the desktop host.
func (a *App) GetApplicationMode() ApplicationMode {
	return a.mode
}

// ChooseProjectDirectory opens the host-owned folder picker used by focused
// launcher recovery flows. An empty result means the user canceled it.
func (a *App) ChooseProjectDirectory() any {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose a project folder",
	})
	if err != nil {
		return application.NewError(err)
	}
	return path
}

// ValidateProjectDirectory validates a folder selected for a context draft.
func (a *App) ValidateProjectDirectory(request application.ValidateProjectDirectoryRequest) any {
	project, err := a.service.ValidateProjectDirectory(request)
	if err != nil {
		return err
	}
	return project
}

// GetLaunchState returns selector state for the current project.
func (a *App) GetLaunchState(request application.GetLaunchStateRequest) any {
	state, err := a.service.GetLaunchState(request)
	if err != nil {
		return err
	}
	return state
}

// PreflightLaunchProject checks the selected launch without opening an codingtool.
func (a *App) PreflightLaunchProject(request application.PreflightLaunchProjectRequest) any {
	result, err := a.service.PreflightLaunchProject(request)
	if err != nil {
		return err
	}
	return result
}

// LaunchProject opens the selected project and context.
func (a *App) LaunchProject(request application.LaunchProjectRequest) any {
	result, err := a.service.LaunchProject(request)
	if err != nil {
		return err
	}
	return result
}

// BindProject remembers a context for a project.
func (a *App) BindProject(request application.BindProjectRequest) any {
	state, err := a.service.BindProject(request)
	if err != nil {
		return err
	}
	return state
}

// UnbindProject removes the remembered context for a project.
func (a *App) UnbindProject(request application.UnbindProjectRequest) any {
	state, err := a.service.UnbindProject(request)
	if err != nil {
		return err
	}
	return state
}

// CreateContext creates a default context during first-run onboarding.
func (a *App) CreateContext(request application.CreateContextRequest) any {
	result, err := a.service.CreateContext(request)
	if err != nil {
		return err
	}
	return result
}

// GetContextTemplates returns safe create-context defaults.
func (a *App) GetContextTemplates() any { return a.service.GetContextTemplates() }

// DuplicateContext copies safe context configuration into a new isolation boundary.
func (a *App) DuplicateContext(request application.DuplicateContextRequest) any {
	result, err := a.service.DuplicateContext(request)
	if err != nil {
		return err
	}
	return result
}

// ExportContextMetadata returns a portable context configuration without
// credentials or integration-owned storage.
func (a *App) ExportContextMetadata(request application.ExportContextMetadataRequest) any {
	result, err := a.service.ExportContextMetadata(request)
	if err != nil {
		return err
	}
	return result
}

// ImportContextMetadata creates a fresh isolated context from portable safe
// metadata. Credential import remains intentionally unavailable.
func (a *App) ImportContextMetadata(request application.ImportContextMetadataRequest) any {
	result, err := a.service.ImportContextMetadata(request)
	if err != nil {
		return err
	}
	return result
}
