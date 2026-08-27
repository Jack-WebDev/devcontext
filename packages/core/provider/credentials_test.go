package provider

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCodexProviderCredentialCapabilities(t *testing.T) {
	homeDir := t.TempDir()
	token := testCredentialJWT(t, map[string]string{"email": "dev@example.com", "chatgpt_plan_type": "team", "chatgpt_account_id": "acct-1"})
	writeProviderCredential(t, filepath.Join(homeDir, ".codex", "auth.json"), map[string]any{"tokens": map[string]string{"id_token": token, "access_token": "secret"}})

	p := CodexProvider{}
	session, found, err := p.DetectGlobalCredentialSession(GlobalCredentialContext{UserHomeDir: homeDir})
	if err != nil || !found || !session.MetadataAvailable {
		t.Fatalf("session = %#v found=%t err=%v", session, found, err)
	}
	if got := metadataField(session.Fields, "Email"); got != "dev@example.com" {
		t.Fatalf("email = %q", got)
	}

	destination := filepath.Join(t.TempDir(), "providers", "codex")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	files := credentialFileOperations{}
	if err := p.ImportCredentials(CredentialImportContext{UserHomeDir: homeDir, Runtime: RuntimeContext{Paths: ContextPaths{StorageDir: destination}}, Files: files}); err != nil {
		t.Fatalf("import credentials: %v", err)
	}
	identity, available, err := p.DetectContextIdentity(RuntimeContext{Paths: ContextPaths{StorageDir: destination}})
	if err != nil || !available || metadataField(identity.Fields, "ChatGPT account ID") != "acct-1" {
		t.Fatalf("identity = %#v available=%t err=%v", identity, available, err)
	}
}

func TestClaudeProviderCredentialCapabilities(t *testing.T) {
	homeDir := t.TempDir()
	writeProviderCredential(t, filepath.Join(homeDir, ".claude", ".credentials.json"), map[string]string{"subscriptionType": "Pro", "organizationUuid": "org-1", "organizationName": "Acme", "accessToken": "secret"})
	writeProviderCredential(t, filepath.Join(homeDir, ".claude", "settings.json"), map[string]string{"theme": "dark"})

	p := ClaudeProvider{}
	session, found, err := p.DetectGlobalCredentialSession(GlobalCredentialContext{UserHomeDir: homeDir})
	if err != nil || !found || metadataField(session.Fields, "Organization") != "Acme" {
		t.Fatalf("session = %#v found=%t err=%v", session, found, err)
	}

	destination := filepath.Join(t.TempDir(), "providers", "claude")
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := p.ImportCredentials(CredentialImportContext{UserHomeDir: homeDir, Runtime: RuntimeContext{Paths: ContextPaths{StorageDir: destination}}, Files: credentialFileOperations{}}); err != nil {
		t.Fatalf("import credentials: %v", err)
	}
	identity, available, err := p.DetectContextIdentity(RuntimeContext{Paths: ContextPaths{StorageDir: destination}})
	if err != nil || !available || metadataField(identity.Fields, "Subscription") != "Pro" {
		t.Fatalf("identity = %#v available=%t err=%v", identity, available, err)
	}
}

type credentialFileOperations struct{}

func (credentialFileOperations) FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil && info.Mode().IsRegular(), err
}

func (credentialFileOperations) CopyOpaqueFile(source, destination string) error {
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if _, err := os.Stat(destination); err == nil {
		return nil
	}
	return os.WriteFile(destination, data, 0o600)
}

func metadataField(fields []MetadataField, label string) string {
	for _, field := range fields {
		if field.Label == label {
			return field.Value
		}
	}
	return ""
}

func writeProviderCredential(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func testCredentialJWT(t *testing.T, claims map[string]string) string {
	t.Helper()
	payload, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join([]string{"header", base64.RawURLEncoding.EncodeToString(payload), "signature"}, ".")
}
