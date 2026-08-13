package filesystem_test

import (
	"errors"
	"os"
	"reflect"
	"runtime"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func TestCreateContextDirectoryTreeCreatesCompleteRestrictedTree(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")

	if err := filesystem.CreateContextDirectoryTree(contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	for _, dir := range []string{
		contextPaths.RootDir,
		contextPaths.ClaudeDir,
		contextPaths.CodexDir,
		contextPaths.VSCodeDir,
		contextPaths.VSCodeUserDataDir,
	} {
		assertDirectoryExists(t, dir)
		assertRestrictedMode(t, dir, filesystem.RestrictedDirectoryMode)
	}

	data, err := os.ReadFile(contextPaths.ConfigPath)
	if err != nil {
		t.Fatalf("read context config: %v", err)
	}
	decoded, err := devcontext.DecodeContextTOML(data, contextID)
	if err != nil {
		t.Fatalf("decode context config: %v", err)
	}
	if !reflect.DeepEqual(decoded, ctx) {
		t.Fatalf("decoded context = %#v, want %#v", decoded, ctx)
	}
	assertRestrictedMode(t, contextPaths.ConfigPath, filesystem.RestrictedFileMode)
	assertDirectoryEmpty(t, contextPaths.ClaudeDir)
	assertDirectoryEmpty(t, contextPaths.CodexDir)
	assertDirectoryEmpty(t, contextPaths.VSCodeUserDataDir)
}

func TestCreateContextDirectoryTreeRejectsMismatchedContextID(t *testing.T) {
	contextPaths := filesystem.ContextPaths{
		ContextID:         devcontext.MustID("company"),
		RootDir:           "/devctx/contexts/company",
		ConfigPath:        "/devctx/contexts/company/context.toml",
		ClaudeDir:         "/devctx/contexts/company/claude",
		CodexDir:          "/devctx/contexts/company/codex",
		VSCodeDir:         "/devctx/contexts/company/vscode",
		VSCodeUserDataDir: "/devctx/contexts/company/vscode/user-data",
	}
	ctx := contextTreeContext(devcontext.MustID("personal"), "Personal")

	err := filesystem.CreateContextDirectoryTreeWithPermissions(contextPaths, ctx, filesystem.NewStoragePermissions(false, nil))
	if !errors.Is(err, devcontext.ErrContextIDMismatch) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextIDMismatch)
	}
}

func assertDirectoryExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat directory %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

func assertRestrictedMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}
	assertMode(t, path, want)
}

func assertDirectoryEmpty(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("read directory %s: %v", path, err)
	}
	if len(entries) != 0 {
		t.Fatalf("directory %s entry count = %d, want 0", path, len(entries))
	}
}

func contextTreeContext(id devcontext.ID, name string) devcontext.Context {
	return devcontext.Context{
		ID:     id,
		Name:   name,
		Editor: editor.DefaultConfig(),
		Providers: provider.Configs{
			"claude": {Enabled: true},
			"codex":  {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"kind": "test",
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}
