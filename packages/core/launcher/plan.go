package launcher

import (
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/project"
)

// Executable identifies the editor executable to start.
type Executable string

// Arguments stores structured process arguments.
type Arguments []string

// Environment stores the process environment by variable name.
type Environment map[string]string

// WorkingDirectory identifies the directory used to start the editor process.
type WorkingDirectory string

// LaunchPlan represents the deterministic operation needed to launch an editor.
type LaunchPlan struct {
	ProjectPath      project.Path
	Context          devcontext.Context
	Editor           editor.Config
	Executable       Executable
	Arguments        Arguments
	WorkingDirectory WorkingDirectory
	Environment      Environment
	Warnings         []ResolutionWarning
	ResolutionSource ResolutionSource
}
