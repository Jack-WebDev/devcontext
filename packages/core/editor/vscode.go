package editor

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

var (
	// ErrExecutableNotFound identifies an editor executable that could not be
	// found through the user's executable search path.
	ErrExecutableNotFound = errors.New("editor executable not found")

	// ErrExecutableNotExecutable identifies a configured editor executable path
	// that exists but cannot be used as an executable file.
	ErrExecutableNotExecutable = errors.New("editor executable is not executable")

	// ErrMissingUserDataDir identifies a VS Code command request without an
	// isolated user-data directory.
	ErrMissingUserDataDir = errors.New("missing VS Code user-data directory")

	// ErrMissingExecutable identifies a command request without a resolved
	// editor executable.
	ErrMissingExecutable = errors.New("missing editor executable")

	// ErrMissingProjectPath identifies a command request without a project path.
	ErrMissingProjectPath = errors.New("missing project path")
)

const (
	// VSCodeUserDataDirFlag identifies the VS Code CLI flag used to isolate
	// runtime and user state.
	VSCodeUserDataDirFlag = "--user-data-dir"
)

// ExecutableProbe finds executables and reads executable file metadata.
type ExecutableProbe interface {
	LookPath(file string) (string, error)
	Stat(path string) (os.FileInfo, error)
}

type defaultExecutableProbe struct{}

func (defaultExecutableProbe) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (defaultExecutableProbe) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

// ExecutableNotFoundError describes a failed editor executable lookup.
type ExecutableNotFoundError struct {
	EditorID   ID
	Candidates []string
}

func (e *ExecutableNotFoundError) Error() string {
	if len(e.Candidates) == 0 {
		return fmt.Sprintf("%s executable was not found", e.EditorID)
	}
	return fmt.Sprintf("%s executable was not found; checked: %s", e.EditorID, strings.Join(e.Candidates, ", "))
}

func (e *ExecutableNotFoundError) Unwrap() error {
	return ErrExecutableNotFound
}

// ExecutableNotExecutableError describes an unusable configured executable
// path.
type ExecutableNotExecutableError struct {
	EditorID ID
	Path     string
}

func (e *ExecutableNotExecutableError) Error() string {
	return fmt.Sprintf("%s executable %q is not executable", e.EditorID, e.Path)
}

func (e *ExecutableNotExecutableError) Unwrap() error {
	return ErrExecutableNotExecutable
}

// VSCodeEditor detects the Visual Studio Code CLI.
type VSCodeEditor struct {
	Probe ExecutableProbe

	// OperatingSystem defaults to runtime.GOOS. Tests can set it to exercise
	// platform-specific command candidates without depending on the host OS.
	OperatingSystem string
}

var _ Editor = VSCodeEditor{}

// ID returns the persisted editor identifier.
func (VSCodeEditor) ID() ID {
	return VSCodeID
}

// DetectExecutable locates the VS Code command available through PATH.
func (e VSCodeEditor) DetectExecutable(config Config) (Executable, error) {
	probe := e.resolveProbe()
	goos := e.resolveOperatingSystem()
	if override := strings.TrimSpace(config.ExecutableOverride); override != "" {
		return validateConfiguredExecutable(probe, goos, VSCodeID, override)
	}

	candidates := vscodeExecutableCandidates(goos)

	for _, candidate := range candidates {
		path, err := probe.LookPath(candidate)
		if err == nil {
			return Executable(path), nil
		}
	}

	return "", &ExecutableNotFoundError{
		EditorID:   VSCodeID,
		Candidates: candidates,
	}
}

// BuildLaunchCommand returns the structured VS Code command for one project and
// context-owned user-data directory.
func (VSCodeEditor) BuildLaunchCommand(request CommandRequest) (Command, error) {
	if strings.TrimSpace(string(request.Executable)) == "" {
		return Command{}, ErrMissingExecutable
	}
	if strings.TrimSpace(request.ProjectPath) == "" {
		return Command{}, ErrMissingProjectPath
	}

	arguments, err := VSCodeUserDataArguments(request.Paths)
	if err != nil {
		return Command{}, err
	}
	arguments = append(append(Arguments(nil), arguments...), request.ProjectPath)

	return Command{
		Executable: request.Executable,
		Arguments:  arguments,
	}, nil
}

func (e VSCodeEditor) resolveProbe() ExecutableProbe {
	if e.Probe != nil {
		return e.Probe
	}
	return defaultExecutableProbe{}
}

// VSCodeUserDataArguments returns the structured arguments that isolate VS Code
// user data for one context.
func VSCodeUserDataArguments(paths ContextPaths) (Arguments, error) {
	if strings.TrimSpace(paths.UserDataDir) == "" {
		return nil, ErrMissingUserDataDir
	}
	return Arguments{VSCodeUserDataDirFlag, paths.UserDataDir}, nil
}

func (e VSCodeEditor) resolveOperatingSystem() string {
	if e.OperatingSystem != "" {
		return e.OperatingSystem
	}
	return runtime.GOOS
}

func vscodeExecutableCandidates(goos string) []string {
	if goos == "windows" {
		return []string{"code", "code.cmd", "Code.exe"}
	}
	return []string{"code"}
}

func validateConfiguredExecutable(probe ExecutableProbe, goos string, editorID ID, path string) (Executable, error) {
	info, err := probe.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", &ExecutableNotFoundError{
				EditorID:   editorID,
				Candidates: []string{path},
			}
		}
		return "", fmt.Errorf("inspect configured %s executable %q: %w", editorID, path, err)
	}

	if !isUsableExecutable(info, goos) {
		return "", &ExecutableNotExecutableError{
			EditorID: editorID,
			Path:     path,
		}
	}
	return Executable(path), nil
}

func isUsableExecutable(info os.FileInfo, goos string) bool {
	if info == nil || info.IsDir() || !info.Mode().IsRegular() {
		return false
	}
	if goos == "windows" {
		return true
	}
	return info.Mode().Perm()&0o111 != 0
}
