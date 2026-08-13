package project

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	// ErrProjectDirectoryNotFound identifies a project path that does not exist.
	ErrProjectDirectoryNotFound = errors.New("project directory not found")

	// ErrProjectPathNotDirectory identifies a project path that exists as a
	// non-directory file.
	ErrProjectPathNotDirectory = errors.New("project path is not a directory")

	// ErrProjectDirectoryUnreadable identifies a project directory that cannot be
	// opened or listed.
	ErrProjectDirectoryUnreadable = errors.New("project directory is unreadable")
)

// ValidateProjectDirectory verifies that path exists and can be opened as a
// readable directory.
func ValidateProjectDirectory(path Path) error {
	if path == "" {
		return fmt.Errorf("%w: path cannot be empty", ErrInvalidProjectPath)
	}

	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrProjectDirectoryNotFound, path)
		}
		return fmt.Errorf("%w: inspect %s: %w", ErrProjectDirectoryUnreadable, path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrProjectPathNotDirectory, path)
	}

	dir, err := os.Open(string(path))
	if err != nil {
		return fmt.Errorf("%w: open %s: %w", ErrProjectDirectoryUnreadable, path, err)
	}
	defer dir.Close()

	if _, err := dir.Readdirnames(1); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: read %s: %w", ErrProjectDirectoryUnreadable, path, err)
	}

	return nil
}
