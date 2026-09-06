package codingtool

// Executable identifies the coding-tool command to start.
type Executable string

// Arguments stores structured coding-tool command arguments.
type Arguments []string

// Command is the coding-tool-owned command specification used by launch planning.
//
// It intentionally excludes environment, working directory, and process
// detachment. Those are process-launcher concerns.
type Command struct {
	Executable Executable
	Arguments  Arguments
}

// ContextPaths contains the selected tool's storage plus a safe shared context
// root. Tool implementations must not depend on another tool's storage.
type ContextPaths struct {
	RootDir    string
	StorageDir string
}

// CommandRequest contains the resolved inputs needed to build a coding tool
// command without starting a process.
type CommandRequest struct {
	Config      Config
	Executable  Executable
	ProjectPath string
	Paths       ContextPaths
}

// CodingTool is the contract implemented by coding-tool integrations.
//
// Implementations detect their executable and build a structured command. They
// must not start processes.
type CodingTool interface {
	ID() ID
	DetectExecutable(Config) (Executable, error)
	BuildLaunchCommand(CommandRequest) (Command, error)
}

// ExecutableDetection describes how a tool executable was resolved. It is an
// optional extension because most tool adapters only need the base CodingTool
// contract to launch successfully.
type ExecutableDetection struct {
	Executable Executable
	Platform   string
	Source     ExecutableDetectionSource
}

// ExecutableDetectionSource is a bounded, presentation-safe explanation for
// an executable resolution result.
type ExecutableDetectionSource string

const (
	ExecutableDetectionPath        ExecutableDetectionSource = "path"
	ExecutableDetectionInstalled   ExecutableDetectionSource = "installed_application"
	ExecutableDetectionConfigured  ExecutableDetectionSource = "configured"
	ExecutableDetectionUnavailable ExecutableDetectionSource = "unavailable"
)

// DetailedExecutableDetector is implemented by tools that can explain how
// executable detection succeeded without exposing implementation-specific
// probing details to application callers.
type DetailedExecutableDetector interface {
	DetectExecutableDetailed(Config) (ExecutableDetection, error)
}

// WorkspaceCapabilities describes the actions a coding-tool integration can
// safely perform for an already launched workspace. These capabilities are
// intentionally separate from observed process and session state: an adapter
// may know that a workspace is active without being able to focus, reveal, or
// stop it.
type WorkspaceCapabilities struct {
	Focusable  bool
	Revealable bool
	Stoppable  bool
}

// WorkspaceTool is an optional extension for integrations that can operate on
// an existing workspace. Actions themselves remain adapter-owned; this
// contract only advertises which requests a future caller may make.
type WorkspaceTool interface {
	CodingTool
	WorkspaceCapabilities() WorkspaceCapabilities
}

// WorkspaceReference contains the safe, adapter-relevant identity of an
// existing workspace. It deliberately excludes credentials and launch args.
type WorkspaceReference struct {
	ID          string
	ProjectPath string
	SessionID   string
	ProcessID   *int
}

// WorkspaceTarget identifies one adapter-owned window eligible for focus.
type WorkspaceTarget struct {
	ID    string
	Label string
}

type WorkspaceRevealer interface {
	RevealTargets(WorkspaceReference) ([]WorkspaceTarget, error)
	RevealWorkspace(WorkspaceReference, string) error
}
type WorkspaceStopper interface {
	StopWorkspace(WorkspaceReference) error
}
