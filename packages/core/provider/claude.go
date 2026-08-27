package provider

import (
	"encoding/json"
	"path/filepath"
)

const (
	// ClaudeID identifies the Claude provider configuration.
	ClaudeID ID = "claude"

	// ClaudeCommand is the executable name used to detect local Claude Code
	// presence.
	ClaudeCommand = "claude"

	// ClaudeConfigDirEnvVar is the environment variable Claude Code uses for
	// its configuration directory.
	ClaudeConfigDirEnvVar = "CLAUDE_CONFIG_DIR"
)

// ClaudeProvider contributes isolated Claude Code process configuration.
type ClaudeProvider struct {
	Probe StatusProbe
}

var _ Provider = ClaudeProvider{}
var _ GlobalCredentialDetector = ClaudeProvider{}
var _ CredentialMetadataExtractor = ClaudeProvider{}
var _ CredentialImporter = ClaudeProvider{}
var _ ContextIdentityDetector = ClaudeProvider{}

// ID returns the persisted provider identifier.
func (ClaudeProvider) ID() ID {
	return ClaudeID
}

// DisplayName returns the user-facing provider name.
func (ClaudeProvider) DisplayName() string {
	return "Claude"
}

// BuildEnvironment points Claude Code at the selected context's isolated config
// directory.
func (ClaudeProvider) BuildEnvironment(ctx RuntimeContext) (EnvironmentContribution, error) {
	return EnvironmentContribution{
		ClaudeConfigDirEnvVar: ctx.Paths.StorageDir,
	}, nil
}

// Status returns local provider readiness.
func (p ClaudeProvider) Status(ctx RuntimeContext) (Status, error) {
	return detectLocalStatus(p.Probe, p.DisplayName(), ctx.Paths.StorageDir)
}

// DetectGlobalCredentialSession identifies the local Claude session without
// exposing credential contents.
func (ClaudeProvider) DetectGlobalCredentialSession(ctx GlobalCredentialContext) (CredentialSession, bool, error) {
	fields, available, exists, err := claudeMetadataFromFile(filepath.Join(ctx.UserHomeDir, ".claude", ".credentials.json"))
	if err != nil || !exists {
		return CredentialSession{}, exists, err
	}
	return CredentialSession{MetadataAvailable: available, Fields: fields}, true, nil
}

// ExtractCredentialMetadata decodes safe Claude metadata from a credentials file.
func (ClaudeProvider) ExtractCredentialMetadata(path string) ([]MetadataField, bool, error) {
	fields, available, _, err := claudeMetadataFromFile(path)
	return fields, available, err
}

// ImportCredentials copies Claude's opaque credentials and settings files into
// its isolated storage.
func (ClaudeProvider) ImportCredentials(ctx CredentialImportContext) error {
	if ctx.Files == nil {
		return nil
	}
	sourceDir := filepath.Join(ctx.UserHomeDir, ".claude")
	credentialsSource := filepath.Join(sourceDir, ".credentials.json")
	exists, err := ctx.Files.FileExists(credentialsSource)
	if err != nil || !exists {
		return err
	}
	if err := ctx.Files.CopyOpaqueFile(credentialsSource, filepath.Join(ctx.Runtime.Paths.StorageDir, ".credentials.json")); err != nil {
		return err
	}
	for _, name := range []string{"settings.json"} {
		source := filepath.Join(sourceDir, name)
		exists, err := ctx.Files.FileExists(source)
		if err != nil {
			return err
		}
		if exists {
			if err := ctx.Files.CopyOpaqueFile(source, filepath.Join(ctx.Runtime.Paths.StorageDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

// DetectContextIdentity returns safe metadata from Claude credentials isolated
// within the selected Dev Context.
func (ClaudeProvider) DetectContextIdentity(ctx RuntimeContext) (Identity, bool, error) {
	fields, available, _, err := claudeMetadataFromFile(filepath.Join(ctx.Paths.StorageDir, ".credentials.json"))
	if err != nil || !available {
		return Identity{}, false, err
	}
	return Identity{Fields: fields}, true, nil
}

func claudeMetadataFromFile(path string) ([]MetadataField, bool, bool, error) {
	data, exists, err := credentialFileData(path)
	if err != nil || !exists {
		return nil, false, exists, err
	}
	var credentials any
	if err := json.Unmarshal(data, &credentials); err != nil {
		return nil, false, true, nil
	}
	fields := metadataFields(
		MetadataField{Label: "Subscription", Value: firstJSONFieldString(credentials, "subscriptionType", "subscription_type")},
		MetadataField{Label: "Organization UUID", Value: firstJSONFieldString(credentials, "organizationUuid", "organizationUUID", "organization_uuid")},
		MetadataField{Label: "Organization", Value: claudeOrganizationName(credentials)},
	)
	return fields, len(fields) > 0, true, nil
}
