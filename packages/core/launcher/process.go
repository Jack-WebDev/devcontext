package launcher

import (
	"errors"
	"fmt"
	"os/exec"
	"sort"
)

var (
	// ErrMissingProcessExecutable identifies a process launch request without an
	// executable.
	ErrMissingProcessExecutable = errors.New("missing process executable")

	// ErrDetachedProcessUnsupported identifies detached launch requests before
	// platform-specific detach behavior is implemented.
	ErrDetachedProcessUnsupported = errors.New("detached process launch is not implemented")
)

// DetachMode describes whether Dev Context should wait on the launched editor
// process. Platform-specific detached behavior is implemented in a later phase.
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

// NativeProcessLauncher starts editor processes through the operating system.
type NativeProcessLauncher struct{}

var _ ProcessLauncher = NativeProcessLauncher{}

// Launch starts the requested process without invoking a shell.
func (NativeProcessLauncher) Launch(request ProcessRequest) error {
	if request.Executable == "" {
		return ErrMissingProcessExecutable
	}
	if request.DetachMode == DetachModeDetached {
		return ErrDetachedProcessUnsupported
	}

	command := exec.Command(string(request.Executable), request.Arguments.strings()...)
	command.Env = request.Environment.Environ()
	command.Dir = string(request.WorkingDirectory)

	if err := command.Run(); err != nil {
		return fmt.Errorf("launch process %q: %w", request.Executable, err)
	}
	return nil
}

func (a Arguments) strings() []string {
	values := make([]string, len(a))
	for i, value := range a {
		values[i] = string(value)
	}
	return values
}

// Environ returns deterministic KEY=value process environment entries.
func (e Environment) Environ() []string {
	keys := make([]string, 0, len(e))
	for key := range e {
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, key+"="+e[key])
	}
	return entries
}
