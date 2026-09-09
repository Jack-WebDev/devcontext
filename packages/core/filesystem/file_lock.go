package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
)

// WithExclusiveFileLock runs action while holding an exclusive lock associated
// with path. The lock is kept in a separate file so replacing path atomically
// does not release the lock held by another process.
func WithExclusiveFileLock(path string, action func() error) error {
	if action == nil {
		return fmt.Errorf("exclusive file lock: action is required")
	}

	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, RestrictedDirectoryMode); err != nil {
		return fmt.Errorf("create lock directory %q: %w", directory, err)
	}

	lockPath := path + ".lock"
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, RestrictedFileMode)
	if err != nil {
		return fmt.Errorf("open lock file %q: %w", lockPath, err)
	}
	defer file.Close()

	if err := lockFileExclusive(file); err != nil {
		return fmt.Errorf("lock file %q: %w", lockPath, err)
	}
	defer unlockFile(file)

	return action()
}
