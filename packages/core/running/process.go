package running

import (
	"errors"
	"fmt"
	"syscall"
)

// ProcessInspector reports whether an operating-system process is still running.
type ProcessInspector interface {
	IsRunning(pid int) (bool, error)
}

// NativeProcessInspector checks processes through the local operating system.
type NativeProcessInspector struct{}

// IsRunning checks a process without changing it. Permission-denied responses
// still prove that the process exists, so they are treated as running.
func (NativeProcessInspector) IsRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("inspect process: invalid PID %d", pid)
	}
	err := syscall.Kill(pid, 0)
	switch {
	case err == nil, errors.Is(err, syscall.EPERM):
		return true, nil
	case errors.Is(err, syscall.ESRCH):
		return false, nil
	default:
		return false, fmt.Errorf("inspect process %d: %w", pid, err)
	}
}
