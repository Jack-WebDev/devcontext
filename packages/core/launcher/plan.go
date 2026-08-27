package launcher

import (
	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

// Executable identifies the coding-tool executable to start.
type Executable string

// Arguments stores structured process arguments.
type Arguments []string

// Environment stores the process environment by variable name.
type Environment map[string]string

// WorkingDirectory identifies the directory used to start the coding-tool process.
type WorkingDirectory string

// Tool identifies the coding tool selected for a launch.
type Tool struct {
	ID          codingtool.ID
	DisplayName string
}

// LaunchPlan represents the deterministic operation needed to launch a coding tool.
type LaunchPlan struct {
	ProjectPath        project.Path
	Context            devcontext.Context
	Tool               Tool
	Executable         Executable
	Arguments          Arguments
	WorkingDirectory   WorkingDirectory
	Environment        Environment
	ContextPaths       filesystem.ContextPaths
	Warnings           []ResolutionWarning
	ResolutionSource   ResolutionSource
	MissingProviderIDs []provider.ID
}
