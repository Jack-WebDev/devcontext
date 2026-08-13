package filesystem

import (
	"fmt"
	"os"
	"time"

	devcontext "devctx/packages/core/context"
)

// CreateContextDirectoryTree creates the isolated storage tree for one context
// and writes its context.toml.
func CreateContextDirectoryTree(paths ContextPaths, ctx devcontext.Context) error {
	return CreateContextDirectoryTreeWithPermissions(paths, ctx, NewDefaultStoragePermissions())
}

// CreateContextDirectoryTreeWithPermissions creates the isolated storage tree
// for one context using the supplied storage permission policy.
func CreateContextDirectoryTreeWithPermissions(paths ContextPaths, ctx devcontext.Context, permissions StoragePermissions) error {
	if permissions == nil {
		permissions = NewDefaultStoragePermissions()
	}
	if err := validateContextTree(paths, ctx); err != nil {
		return err
	}
	if err := ensureContextRootDoesNotExist(paths.RootDir); err != nil {
		return err
	}

	for _, dir := range contextTreeDirectories(paths) {
		if err := os.MkdirAll(dir, permissions.DirectoryMode()); err != nil {
			return fmt.Errorf("create context directory %q: %w", dir, err)
		}
		if err := permissions.ApplyDirectory(dir); err != nil {
			return err
		}
	}

	if err := devcontext.WriteContextFile(paths.ConfigPath, ctx); err != nil {
		return fmt.Errorf("write context configuration %q: %w", paths.ConfigPath, err)
	}
	if err := permissions.ApplyFile(paths.ConfigPath); err != nil {
		return err
	}

	return nil
}

// BootstrapPersonalContext creates the built-in Personal context.
func BootstrapPersonalContext(paths PlatformPaths, createdAt time.Time) (devcontext.Context, error) {
	return BootstrapPersonalContextWithPermissions(paths, createdAt, NewDefaultStoragePermissions())
}

// BootstrapPersonalContextWithPermissions creates the built-in Personal context
// using the supplied storage permission policy.
func BootstrapPersonalContextWithPermissions(paths PlatformPaths, createdAt time.Time, permissions StoragePermissions) (devcontext.Context, error) {
	return bootstrapDefaultContext(paths, devcontext.DefaultPersonalContext(createdAt), permissions)
}

// BootstrapCompanyContext creates the built-in Company context.
func BootstrapCompanyContext(paths PlatformPaths, createdAt time.Time) (devcontext.Context, error) {
	return BootstrapCompanyContextWithPermissions(paths, createdAt, NewDefaultStoragePermissions())
}

// BootstrapCompanyContextWithPermissions creates the built-in Company context
// using the supplied storage permission policy.
func BootstrapCompanyContextWithPermissions(paths PlatformPaths, createdAt time.Time, permissions StoragePermissions) (devcontext.Context, error) {
	return bootstrapDefaultContext(paths, devcontext.DefaultCompanyContext(createdAt), permissions)
}

func validateContextTree(paths ContextPaths, ctx devcontext.Context) error {
	if paths.ContextID.String() == "" {
		return fmt.Errorf("%w: cannot be empty", devcontext.ErrInvalidID)
	}
	if ctx.ID != paths.ContextID {
		return fmt.Errorf("%w: context id %q, path id %q", devcontext.ErrContextIDMismatch, ctx.ID.String(), paths.ContextID.String())
	}
	for _, path := range append(contextTreeDirectories(paths), paths.ConfigPath) {
		if path == "" {
			return fmt.Errorf("context storage path cannot be empty")
		}
	}
	return nil
}

func ensureContextRootDoesNotExist(path string) error {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return fmt.Errorf("%w: %q", devcontext.ErrContextAlreadyExists, path)
	case os.IsNotExist(err):
		return nil
	default:
		return fmt.Errorf("inspect context directory %q: %w", path, err)
	}
}

func bootstrapDefaultContext(paths PlatformPaths, ctx devcontext.Context, permissions StoragePermissions) (devcontext.Context, error) {
	contextPaths, err := DeriveContextPaths(paths, ctx.ID)
	if err != nil {
		return devcontext.Context{}, err
	}
	if err := CreateContextDirectoryTreeWithPermissions(contextPaths, ctx, permissions); err != nil {
		return devcontext.Context{}, err
	}
	return ctx, nil
}

func contextTreeDirectories(paths ContextPaths) []string {
	return []string{
		paths.RootDir,
		paths.ClaudeDir,
		paths.CodexDir,
		paths.VSCodeDir,
		paths.VSCodeUserDataDir,
	}
}
