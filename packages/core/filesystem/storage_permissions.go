package filesystem

import (
	"fmt"
	"os"
	"runtime"
)

const (
	// RestrictedDirectoryMode permits only the current user to access storage
	// directories.
	RestrictedDirectoryMode os.FileMode = 0o700

	// RestrictedFileMode permits only the current user to read and write storage
	// files.
	RestrictedFileMode os.FileMode = 0o600
)

// ChmodFunc changes permissions for one filesystem path.
type ChmodFunc func(string, os.FileMode) error

// StoragePermissions applies platform-specific storage permission policy.
type StoragePermissions interface {
	DirectoryMode() os.FileMode
	FileMode() os.FileMode
	ApplyDirectory(path string) error
	ApplyFile(path string) error
}

// DefaultStoragePermissions applies owner-only permissions where supported.
type DefaultStoragePermissions struct {
	supported bool
	chmod     ChmodFunc
}

var _ StoragePermissions = (*DefaultStoragePermissions)(nil)

// NewDefaultStoragePermissions creates the default storage permission policy.
func NewDefaultStoragePermissions() *DefaultStoragePermissions {
	return NewStoragePermissions(runtime.GOOS != "windows", os.Chmod)
}

// NewStoragePermissions creates a storage permission policy with injected
// platform behavior for tests.
func NewStoragePermissions(supported bool, chmod ChmodFunc) *DefaultStoragePermissions {
	if chmod == nil {
		chmod = os.Chmod
	}
	return &DefaultStoragePermissions{
		supported: supported,
		chmod:     chmod,
	}
}

// DirectoryMode returns the mode to use when creating storage directories.
func (p *DefaultStoragePermissions) DirectoryMode() os.FileMode {
	return RestrictedDirectoryMode
}

// FileMode returns the mode to use when creating storage files.
func (p *DefaultStoragePermissions) FileMode() os.FileMode {
	return RestrictedFileMode
}

// ApplyDirectory applies storage directory permissions where supported.
func (p *DefaultStoragePermissions) ApplyDirectory(path string) error {
	return p.apply(path, RestrictedDirectoryMode, "directory")
}

// ApplyFile applies storage file permissions where supported.
func (p *DefaultStoragePermissions) ApplyFile(path string) error {
	return p.apply(path, RestrictedFileMode, "file")
}

func (p *DefaultStoragePermissions) apply(path string, mode os.FileMode, kind string) error {
	if !p.supported {
		return nil
	}
	if err := p.chmod(path, mode); err != nil {
		if wrapped := WrapStoragePermissionError("set "+kind+" permissions", path, err); wrapped != err {
			return wrapped
		}
		return fmt.Errorf("apply storage permissions to %s %q: %w", kind, path, err)
	}
	return nil
}
