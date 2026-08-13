package launcher

// DetachMode describes whether Dev Context should wait on the launched editor
// process. Native process behavior is implemented in a later phase.
type DetachMode string

const (
	// DetachModeAttached keeps the launched process attached to Dev Context.
	DetachModeAttached DetachMode = "attached"

	// DetachModeDetached allows Dev Context to exit after the editor starts.
	DetachModeDetached DetachMode = "detached"
)

// ProcessRequest describes one native process launch without starting it.
type ProcessRequest struct {
	Executable       Executable
	Arguments        Arguments
	Environment      Environment
	WorkingDirectory WorkingDirectory
	DetachMode       DetachMode
}

// ProcessLauncher starts a native process from a structured request.
type ProcessLauncher interface {
	Launch(ProcessRequest) error
}
