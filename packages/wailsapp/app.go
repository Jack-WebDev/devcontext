package wailsapp

import (
	"context"
)

type service interface {
	Greet(name string) string
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
