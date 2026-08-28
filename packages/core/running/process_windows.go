//go:build windows

package running

import (
	"errors"
	"fmt"

	"golang.org/x/sys/windows"
)

const windowsStillActive = 259

// IsRunning checks the Windows process exit code without changing the process.
func (NativeProcessInspector) IsRunning(pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("inspect process: invalid PID %d", pid)
	}
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
		return true, nil
	}
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)
	var exitCode uint32
	if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil {
		return false, fmt.Errorf("inspect process %d: %w", pid, err)
	}
	return exitCode == windowsStillActive, nil
}
