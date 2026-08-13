package context

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
			return Context{}, fmt.Errorf("%w: %s", ErrContextNotFound, contextID.String())
		}
		return Context{}, fmt.Errorf("%w: read %q: %w", ErrUnreadableContextConfig, path, err)
	}

	ctx, err := DecodeContextTOML(data, contextID)
	if err != nil {
		return Context{}, fmt.Errorf("%w: decode %q: %w", ErrUnreadableContextConfig, path, err)
	}
	return ctx, nil
}

func (r Repository) contextConfigPath(contextID ID) string {
	return filepath.Join(r.contextsDir, contextID.String(), contextConfigFileName)
}

type contextAtomicWriteFunc func(file *os.File) error

func writeContextFileAtomically(path string, write contextAtomicWriteFunc) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	file, err := os.CreateTemp(dir, "."+base+".tmp-*")
	if err != nil {
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
		return fmt.Errorf("write temporary file for %q: %w", path, err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync temporary file for %q: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close temporary file for %q: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
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
