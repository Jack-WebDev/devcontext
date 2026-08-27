package wailsapp

import (
	"context"

	"devctx/packages/application"
)

type service interface {
	GetLaunchState(application.GetLaunchStateRequest) (application.LaunchState, *application.Error)
	GetHomeDashboard(application.GetHomeDashboardRequest) (application.HomeDashboardState, *application.Error)
	PreflightLaunchProject(application.PreflightLaunchProjectRequest) (application.PreflightLaunchProjectResult, *application.Error)
	LaunchProject(application.LaunchProjectRequest) (application.LaunchProjectResult, *application.Error)
	BindProject(application.BindProjectRequest) (application.ProjectBindingState, *application.Error)
	UnbindProject(application.UnbindProjectRequest) (application.ProjectBindingState, *application.Error)
	CreateContext(application.CreateContextRequest) (application.CreateContextResult, *application.Error)
}

// GetHomeDashboard returns the Home screen state for the requested project.
func (a *App) GetHomeDashboard(request application.GetHomeDashboardRequest) any {
	dashboard, err := a.service.GetHomeDashboard(request)
	if err != nil {
		return err
	}
	return dashboard
}

// App is the Wails-bound application surface.
type App struct {
	ctx     context.Context
	service service
}

// New creates the application surface bound by Wails.
func New(service service) *App {
	return &App{service: service}
}

// Startup stores the process context supplied by Wails.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
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
