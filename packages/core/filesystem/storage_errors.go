package filesystem

import (
	"errors"
	"fmt"
	"os"
)

// ErrStoragePermissionDenied identifies local Dev Context storage operations
// blocked by operating-system permissions.
var ErrStoragePermissionDenied = errors.New("storage permission denied")

// StoragePermissionError describes a denied storage operation without exposing
// file contents.
type StoragePermissionError struct {
	Operation string
	Path      string
	Err       error
}

func (e *StoragePermissionError) Error() string {
	switch {
	case e == nil:
		return ""
	case e.Operation != "" && e.Path != "" && e.Err != nil:
		return fmt.Sprintf("storage permission denied during %s at %q: %v", e.Operation, e.Path, e.Err)
	case e.Operation != "" && e.Path != "":
		return fmt.Sprintf("storage permission denied during %s at %q", e.Operation, e.Path)
	case e.Path != "":
		return fmt.Sprintf("storage permission denied at %q", e.Path)
	default:
		return ErrStoragePermissionDenied.Error()
	}
}

func (e *StoragePermissionError) Unwrap() []error {
	if e == nil || e.Err == nil {
		return []error{ErrStoragePermissionDenied}
	}
	return []error{ErrStoragePermissionDenied, e.Err}
}

// StorageOperation returns the denied storage operation.
func (e *StoragePermissionError) StorageOperation() string {
	if e == nil {
		return ""
	}
	return e.Operation
}

// StoragePath returns the affected local storage path.
func (e *StoragePermissionError) StoragePath() string {
	if e == nil {
		return ""
	}
	return e.Path
}

// WrapStoragePermissionError annotates permission-denied storage failures with
// the affected operation and path. Non-permission failures are returned as-is.
func WrapStoragePermissionError(operation string, path string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrStoragePermissionDenied) {
		return err
	}
	if os.IsPermission(err) || errors.Is(err, os.ErrPermission) {
		return &StoragePermissionError{
			Operation: operation,
			Path:      path,
			Err:       err,
		}
	}
	return err
}
