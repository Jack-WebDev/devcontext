package filesystem_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

func TestCreateContextDirectoryTreeRejectsDuplicateWithoutModifyingTree(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	original := contextTreeContext(contextID, "Client A")
	replacement := contextTreeContext(contextID, "Replacement")

	if err := filesystem.CreateContextDirectoryTree(contextPaths, original); err != nil {
		t.Fatalf("create original context directory tree: %v", err)
	}
	sentinelPath := filepath.Join(contextPaths.ClaudeDir, "auth-state.json")
	if err := os.WriteFile(sentinelPath, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("write sentinel file: %v", err)
	}
	before := snapshotTree(t, contextPaths.RootDir)

	err = filesystem.CreateContextDirectoryTree(contextPaths, replacement)
	if !errors.Is(err, devcontext.ErrContextAlreadyExists) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextAlreadyExists)
	}

	after := snapshotTree(t, contextPaths.RootDir)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("tree after duplicate create = %#v, want %#v", after, before)
	}

	data, err := os.ReadFile(contextPaths.ConfigPath)
	if err != nil {
		t.Fatalf("read context config: %v", err)
	}
	decoded, err := devcontext.DecodeContextTOML(data, contextID)
	if err != nil {
		t.Fatalf("decode context config: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("decoded context = %#v, want %#v", decoded, original)
	}
}

func TestValidateContextDirectoryTreeReportsIncompleteStorage(t *testing.T) {
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
	if err := os.RemoveAll(contextPaths.CodexDir); err != nil {
		t.Fatalf("remove codex dir: %v", err)
	}
	if err := os.RemoveAll(contextPaths.VSCodeDir); err != nil {
		t.Fatalf("remove vscode dir: %v", err)
	}
	if err := os.WriteFile(contextPaths.VSCodeDir, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write vscode file: %v", err)
	}

	err = filesystem.ValidateContextDirectoryTree(contextPaths)
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	var storageErr *filesystem.ContextStorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("error = %T, want *filesystem.ContextStorageError", err)
	}
	wantMissing := []filesystem.MissingContextDirectory{
		{
			Kind:   filesystem.ContextDirectoryCodex,
			Path:   contextPaths.CodexDir,
			Reason: "missing",
		},
		{
			Kind:   filesystem.ContextDirectoryVSCode,
			Path:   contextPaths.VSCodeDir,
			Reason: "not a directory",
		},
		{
			Kind:   filesystem.ContextDirectoryVSCodeUserData,
			Path:   contextPaths.VSCodeUserDataDir,
			Reason: "missing",
		},
	}
	if !reflect.DeepEqual(storageErr.Missing, wantMissing) {
		t.Fatalf("missing directories = %#v, want %#v", storageErr.Missing, wantMissing)
	}
}

func TestBootstrapDefaultContextsCreateLoadableSeeds(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	devContextHomeDir, err := platformPaths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	createdAt := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		bootstrap func(filesystem.PlatformPaths, time.Time) (devcontext.Context, error)
		want      devcontext.Context
	}{
		{
			name:      "personal",
			bootstrap: filesystem.BootstrapPersonalContext,
			want:      devcontext.DefaultPersonalContext(createdAt),
		},
		{
			name:      "company",
			bootstrap: filesystem.BootstrapCompanyContext,
			want:      devcontext.DefaultCompanyContext(createdAt),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seeded, err := tt.bootstrap(platformPaths, createdAt)
			if err != nil {
				t.Fatalf("bootstrap context: %v", err)
			}
			if !reflect.DeepEqual(seeded, tt.want) {
				t.Fatalf("seeded context = %#v, want %#v", seeded, tt.want)
			}

			repository := devcontext.NewRepository(filepath.Join(devContextHomeDir, "contexts"))
			stored, err := repository.Get(tt.want.ID)
			if err != nil {
				t.Fatalf("get bootstrapped context: %v", err)
			}
			if !reflect.DeepEqual(stored, tt.want) {
				t.Fatalf("stored context = %#v, want %#v", stored, tt.want)
			}
		})
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

func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()

	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			snapshot[rel] = "dir"
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[rel] = "file:" + string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot tree %s: %v", root, err)
	}
	return snapshot
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
