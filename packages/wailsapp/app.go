package wailsapp

import (
	"context"

	"devctx/packages/application"
)

type service interface {
	Greet(name string) string
	GetLaunchState(application.GetLaunchStateRequest) (application.LaunchState, *application.Error)
	LaunchProject(application.LaunchProjectRequest) (application.LaunchProjectResult, *application.Error)
	BindProject(application.BindProjectRequest) (application.ProjectBindingState, *application.Error)
	UnbindProject(application.UnbindProjectRequest) (application.ProjectBindingState, *application.Error)
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

// Greet returns a greeting for the current starter UI.
func (a *App) Greet(name string) string {
	return a.service.Greet(name)
}

// GetLaunchState returns selector state for the current project.
func (a *App) GetLaunchState(request application.GetLaunchStateRequest) (application.LaunchState, *application.Error) {
	return a.service.GetLaunchState(request)
}

// LaunchProject opens the selected project and context.
func (a *App) LaunchProject(request application.LaunchProjectRequest) (application.LaunchProjectResult, *application.Error) {
	return a.service.LaunchProject(request)
}

// BindProject remembers a context for a project.
func (a *App) BindProject(request application.BindProjectRequest) (application.ProjectBindingState, *application.Error) {
	return a.service.BindProject(request)
}

// UnbindProject removes the remembered context for a project.
func (a *App) UnbindProject(request application.UnbindProjectRequest) (application.ProjectBindingState, *application.Error) {
	return a.service.UnbindProject(request)
}
