package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
)

func TestCreateContextDirectoryTreeWithProviderCredentialsImportsSelectedProviderFiles(t *testing.T) {
	homeDir := t.TempDir()
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) { return homeDir, nil })
	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("codex-auth"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("claude-credentials"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("claude-settings"))

	ctx := devcontext.DefaultPersonalContext(time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC))
	contextPaths, err := filesystem.DeriveContextPaths(paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(paths, contextPaths, ctx, []string{"codex", "claude"}); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertFileBytes(t, filepath.Join(contextPaths.ProviderStorageDir("codex"), "auth.json"), []byte("codex-auth"))
	assertFileBytes(t, filepath.Join(contextPaths.ProviderStorageDir("claude"), ".credentials.json"), []byte("claude-credentials"))
	assertFileBytes(t, filepath.Join(contextPaths.ProviderStorageDir("claude"), "settings.json"), []byte("claude-settings"))
}

func TestCreateContextDirectoryTreeWithProviderCredentialsDoesNotImportUnselectedOrMissingCredentials(t *testing.T) {
	homeDir := t.TempDir()
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) { return homeDir, nil })
	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("codex-auth"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("settings-only"))

	ctx := devcontext.DefaultPersonalContext(time.Now())
	contextPaths, err := filesystem.DeriveContextPaths(paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(paths, contextPaths, ctx, []string{"claude"}); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}
	assertCredentialDirectoryEmpty(t, contextPaths.ProviderStorageDir("codex"))
	assertCredentialDirectoryEmpty(t, contextPaths.ProviderStorageDir("claude"))
}

func TestGenericProviderCredentialHelpersDoNotOverwrite(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source", "credentials.json")
	destination := filepath.Join(root, "destination", "credentials.json")
	writeCredentialFixture(t, source, []byte("original"))
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		t.Fatalf("create destination directory: %v", err)
	}
	operations := filesystem.NewProviderCredentialFileOperations(filesystem.NewDefaultStoragePermissions())
	if err := operations.CopyOpaqueFile(source, destination); err != nil {
		t.Fatalf("copy credential: %v", err)
	}
	writeCredentialFixture(t, source, []byte("updated"))
	if err := operations.CopyOpaqueFile(source, destination); err != nil {
		t.Fatalf("copy existing credential: %v", err)
	}
	assertFileBytes(t, destination, []byte("original"))
}

func writeCredentialFixture(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create credential fixture directory: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write credential fixture: %v", err)
	}
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("file %q = %q, want %q", path, got, want)
	}
}

func assertCredentialDirectoryEmpty(t *testing.T, path string) {
	t.Helper()
	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("read directory %q: %v", path, err)
	}
	if len(entries) != 0 {
		t.Fatalf("directory %q entries = %#v, want empty", path, entries)
	}
}
