package editor

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

var (
	// ErrExecutableNotFound identifies an editor executable that could not be
	// found through the user's executable search path.
	ErrExecutableNotFound = errors.New("editor executable not found")
)

// ExecutableProbe finds executables through the platform search path.
type ExecutableProbe interface {
	LookPath(file string) (string, error)
}

type defaultExecutableProbe struct{}

func (defaultExecutableProbe) LookPath(file string) (string, error) {
	return exec.LookPath(file)
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

// VSCodeEditor detects the Visual Studio Code CLI.
type VSCodeEditor struct {
	Probe ExecutableProbe

	// OperatingSystem defaults to runtime.GOOS. Tests can set it to exercise
	// platform-specific command candidates without depending on the host OS.
	OperatingSystem string
}

// ID returns the persisted editor identifier.
func (VSCodeEditor) ID() ID {
	return VSCodeID
}

// DetectExecutable locates the VS Code command available through PATH.
func (e VSCodeEditor) DetectExecutable(Config) (Executable, error) {
	probe := e.resolveProbe()
	candidates := vscodeExecutableCandidates(e.resolveOperatingSystem())

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

func (e VSCodeEditor) resolveProbe() ExecutableProbe {
	if e.Probe != nil {
		return e.Probe
	}
	return defaultExecutableProbe{}
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
