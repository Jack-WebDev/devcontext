package filesystem

import (
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

const (
	contextsDirectoryName        = "contexts"
	contextConfigFileName        = "context.toml"
	providerStorageDirectoryName = "providers"
	vscodeDirectoryName          = "vscode"
	vscodeUserDataDirName        = "user-data"
)

// ContextPaths contains every context-owned storage path derived from the Dev
// Context home directory and context ID.
type ContextPaths struct {
	ContextID              devcontext.ID
	RootDir                string
	ConfigPath             string
	ProviderStorageRootDir string
	ProviderStorageDirs    map[provider.ID]string
	VSCodeDir              string
	VSCodeUserDataDir      string
}

// DeriveContextPaths derives all context-owned paths without reading the
// filesystem.
func DeriveContextPaths(paths PlatformPaths, contextID devcontext.ID) (ContextPaths, error) {
	homeDir, err := paths.DevContextHomeDir()
	if err != nil {
		return ContextPaths{}, err
	}

	contextsDir := joinPlatformPath(homeDir, contextsDirectoryName)
	rootDir := joinPlatformPath(contextsDir, contextID.String())
	vscodeDir := joinPlatformPath(rootDir, vscodeDirectoryName)

	return ContextPaths{
		ContextID:              contextID,
		RootDir:                rootDir,
		ConfigPath:             joinPlatformPath(rootDir, contextConfigFileName),
		ProviderStorageRootDir: joinPlatformPath(rootDir, providerStorageDirectoryName),
		ProviderStorageDirs:    map[provider.ID]string{},
		VSCodeDir:              vscodeDir,
		VSCodeUserDataDir:      joinPlatformPath(vscodeDir, vscodeUserDataDirName),
	}, nil
}

// ProviderStorageDir returns the context-owned storage directory for providerID.
func (p ContextPaths) ProviderStorageDir(providerID provider.ID) string {
	if p.ProviderStorageDirs != nil {
		if dir := p.ProviderStorageDirs[providerID]; dir != "" {
			return dir
		}
	}
	return joinPlatformPath(p.ProviderStorageRootDir, string(providerID))
}

// WithProviderStorageDirs returns paths with provider storage directories
// materialized for the supplied provider IDs.
func (p ContextPaths) WithProviderStorageDirs(providerIDs []provider.ID) ContextPaths {
	next := p
	next.ProviderStorageDirs = make(map[provider.ID]string, len(providerIDs))
	for _, providerID := range providerIDs {
		if providerID == "" {
			continue
		}
		next.ProviderStorageDirs[providerID] = p.ProviderStorageDir(providerID)
	}
	return next
}
