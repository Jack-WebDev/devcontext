package filesystem

import devcontext "devctx/packages/core/context"

const (
	contextsDirectoryName = "contexts"
	contextConfigFileName = "context.toml"
	claudeDirectoryName   = "claude"
	codexDirectoryName    = "codex"
	vscodeDirectoryName   = "vscode"
	vscodeUserDataDirName = "user-data"
)

// ContextPaths contains every context-owned storage path derived from the Dev
// Context home directory and context ID.
type ContextPaths struct {
	ContextID         devcontext.ID
	RootDir           string
	ConfigPath        string
	ClaudeDir         string
	CodexDir          string
	VSCodeDir         string
	VSCodeUserDataDir string
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
		ContextID:         contextID,
		RootDir:           rootDir,
		ConfigPath:        joinPlatformPath(rootDir, contextConfigFileName),
		ClaudeDir:         joinPlatformPath(rootDir, claudeDirectoryName),
		CodexDir:          joinPlatformPath(rootDir, codexDirectoryName),
		VSCodeDir:         vscodeDir,
		VSCodeUserDataDir: joinPlatformPath(vscodeDir, vscodeUserDataDirName),
	}, nil
}
