package filesystem

import (
	"fmt"
	"os"
	"sort"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

// CreateContextDirectoryTree creates the isolated storage tree for one context
// and writes its context.toml.
func CreateContextDirectoryTree(paths ContextPaths, ctx devcontext.Context) error {
	return CreateContextDirectoryTreeWithPermissions(paths, ctx, NewDefaultStoragePermissions())
}

// CreateContextDirectoryTreeWithProviderRegistry creates the isolated storage
// tree for one context using registered provider definitions.
func CreateContextDirectoryTreeWithProviderRegistry(paths ContextPaths, ctx devcontext.Context, registry provider.Registry) error {
	return CreateContextDirectoryTreeWithProviderRegistryAndPermissions(paths, ctx, registry, NewDefaultStoragePermissions())
}

// CreateContextDirectoryTreeWithProviderRegistryAndPermissions creates the
// isolated storage tree for one context using registered provider definitions
// and the supplied storage permission policy.
func CreateContextDirectoryTreeWithProviderRegistryAndPermissions(paths ContextPaths, ctx devcontext.Context, registry provider.Registry, permissions StoragePermissions) error {
	return createContextDirectoryTree(paths, ctx, registry, permissions, nil)
}

// CreateContextDirectoryTreeWithProviderCredentials creates the isolated
// storage tree for one context and imports supported global provider credential
// files into it.
func CreateContextDirectoryTreeWithProviderCredentials(platformPaths PlatformPaths, paths ContextPaths, ctx devcontext.Context, providerIDs []string) error {
	return CreateContextDirectoryTreeWithProviderCredentialsAndPermissions(platformPaths, paths, ctx, providerIDs, NewDefaultStoragePermissions())
}

// CreateContextDirectoryTreeWithProviderCredentialsAndPermissions creates the
// isolated storage tree for one context, imports supported global provider
// credential files, and uses the supplied storage permission policy.
func CreateContextDirectoryTreeWithProviderCredentialsAndPermissions(platformPaths PlatformPaths, paths ContextPaths, ctx devcontext.Context, providerIDs []string, permissions StoragePermissions) error {
	return CreateContextDirectoryTreeWithProviderRegistryCredentialsAndPermissions(platformPaths, paths, ctx, provider.BuiltInRegistry(), providerIDs, permissions)
}

// CreateContextDirectoryTreeWithProviderRegistryCredentialsAndPermissions
// creates the isolated storage tree for one context using registered provider
// definitions, imports supported global provider credential files, and uses the
// supplied storage permission policy.
func CreateContextDirectoryTreeWithProviderRegistryCredentialsAndPermissions(platformPaths PlatformPaths, paths ContextPaths, ctx devcontext.Context, registry provider.Registry, providerIDs []string, permissions StoragePermissions) error {
	return createContextDirectoryTree(paths, ctx, registry, permissions, func(createdPaths ContextPaths) error {
		return ImportProviderCredentialsWithPermissions(platformPaths, createdPaths, providerIDs, permissions)
	})
}

// CreateContextDirectoryTreeWithPermissions creates the isolated storage tree
// for one context using the supplied storage permission policy.
func CreateContextDirectoryTreeWithPermissions(paths ContextPaths, ctx devcontext.Context, permissions StoragePermissions) error {
	return createContextDirectoryTree(paths, ctx, provider.BuiltInRegistry(), permissions, nil)
}

func createContextDirectoryTree(paths ContextPaths, ctx devcontext.Context, registry provider.Registry, permissions StoragePermissions, importProviderCredentials func(ContextPaths) error) error {
	if permissions == nil {
		permissions = NewDefaultStoragePermissions()
	}
	if registry.IsZero() {
		registry = provider.BuiltInRegistry()
	}
	paths = paths.WithProviderStorageDirs(registeredEnabledProviderIDs(ctx, registry))
	if err := validateContextTree(paths, ctx); err != nil {
		return err
	}
	if err := ensureContextRootDoesNotExist(paths.RootDir); err != nil {
		return err
	}

	for _, dir := range contextTreeDirectories(paths) {
		if err := os.MkdirAll(dir, permissions.DirectoryMode()); err != nil {
			if wrapped := WrapStoragePermissionError("create directory", dir, err); wrapped != err {
				return fmt.Errorf("create context directory %q: %w", dir, wrapped)
			}
			return fmt.Errorf("create context directory %q: %w", dir, err)
		}
		if err := permissions.ApplyDirectory(dir); err != nil {
			return err
		}
	}

	if importProviderCredentials != nil {
		if err := importProviderCredentials(paths); err != nil {
			return fmt.Errorf("import provider credentials for context %q: %w", ctx.ID.String(), err)
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
		if wrapped := WrapStoragePermissionError("inspect directory", path, err); wrapped != err {
			return fmt.Errorf("inspect context directory %q: %w", path, wrapped)
		}
		return fmt.Errorf("inspect context directory %q: %w", path, err)
	}
}

func bootstrapDefaultContext(paths PlatformPaths, ctx devcontext.Context, permissions StoragePermissions) (devcontext.Context, error) {
	contextPaths, err := DeriveContextPaths(paths, ctx.ID)
	if err != nil {
		return devcontext.Context{}, err
	}
	if err := CreateContextDirectoryTreeWithProviderRegistryAndPermissions(contextPaths, ctx, provider.BuiltInRegistry(), permissions); err != nil {
		return devcontext.Context{}, err
	}
	return ctx, nil
}

func contextTreeDirectories(paths ContextPaths) []string {
	dirs := []string{
		paths.RootDir,
		paths.VSCodeDir,
		paths.VSCodeUserDataDir,
	}
	for _, providerID := range sortedProviderStorageIDs(paths.ProviderStorageDirs) {
		dirs = append(dirs, paths.ProviderStorageDirs[providerID])
	}
	return dirs
}

func registeredEnabledProviderIDs(ctx devcontext.Context, registry provider.Registry) []provider.ID {
	ids := make([]provider.ID, 0, len(ctx.Providers))
	for _, integration := range registry.All() {
		providerID := integration.ID()
		if config, ok := ctx.Providers[providerID]; ok && config.Enabled {
			ids = append(ids, providerID)
		}
	}
	return ids
}

func sortedProviderStorageIDs(providerDirs map[provider.ID]string) []provider.ID {
	ids := make([]provider.ID, 0, len(providerDirs))
	for providerID := range providerDirs {
		ids = append(ids, providerID)
	}
	sort.Slice(ids, func(i int, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}
