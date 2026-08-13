package context

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const contextConfigFileName = "context.toml"

var (
	// ErrContextNotFound identifies a requested context that is not stored.
	ErrContextNotFound = errors.New("context not found")

	// ErrUnreadableContextConfig identifies a stored context config that cannot
	// be read or decoded.
	ErrUnreadableContextConfig = errors.New("unreadable context configuration")

	// ErrContextAlreadyExists identifies a context creation conflict.
	ErrContextAlreadyExists = errors.New("context already exists")
)

// Repository stores contexts below a contexts directory.
type Repository struct {
	contextsDir string
}

// NewRepository creates a context repository rooted at contextsDir.
func NewRepository(contextsDir string) Repository {
	return Repository{contextsDir: contextsDir}
}

// WriteContextFile writes one context configuration through a same-directory
// temporary file and atomic rename.
func WriteContextFile(path string, ctx Context) error {
	data, err := EncodeContextTOML(ctx)
	if err != nil {
		return err
	}

	return writeContextFileAtomically(path, func(file *os.File) error {
		_, err := file.Write(data)
		return err
	})
}

// Write persists one context in this repository.
func (r Repository) Write(ctx Context) error {
	if err := validateContextForTOML(ctx); err != nil {
		return err
	}
	return WriteContextFile(r.contextConfigPath(ctx.ID), ctx)
}

// List returns all valid contexts in deterministic ID order.
func (r Repository) List() ([]Context, error) {
	entries, err := os.ReadDir(r.contextsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		if wrapped := wrapStoragePermissionError("list directory", r.contextsDir, err); wrapped != err {
			return nil, fmt.Errorf("list contexts in %q: %w", r.contextsDir, wrapped)
		}
		return nil, fmt.Errorf("list contexts in %q: %w", r.contextsDir, err)
	}

	contexts := make([]Context, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		contextID, err := NewID(entry.Name())
		if err != nil {
			continue
		}

		ctx, err := r.Get(contextID)
		if err != nil {
			if errors.Is(err, ErrContextNotFound) || errors.Is(err, ErrUnreadableContextConfig) {
				continue
			}
			return nil, err
		}
		contexts = append(contexts, ctx)
	}

	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].ID.String() < contexts[j].ID.String()
	})

	return contexts, nil
}

// Get returns the stored context identified by contextID.
func (r Repository) Get(contextID ID) (Context, error) {
	if contextID.String() == "" {
		return Context{}, fmt.Errorf("%w: cannot be empty", ErrInvalidID)
	}

	path := r.contextConfigPath(contextID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Context{}, r.missingContextError(contextID)
		}
		if wrapped := wrapStoragePermissionError("read file", path, err); wrapped != err {
			return Context{}, fmt.Errorf("%w: read %q: %w", ErrUnreadableContextConfig, path, wrapped)
		}
		return Context{}, fmt.Errorf("%w: read %q: %w", ErrUnreadableContextConfig, path, err)
	}

	ctx, err := DecodeContextTOML(data, contextID)
	if err != nil {
		return Context{}, fmt.Errorf("%w: decode %q: %w", ErrUnreadableContextConfig, path, err)
	}
	return ctx, nil
}

// MissingContextError describes a requested context that is not available and
// includes configured alternatives when they can be determined.
type MissingContextError struct {
	ContextID    ID
	AvailableIDs []ID
}

func (e *MissingContextError) Error() string {
	if e == nil {
		return ""
	}
	if len(e.AvailableIDs) == 0 {
		return fmt.Sprintf("%v: %s", ErrContextNotFound, e.ContextID.String())
	}
	available := make([]string, len(e.AvailableIDs))
	for i, id := range e.AvailableIDs {
		available[i] = id.String()
	}
	return fmt.Sprintf("%v: %s; available contexts: %s", ErrContextNotFound, e.ContextID.String(), strings.Join(available, ", "))
}

func (e *MissingContextError) Unwrap() error {
	return ErrContextNotFound
}

// StoragePermissionError describes a denied context storage operation without
// exposing file contents.
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
		return fmt.Sprintf("context storage permission denied during %s at %q: %v", e.Operation, e.Path, e.Err)
	case e.Operation != "" && e.Path != "":
		return fmt.Sprintf("context storage permission denied during %s at %q", e.Operation, e.Path)
	case e.Path != "":
		return fmt.Sprintf("context storage permission denied at %q", e.Path)
	default:
		return "context storage permission denied"
	}
}

func (e *StoragePermissionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
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

func (r Repository) contextConfigPath(contextID ID) string {
	return filepath.Join(r.contextsDir, contextID.String(), contextConfigFileName)
}

func (r Repository) missingContextError(contextID ID) error {
	availableIDs, err := r.availableContextIDs()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrContextNotFound, contextID.String())
	}
	filteredIDs := make([]ID, 0, len(availableIDs))
	for _, availableID := range availableIDs {
		if availableID == contextID {
			continue
		}
		filteredIDs = append(filteredIDs, availableID)
	}
	return &MissingContextError{
		ContextID:    contextID,
		AvailableIDs: filteredIDs,
	}
}

func (r Repository) availableContextIDs() ([]ID, error) {
	entries, err := os.ReadDir(r.contextsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	ids := make([]ID, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id, err := NewID(entry.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(r.contextConfigPath(id))
		if err != nil {
			continue
		}
		if _, err := DecodeContextTOML(data, id); err != nil {
			continue
		}
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
	return ids, nil
}

type contextAtomicWriteFunc func(file *os.File) error

func writeContextFileAtomically(path string, write contextAtomicWriteFunc) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	file, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
		if wrapped := wrapStoragePermissionError("create temporary file", dir, err); wrapped != err {
			return fmt.Errorf("create temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("create temporary file for %q: %w", path, err)
	}

	tempPath := file.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := write(file); err != nil {
		_ = file.Close()
		if wrapped := wrapStoragePermissionError("write temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("write temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("write temporary file for %q: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		if wrapped := wrapStoragePermissionError("sync temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("sync temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("sync temporary file for %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		if wrapped := wrapStoragePermissionError("close temporary file", tempPath, err); wrapped != err {
			return fmt.Errorf("close temporary file for %q: %w", path, wrapped)
		}
		return fmt.Errorf("close temporary file for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		if wrapped := wrapStoragePermissionError("replace file", path, err); wrapped != err {
			return fmt.Errorf("replace %q atomically: %w", path, wrapped)
		}
		return fmt.Errorf("replace %q atomically: %w", path, err)
	}

	removeTemp = false
	syncContextDirectory(dir)
	return nil
}

func syncContextDirectory(path string) {
	dir, err := os.Open(path)
	if err != nil {
		return
	}
	defer dir.Close()

	_ = dir.Sync()
}

func wrapStoragePermissionError(operation string, path string, err error) error {
	if err == nil {
		return nil
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
