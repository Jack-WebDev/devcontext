package filesystem_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, contextPaths, ctx, []string{"codex", "claude"}); err != nil {
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

	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, contextPaths, ctx, []string{"codex", "claude"}); err != nil {
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

	if err := filesystem.ImportProviderCredentials(platformPaths, contextPaths, []string{"codex", "claude"}); err != nil {
		t.Fatalf("import provider credentials: %v", err)
	}

	assertFileBytes(t, filepath.Join(contextPaths.CodexDir, "auth.json"), []byte("isolated-codex"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, ".credentials.json"), []byte("isolated-claude"))
	assertFileBytes(t, filepath.Join(contextPaths.ClaudeDir, "settings.json"), []byte("isolated-settings"))
}

func TestCreateContextDirectoryTreeWithProviderCredentialsImportsOnlySelectedProviders(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("global-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("global-claude"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", "settings.json"), []byte("global-settings"))

	contextID := devcontext.MustID("personal")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := devcontext.DefaultPersonalContext(time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC))

	if err := filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, contextPaths, ctx, []string{"codex"}); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertFileBytes(t, filepath.Join(contextPaths.CodexDir, "auth.json"), []byte("global-codex"))
	assertDirectoryEmpty(t, contextPaths.ClaudeDir)
}

func TestDetectProviderCredentialSessionsReturnsOnlySafeMetadata(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	codexToken := testJWT(t, map[string]string{
		"email":               "user@company.com",
		"chatgpt_plan_type":   "business",
		"chatgpt_account_id":  "acct-123",
		"ignored_token_claim": "jwt-secret-claim",
	})
	writeJSONCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), map[string]string{
		"id_token":     codexToken,
		"access_token": "codex-access-token",
	})
	writeJSONCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), map[string]string{
		"subscriptionType": "Pro",
		"organizationUuid": "e783-organization",
		"organizationName": "Jishin Labs",
		"accessToken":      "claude-access-token",
		"refreshToken":     "claude-refresh-token",
	})

	sessions, err := filesystem.DetectProviderCredentialSessions(platformPaths)
	if err != nil {
		t.Fatalf("detect provider credential sessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("session count = %d, want 2: %#v", len(sessions), sessions)
	}

	codex := sessions[0]
	if codex.ProviderID != "codex" || !codex.MetadataAvailable {
		t.Fatalf("codex session = %#v, want available codex metadata", codex)
	}
	if codex.Codex.Email != "user@company.com" ||
		codex.Codex.ChatGPTPlanType != "business" ||
		codex.Codex.ChatGPTAccountID != "acct-123" {
		t.Fatalf("codex metadata = %#v", codex.Codex)
	}

	claude := sessions[1]
	if claude.ProviderID != "claude" || !claude.MetadataAvailable {
		t.Fatalf("claude session = %#v, want available claude metadata", claude)
	}
	if claude.Claude.SubscriptionType != "Pro" ||
		claude.Claude.OrganizationUUID != "e783-organization" ||
		claude.Claude.OrganizationName != "Jishin Labs" {
		t.Fatalf("claude metadata = %#v", claude.Claude)
	}

	rendered := fmt.Sprintf("%#v", sessions)
	for _, secret := range []string{
		codexToken,
		"codex-access-token",
		"jwt-secret-claim",
		"claude-access-token",
		"claude-refresh-token",
	} {
		if contains := strings.Contains(rendered, secret); contains {
			t.Fatalf("detected sessions exposed credential value %q: %#v", secret, sessions)
		}
	}
}

func TestDetectProviderCredentialSessionsReadsNestedClaudeOrganizationName(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	writeJSONCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), map[string]any{
		"account": map[string]any{
			"organization": map[string]string{
				"uuid":        "ignored-uuid-field",
				"displayName": "Acme Research",
			},
		},
		"subscriptionType":  "Team",
		"organization_uuid": "e783-organization",
		"accessToken":       "claude-access-token",
	})

	sessions, err := filesystem.DetectProviderCredentialSessions(platformPaths)
	if err != nil {
		t.Fatalf("detect provider credential sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("session count = %d, want 1: %#v", len(sessions), sessions)
	}

	claude := sessions[0]
	if claude.ProviderID != "claude" || !claude.MetadataAvailable {
		t.Fatalf("claude session = %#v, want available claude metadata", claude)
	}
	if claude.Claude.SubscriptionType != "Team" ||
		claude.Claude.OrganizationUUID != "e783-organization" ||
		claude.Claude.OrganizationName != "Acme Research" {
		t.Fatalf("claude metadata = %#v", claude.Claude)
	}

	rendered := fmt.Sprintf("%#v", sessions)
	if strings.Contains(rendered, "claude-access-token") {
		t.Fatalf("detected sessions exposed credential value: %#v", sessions)
	}
}

func TestDetectProviderCredentialSessionsReadsNestedCodexIDToken(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	codexToken := testJWT(t, map[string]string{
		"email":              "work@example.com",
		"chatgpt_plan_type":  "team",
		"chatgpt_account_id": "acct-work",
	})
	writeJSONCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), map[string]any{
		"tokens": map[string]string{
			"id_token":     codexToken,
			"access_token": "codex-access-token",
		},
	})

	sessions, err := filesystem.DetectProviderCredentialSessions(platformPaths)
	if err != nil {
		t.Fatalf("detect provider credential sessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("session count = %d, want 1: %#v", len(sessions), sessions)
	}
	codex := sessions[0]
	if codex.ProviderID != "codex" || !codex.MetadataAvailable {
		t.Fatalf("codex session = %#v, want available codex metadata", codex)
	}
	if codex.Codex.Email != "work@example.com" ||
		codex.Codex.ChatGPTPlanType != "team" ||
		codex.Codex.ChatGPTAccountID != "acct-work" {
		t.Fatalf("codex metadata = %#v", codex.Codex)
	}

	rendered := fmt.Sprintf("%#v", sessions)
	for _, secret := range []string{codexToken, "codex-access-token"} {
		if strings.Contains(rendered, secret) {
			t.Fatalf("detected sessions exposed credential value %q: %#v", secret, sessions)
		}
	}
}

func TestCreateContextDirectoryTreeWithProviderCredentialsKeepsContextsIsolated(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	createdAt := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("personal-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("personal-claude"))
	personal := devcontext.DefaultPersonalContext(createdAt)
	personalPaths, err := filesystem.DeriveContextPaths(platformPaths, personal.ID)
	if err != nil {
		t.Fatalf("derive personal paths: %v", err)
	}
	err = filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, personalPaths, personal, []string{"codex", "claude"})
	if err != nil {
		t.Fatalf("create personal context: %v", err)
	}

	writeCredentialFixture(t, filepath.Join(homeDir, ".codex", "auth.json"), []byte("company-codex"))
	writeCredentialFixture(t, filepath.Join(homeDir, ".claude", ".credentials.json"), []byte("company-claude"))
	company := devcontext.DefaultCompanyContext(createdAt)
	companyPaths, err := filesystem.DeriveContextPaths(platformPaths, company.ID)
	if err != nil {
		t.Fatalf("derive company paths: %v", err)
	}
	err = filesystem.CreateContextDirectoryTreeWithProviderCredentials(platformPaths, companyPaths, company, []string{"codex", "claude"})
	if err != nil {
		t.Fatalf("create company context: %v", err)
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

func writeJSONCredentialFixture(t *testing.T, path string, value any) {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal credential fixture: %v", err)
	}
	writeCredentialFixture(t, path, data)
}

func testJWT(t *testing.T, claims map[string]string) string {
	t.Helper()

	header, err := json.Marshal(map[string]string{"alg": "none"})
	if err != nil {
		t.Fatalf("marshal jwt header: %v", err)
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal jwt payload: %v", err)
	}
	return strings.Join([]string{
		base64.RawURLEncoding.EncodeToString(header),
		base64.RawURLEncoding.EncodeToString(payload),
		"signature",
	}, ".")
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
