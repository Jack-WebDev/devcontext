package editor

// Executable identifies the editor command to start.
type Executable string

// Arguments stores structured editor command arguments.
type Arguments []string

// Command is the editor-owned command specification used by launch planning.
//
// It intentionally excludes environment, working directory, and process
// detachment. Those are process-launcher concerns.
type Command struct {
	Executable Executable
	Arguments  Arguments
}

// ContextPaths contains context-owned storage locations an editor may use while
// building its command.
type ContextPaths struct {
	RootDir     string
	DataDir     string
	UserDataDir string
}

// CommandRequest contains the resolved inputs needed to build an editor
// command without starting a process.
type CommandRequest struct {
	Config      Config
	Executable  Executable
	ProjectPath string
	Paths       ContextPaths
}

// Editor is the contract implemented by editor integrations.
//
// Implementations detect their executable and build a structured command. They
// must not start processes.
type Editor interface {
	ID() ID
	DetectExecutable(Config) (Executable, error)
	BuildLaunchCommand(CommandRequest) (Command, error)
}
