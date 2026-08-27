package launcher

import (
	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

// Executable identifies the editor executable to start.
type Executable string

// Arguments stores structured process arguments.
type Arguments []string

// Environment stores the process environment by variable name.
type Environment map[string]string

// WorkingDirectory identifies the directory used to start the editor process.
type WorkingDirectory string

// LaunchPlan represents the deterministic operation needed to launch an codingtool.
type LaunchPlan struct {
	ProjectPath        project.Path
	Context            devcontext.Context
	Tool               codingtool.Config
	Executable         Executable
	Arguments          Arguments
	WorkingDirectory   WorkingDirectory
	Environment        Environment
	ContextPaths       filesystem.ContextPaths
	Warnings           []ResolutionWarning
	ResolutionSource   ResolutionSource
	MissingProviderIDs []provider.ID
}
