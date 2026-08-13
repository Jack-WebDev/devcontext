package launcher

import (
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
)

// InvocationSource identifies where a launch request originated.
type InvocationSource string

const (
	// InvocationSourceCLI identifies requests created from CLI input.
	InvocationSourceCLI InvocationSource = "cli"

	// InvocationSourceGUI identifies requests created from GUI input.
	InvocationSourceGUI InvocationSource = "gui"
)

// LaunchRequest represents user input needed to begin context resolution.
//
// RequestedContext is nil when the user has not chosen or supplied a context
// yet. Resolution, validation of the project path, and process launch are
// handled by later phases.
type LaunchRequest struct {
	ProjectPath      project.Path
	RequestedContext *devcontext.ID
	Interactive      bool
	Source           InvocationSource
}
