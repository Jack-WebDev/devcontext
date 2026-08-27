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

// ContextPaths contains context-owned storage locations a coding tool may use while
// building its command.
type ContextPaths struct {
	RootDir     string
	DataDir     string
	UserDataDir string
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
