package launcher

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
)

var (
	// ErrMissingProcessExecutable identifies a process launch request without an
	// executable.
	ErrMissingProcessExecutable = errors.New("missing process executable")

	// ErrProcessExecutableNotFound identifies a launch request whose executable
	// cannot be found by the operating system.
	ErrProcessExecutableNotFound = errors.New("process executable not found")

	// ErrProcessPermissionDenied identifies a launch request blocked by
	// operating-system permissions.
	ErrProcessPermissionDenied = errors.New("process launch permission denied")

	// ErrProcessWorkingDirectoryInvalid identifies a launch request with a
	// missing or non-directory working directory.
	ErrProcessWorkingDirectoryInvalid = errors.New("process working directory invalid")

	// ErrProcessStartFailed identifies a process launch failure that does not
	// match a more specific category.
	ErrProcessStartFailed = errors.New("process start failed")
)

// DetachMode describes whether Dev Context should wait on the launched coding tool
// process. Platform-specific detached behavior is implemented in a later phase.
type DetachMode string

const (
	// DetachModeAttached keeps the launched process attached to Dev Context.
	DetachModeAttached DetachMode = "attached"

	// DetachModeDetached allows Dev Context to exit after the coding tool starts.
	DetachModeDetached DetachMode = "detached"
)

// ProcessRequest describes one native process launch without starting it.
type ProcessRequest struct {
	Tool             Tool
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

// ProcessLaunchError describes a categorized native process launch failure.
type ProcessLaunchError struct {
	Tool             Tool
	Executable       Executable
	WorkingDirectory WorkingDirectory
	Err              error
	Cause            error
}

func (e *ProcessLaunchError) Error() string {
	if e.WorkingDirectory != "" {
		return fmt.Sprintf("%v: executable %q in %q: %v", e.Err, e.Executable, e.WorkingDirectory, e.Cause)
	}
	return fmt.Sprintf("%v: executable %q: %v", e.Err, e.Executable, e.Cause)
}

func (e *ProcessLaunchError) Unwrap() []error {
	if e.Cause == nil {
		return []error{e.Err}
	}
	return []error{e.Err, e.Cause}
}

// NativeProcessLauncher starts coding-tool processes through the operating system.
type NativeProcessLauncher struct{}

var _ ProcessLauncher = NativeProcessLauncher{}

// Launch starts the requested process without invoking a shell.
func (NativeProcessLauncher) Launch(request ProcessRequest) error {
	if request.Executable == "" {
		return ErrMissingProcessExecutable
	}
	if err := validateProcessWorkingDirectory(request); err != nil {
		return err
	}

	command := exec.Command(string(request.Executable), request.Arguments.strings()...)
	command.Env = request.Environment.Environ()
	command.Dir = string(request.WorkingDirectory)

	if request.DetachMode == DetachModeDetached {
		configureDetachedCommand(command)
		if err := command.Start(); err != nil {
			return mapProcessLaunchError(request, err)
		}
		if err := command.Process.Release(); err != nil {
			return newProcessLaunchError(request, ErrProcessStartFailed, err)
		}
		return nil
	}

	if err := command.Run(); err != nil {
		return mapProcessLaunchError(request, err)
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

func validateProcessWorkingDirectory(request ProcessRequest) error {
	if request.WorkingDirectory == "" {
		return nil
	}

	info, err := os.Stat(string(request.WorkingDirectory))
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return newProcessLaunchError(request, ErrProcessPermissionDenied, err)
		}
		return newProcessLaunchError(request, ErrProcessWorkingDirectoryInvalid, err)
	}
	if !info.IsDir() {
		return newProcessLaunchError(request, ErrProcessWorkingDirectoryInvalid, fmt.Errorf("%q is not a directory", request.WorkingDirectory))
	}
	return nil
}

func mapProcessLaunchError(request ProcessRequest, err error) error {
	switch {
	case errors.Is(err, os.ErrPermission):
		return newProcessLaunchError(request, ErrProcessPermissionDenied, err)
	case errors.Is(err, exec.ErrNotFound), errors.Is(err, os.ErrNotExist):
		return newProcessLaunchError(request, ErrProcessExecutableNotFound, err)
	default:
		return newProcessLaunchError(request, ErrProcessStartFailed, err)
	}
}

func newProcessLaunchError(request ProcessRequest, category error, cause error) error {
	return &ProcessLaunchError{
		Tool:             request.Tool,
		Executable:       request.Executable,
		WorkingDirectory: request.WorkingDirectory,
		Err:              category,
		Cause:            cause,
	}
}
