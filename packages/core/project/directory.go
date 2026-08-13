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
		return newPathError("", "validate", ErrInvalidProjectPath, fmt.Errorf("path cannot be empty"))
	}

	info, err := os.Stat(string(path))
	if err != nil {
		if os.IsNotExist(err) {
			return newPathError(string(path), "inspect", ErrProjectDirectoryNotFound, err)
		}
		return newPathError(string(path), "inspect", ErrProjectDirectoryUnreadable, err)
	}
	if !info.IsDir() {
		return newPathError(string(path), "validate", ErrProjectPathNotDirectory, fmt.Errorf("%q is not a directory", path))
	}

	dir, err := os.Open(string(path))
	if err != nil {
		return newPathError(string(path), "open", ErrProjectDirectoryUnreadable, err)
	}
	defer dir.Close()

	if _, err := dir.Readdirnames(1); err != nil && !errors.Is(err, io.EOF) {
		return newPathError(string(path), "read", ErrProjectDirectoryUnreadable, err)
	}

	return nil
}
