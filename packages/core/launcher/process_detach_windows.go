//go:build windows

package launcher

import (
	"os/exec"
	"syscall"
)

const (
	windowsDetachedProcess       = 0x00000008
	windowsCreateNewProcessGroup = 0x00000200
)

func configureDetachedCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: windowsDetachedProcess | windowsCreateNewProcessGroup,
	}
}
