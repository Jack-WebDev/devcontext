package running

// ProcessInspector reports whether an operating-system process is still running.
type ProcessInspector interface {
	IsRunning(pid int) (bool, error)
}

// NativeProcessInspector checks processes through the local operating system.
type NativeProcessInspector struct{}
