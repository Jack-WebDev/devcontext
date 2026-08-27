package provider

import (
	"encoding/base64"
	"encoding/json"
	"path/filepath"
	"strings"
)

const (
	// CodexID identifies the Codex provider configuration.
	CodexID ID = "codex"

	// CodexCommand is the executable name used to detect local Codex presence.
	CodexCommand = "codex"

	// CodexHomeEnvVar is the environment variable Codex uses for its home.
	CodexHomeEnvVar = "CODEX_HOME"
)

// CodexProvider contributes isolated Codex process configuration.
type CodexProvider struct {
	Probe StatusProbe
}

var _ Provider = CodexProvider{}
var _ GlobalCredentialDetector = CodexProvider{}
var _ CredentialMetadataExtractor = CodexProvider{}
var _ CredentialImporter = CodexProvider{}
var _ ContextIdentityDetector = CodexProvider{}
var _ SetupGuidanceProvider = CodexProvider{}

// ID returns the persisted provider identifier.
func (CodexProvider) ID() ID {
	return CodexID
}

// DisplayName returns the user-facing provider name.
func (CodexProvider) DisplayName() string {
	return "Codex"
}

// BuildEnvironment points Codex at the selected context's isolated home.
func (CodexProvider) BuildEnvironment(ctx RuntimeContext) (EnvironmentContribution, error) {
	return EnvironmentContribution{
		CodexHomeEnvVar: ctx.Paths.StorageDir,
	}, nil
}

// Status returns local provider readiness.
func (p CodexProvider) Status(ctx RuntimeContext) (Status, error) {
	return detectLocalStatus(p.Probe, p.DisplayName(), ctx.Paths.StorageDir)
}

// DetectGlobalCredentialSession identifies the local Codex session without
// exposing credential contents.
func (CodexProvider) DetectGlobalCredentialSession(ctx GlobalCredentialContext) (CredentialSession, bool, error) {
	fields, available, exists, err := codexMetadataFromFile(filepath.Join(ctx.UserHomeDir, ".codex", "auth.json"))
	if err != nil || !exists {
		return CredentialSession{}, exists, err
	}
	return CredentialSession{MetadataAvailable: available, Fields: fields}, true, nil
}

// ExtractCredentialMetadata decodes safe Codex claims from an auth file.
func (CodexProvider) ExtractCredentialMetadata(path string) ([]MetadataField, bool, error) {
	fields, available, _, err := codexMetadataFromFile(path)
	return fields, available, err
}

// ImportCredentials copies Codex's opaque auth file into its isolated storage.
func (CodexProvider) ImportCredentials(ctx CredentialImportContext) error {
	if ctx.Files == nil {
		return nil
	}
	source := filepath.Join(ctx.UserHomeDir, ".codex", "auth.json")
	exists, err := ctx.Files.FileExists(source)
	if err != nil || !exists {
		return err
	}
	return ctx.Files.CopyOpaqueFile(source, filepath.Join(ctx.Runtime.Paths.StorageDir, "auth.json"))
}

// DetectContextIdentity returns safe metadata from Codex credentials isolated
// within the selected Dev Context.
func (CodexProvider) DetectContextIdentity(ctx RuntimeContext) (Identity, bool, error) {
	fields, available, _, err := codexMetadataFromFile(filepath.Join(ctx.Paths.StorageDir, "auth.json"))
	if err != nil || !available {
		return Identity{}, false, err
	}
	return Identity{Fields: fields}, true, nil
}

// SetupGuidance describes the next safe setup action for Codex.
func (CodexProvider) SetupGuidance(RuntimeContext) SetupGuidance {
	return SetupGuidance{ActionHint: "Sign in to Codex for this context."}
}

type codexIDTokenClaims struct {
	Email            string `json:"email"`
	ChatGPTPlanType  string `json:"chatgpt_plan_type"`
	ChatGPTAccountID string `json:"chatgpt_account_id"`
}

func codexMetadataFromFile(path string) ([]MetadataField, bool, bool, error) {
	data, exists, err := credentialFileData(path)
	if err != nil || !exists {
		return nil, false, exists, err
	}
	var auth any
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, false, true, nil
	}
	idToken, ok := findJSONFieldString(auth, "id_token", "idToken")
	if !ok {
		return nil, false, true, nil
	}
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, false, true, nil
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, false, true, nil
		}
	}
	var claims codexIDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, false, true, nil
	}
	fields := metadataFields(
		MetadataField{Label: "Email", Value: claims.Email},
		MetadataField{Label: "ChatGPT plan", Value: claims.ChatGPTPlanType},
		MetadataField{Label: "ChatGPT account ID", Value: claims.ChatGPTAccountID},
	)
	return fields, len(fields) > 0, true, nil
}
