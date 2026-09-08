//go:build !windows

package filesystem

import (
	"os"

	"golang.org/x/sys/unix"
)

func lockFileExclusive(file *os.File) error {
	return unix.Flock(int(file.Fd()), unix.LOCK_EX)
}

func unlockFile(file *os.File) {
	_ = unix.Flock(int(file.Fd()), unix.LOCK_UN)
}
