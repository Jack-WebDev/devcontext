package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
)

func TestCreateContextDirectoryTreeWithProviderCredentialsImportsGlobalFiles(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("codex-auth-fixture"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("claude-credentials-fixture"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("claude-settings-fixture"))

	contextID := devcontext.MustID("personal")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := devcontext.DefaultPersonalContext(time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC))

	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertFileBytes(t, filepath.Join(contextPaths.CodexDir, "auth.json"), []byte("codex-auth-fixture"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, ".credentials.json"), []byte("claude-credentials-fixture"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, "settings.json"), []byte("claude-settings-fixture"))
	assertRestrictedMode(t, filepath.Join(contextPaths.CodexDir, "auth.json"), filesystem.RestrictedFileMode)
	assertRestrictedMode(t, filepath.Join(contextPaths.ClaudeDir, ".credentials.json"), filesystem.RestrictedFileMode)
	assertRestrictedMode(t, filepath.Join(contextPaths.ClaudeDir, "settings.json"), filesystem.RestrictedFileMode)
}

func TestImportProviderCredentialsLeavesProvidersEmptyWhenGlobalCredentialsAreMissing(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("settings-without-credentials"))

	contextID := devcontext.MustID("personal")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := devcontext.DefaultPersonalContext(time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC))

	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertDirectoryEmpty(t, contextPaths.CodexDir)
	assertDirectoryEmpty(t, contextPaths.ClaudeDir)
}

func TestImportProviderCredentialsDoesNotOverwriteExistingIsolatedFiles(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("global-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("global-claude"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("global-settings"))

	contextPaths := createEmptyContextProviderDirs(t, platformPaths, "personal")
	writeCredentialFixture(t, filepath.Join(contextPaths.CodexDir, "auth.json"), []byte("isolated-codex"))
	writeCredentialFixture(t, filepath.Join(contextPaths.ClaudeDir, ".credentials.json"), []byte("isolated-claude"))
	writeCredentialFixture(t, filepath.Join(contextPaths.ClaudeDir, "settings.json"), []byte("isolated-settings"))

	if err := filesystem.ImportProviderCredentials(platformPaths, contextPaths); err != nil {
		t.Fatalf("import provider credentials: %v", err)
	}

	assertFileBytes(t, filepath.Join(contextPaths.CodexDir, "auth.json"), []byte("isolated-codex"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, ".credentials.json"), []byte("isolated-claude"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, "settings.json"), []byte("isolated-settings"))
}

func TestCreateContextDirectoryTreeWithProviderCredentialsKeepsContextsIsolated(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	createdAt := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("personal-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("personal-claude"))
	personal, err := filesystem.BootstrapPersonalContext(platformPaths, createdAt)
	if err != nil {
		t.Fatalf("bootstrap personal context: %v", err)
	}

	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("company-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("company-claude"))
	company, err := filesystem.BootstrapCompanyContext(platformPaths, createdAt)
	if err != nil {
		t.Fatalf("bootstrap company context: %v", err)
	}

	personalPaths, err := filesystem.DeriveContextPaths(platformPaths, personal.ID)
	if err != nil {
		t.Fatalf("derive personal paths: %v", err)
	}
	companyPaths, err := filesystem.DeriveContextPaths(platformPaths, company.ID)
	if err != nil {
		t.Fatalf("derive company paths: %v", err)
	}
	if personalPaths.CodexDir == companyPaths.CodexDir || personalPaths.ClaudeDir == companyPaths.ClaudeDir {
		t.Fatalf("provider directories are shared: personal=%#v company=%#v", personalPaths, companyPaths)
	}

	assertFileBytes(t, filepath.Join(personalPaths.CodexDir, "auth.json"), []byte("personal-codex"))
	assertFileBytes(t, filepath.Join(personalPaths.ClaudeDir, ".credentials.json"), []byte("personal-claude"))
	assertFileBytes(t, filepath.Join(companyPaths.CodexDir, "auth.json"), []byte("company-codex"))
	assertFileBytes(t, filepath.Join(companyPaths.ClaudeDir, ".credentials.json"), []byte("company-claude"))
}

func createEmptyContextProviderDirs(t *testing.T, platformPaths filesystem.PlatformPaths, contextID string) filesystem.ContextPaths {
	t.Helper()

	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, devcontext.MustID(contextID))
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	for _, dir := range []string{contextPaths.ClaudeDir, contextPaths.CodexDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create provider directory %q: %v", dir, err)
		}
	}
	return contextPaths
}

func writeCredentialFixture(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create credential fixture directory %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write credential fixture %q: %v", path, err)
	}
}

func assertFileBytes(t *testing.T, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %q: %v", path, err)
	}
	if string(got) != string(want) {
		t.Fatalf("file %q = %q, want %q", path, string(got), string(want))
	}
}
