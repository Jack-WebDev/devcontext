package filesystem

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	providerIDCodex  = "codex"
	providerIDClaude = "claude"

	globalCodexDirectoryName  = ".codex"
	globalClaudeDirectoryName = ".claude"

	codexAuthFileName         = "auth.json"
	claudeCredentialsFileName = ".credentials.json"
	claudeSettingsFileName    = "settings.json"
)

// DetectedProviderCredentialSession contains only non-secret metadata that can
// help a user recognize a globally authenticated provider session.
type DetectedProviderCredentialSession struct {
	ProviderID        string
	MetadataAvailable bool
	Codex             CodexCredentialMetadata
	Claude            ClaudeCredentialMetadata
}

// CodexCredentialMetadata contains safe identity claims decoded from the Codex
// id_token JWT payload.
type CodexCredentialMetadata struct {
	Email            string
	ChatGPTPlanType  string
	ChatGPTAccountID string
}

// ClaudeCredentialMetadata contains safe identity fields decoded from Claude
// credentials.
type ClaudeCredentialMetadata struct {
	SubscriptionType string
	OrganizationUUID string
	OrganizationName string
}

// DetectProviderCredentialSessions reads supported global provider credential
// files and returns only safe metadata for user classification.
func DetectProviderCredentialSessions(paths PlatformPaths) ([]DetectedProviderCredentialSession, error) {
	if paths == nil {
		return nil, ErrUserHomeUnavailable
	}

	homeDir, err := paths.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve provider credential source directory: %w", err)
	}

	sessions := make([]DetectedProviderCredentialSession, 0, 2)
	codexSession, ok, err := detectCodexCredentialSession(homeDir)
	if err != nil {
		return nil, err
	}
	if ok {
		sessions = append(sessions, codexSession)
	}

	claudeSession, ok, err := detectClaudeCredentialSession(homeDir)
	if err != nil {
		return nil, err
	}
	if ok {
		sessions = append(sessions, claudeSession)
	}
	return sessions, nil
}

// DetectCodexContextCredentialMetadata reads Codex identity metadata from one
// context-owned provider directory. It returns only verified safe fields decoded
// from local credential state.
func DetectCodexContextCredentialMetadata(paths ContextPaths) (CodexCredentialMetadata, bool, error) {
	path := joinPlatformPath(paths.CodexDir, codexAuthFileName)
	exists, err := regularFileExists(path)
	if err != nil {
		return CodexCredentialMetadata{}, false, err
	}
	if !exists {
		return CodexCredentialMetadata{}, false, nil
	}

	metadata, available := readCodexCredentialMetadata(path)
	return metadata, available, nil
}

// DetectClaudeContextCredentialMetadata reads Claude identity metadata from one
// context-owned provider directory. It returns only verified safe fields decoded
// from local credential state.
func DetectClaudeContextCredentialMetadata(paths ContextPaths) (ClaudeCredentialMetadata, bool, error) {
	path := joinPlatformPath(paths.ClaudeDir, claudeCredentialsFileName)
	exists, err := regularFileExists(path)
	if err != nil {
		return ClaudeCredentialMetadata{}, false, err
	}
	if !exists {
		return ClaudeCredentialMetadata{}, false, nil
	}

	metadata, available := readClaudeCredentialMetadata(path)
	return metadata, available, nil
}

// ImportProviderCredentials copies supported global provider credential files
// into one context-owned provider tree. Credential files are treated as opaque
// bytes and existing isolated files are never overwritten.
func ImportProviderCredentials(paths PlatformPaths, contextPaths ContextPaths, providerIDs []string) error {
	return ImportProviderCredentialsWithPermissions(paths, contextPaths, providerIDs, NewDefaultStoragePermissions())
}

// ImportProviderCredentialsWithPermissions copies supported global provider
// credential files using the supplied storage permission policy.
func ImportProviderCredentialsWithPermissions(paths PlatformPaths, contextPaths ContextPaths, providerIDs []string, permissions StoragePermissions) error {
	if paths == nil {
		return ErrUserHomeUnavailable
	}
	if permissions == nil {
		permissions = NewDefaultStoragePermissions()
	}

	homeDir, err := paths.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve provider credential source directory: %w", err)
	}

	selected := selectedProviderIDs(providerIDs)
	if selected[providerIDCodex] {
		if err := importCodexCredentials(homeDir, contextPaths, permissions); err != nil {
			return err
		}
	}
	if selected[providerIDClaude] {
		if err := importClaudeCredentials(homeDir, contextPaths, permissions); err != nil {
			return err
		}
	}
	return nil
}

func detectCodexCredentialSession(homeDir string) (DetectedProviderCredentialSession, bool, error) {
	source := codexAuthPath(homeDir)
	exists, err := regularFileExists(source)
	if err != nil {
		return DetectedProviderCredentialSession{}, false, err
	}
	if !exists {
		return DetectedProviderCredentialSession{}, false, nil
	}

	metadata, available := readCodexCredentialMetadata(source)
	return DetectedProviderCredentialSession{
		ProviderID:        providerIDCodex,
		MetadataAvailable: available,
		Codex:             metadata,
	}, true, nil
}

func detectClaudeCredentialSession(homeDir string) (DetectedProviderCredentialSession, bool, error) {
	source := claudeCredentialsPath(homeDir)
	exists, err := regularFileExists(source)
	if err != nil {
		return DetectedProviderCredentialSession{}, false, err
	}
	if !exists {
		return DetectedProviderCredentialSession{}, false, nil
	}

	metadata, available := readClaudeCredentialMetadata(source)
	return DetectedProviderCredentialSession{
		ProviderID:        providerIDClaude,
		MetadataAvailable: available,
		Claude:            metadata,
	}, true, nil
}

func importCodexCredentials(homeDir string, paths ContextPaths, permissions StoragePermissions) error {
	source := codexAuthPath(homeDir)
	exists, err := regularFileExists(source)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	destination := joinPlatformPath(paths.CodexDir, codexAuthFileName)
	return copyOpaqueCredentialFile(source, destination, permissions)
}

func importClaudeCredentials(homeDir string, paths ContextPaths, permissions StoragePermissions) error {
	credentialsSource := claudeCredentialsPath(homeDir)
	exists, err := regularFileExists(credentialsSource)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	credentialsDestination := joinPlatformPath(paths.ClaudeDir, claudeCredentialsFileName)
	if err := copyOpaqueCredentialFile(credentialsSource, credentialsDestination, permissions); err != nil {
		return err
	}

	settingsSource := claudeSettingsPath(homeDir)
	exists, err = regularFileExists(settingsSource)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	settingsDestination := joinPlatformPath(paths.ClaudeDir, claudeSettingsFileName)
	return copyOpaqueCredentialFile(settingsSource, settingsDestination, permissions)
}

func selectedProviderIDs(providerIDs []string) map[string]bool {
	selected := make(map[string]bool, len(providerIDs))
	for _, providerID := range providerIDs {
		switch strings.TrimSpace(providerID) {
		case providerIDCodex:
			selected[providerIDCodex] = true
		case providerIDClaude:
			selected[providerIDClaude] = true
		}
	}
	return selected
}

func codexAuthPath(homeDir string) string {
	return joinPlatformPath(joinPlatformPath(homeDir, globalCodexDirectoryName), codexAuthFileName)
}

func claudeCredentialsPath(homeDir string) string {
	return joinPlatformPath(joinPlatformPath(homeDir, globalClaudeDirectoryName), claudeCredentialsFileName)
}

func claudeSettingsPath(homeDir string) string {
	return joinPlatformPath(joinPlatformPath(homeDir, globalClaudeDirectoryName), claudeSettingsFileName)
}

type codexIDTokenClaims struct {
	Email            string `json:"email"`
	ChatGPTPlanType  string `json:"chatgpt_plan_type"`
	ChatGPTAccountID string `json:"chatgpt_account_id"`
}

func readCodexCredentialMetadata(path string) (CodexCredentialMetadata, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return CodexCredentialMetadata{}, false
	}

	var auth any
	if err := json.Unmarshal(data, &auth); err != nil {
		return CodexCredentialMetadata{}, false
	}
	idToken, ok := findJSONFieldString(auth, "id_token", "idToken")
	if !ok {
		return CodexCredentialMetadata{}, false
	}
	payload, ok := decodeJWTPayload(idToken)
	if !ok {
		return CodexCredentialMetadata{}, false
	}

	var claims codexIDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return CodexCredentialMetadata{}, false
	}

	metadata := CodexCredentialMetadata{
		Email:            claims.Email,
		ChatGPTPlanType:  claims.ChatGPTPlanType,
		ChatGPTAccountID: claims.ChatGPTAccountID,
	}
	return metadata, metadata != (CodexCredentialMetadata{})
}

func findJSONFieldString(value any, fieldNames ...string) (string, bool) {
	switch typed := value.(type) {
	case map[string]any:
		for _, fieldName := range fieldNames {
			if value, ok := typed[fieldName].(string); ok && value != "" {
				return value, true
			}
		}
		for _, nested := range typed {
			if value, ok := findJSONFieldString(nested, fieldNames...); ok {
				return value, true
			}
		}
	case []any:
		for _, nested := range typed {
			if value, ok := findJSONFieldString(nested, fieldNames...); ok {
				return value, true
			}
		}
	}
	return "", false
}

func decodeJWTPayload(token string) ([]byte, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil, false
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err == nil {
		return payload, true
	}
	payload, err = base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	return payload, true
}

func readClaudeCredentialMetadata(path string) (ClaudeCredentialMetadata, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ClaudeCredentialMetadata{}, false
	}

	var credentials any
	if err := json.Unmarshal(data, &credentials); err != nil {
		return ClaudeCredentialMetadata{}, false
	}

	metadata := ClaudeCredentialMetadata{
		SubscriptionType: firstJSONFieldString(credentials, "subscriptionType", "subscription_type"),
		OrganizationUUID: firstJSONFieldString(credentials, "organizationUuid", "organizationUUID", "organization_uuid"),
		OrganizationName: claudeOrganizationName(credentials),
	}
	return metadata, metadata != (ClaudeCredentialMetadata{})
}

func firstJSONFieldString(value any, fieldNames ...string) string {
	found, ok := findJSONFieldString(value, fieldNames...)
	if !ok {
		return ""
	}
	return found
}

func claudeOrganizationName(credentials any) string {
	if name := firstJSONFieldString(
		credentials,
		"organizationName",
		"organizationDisplayName",
		"organization_name",
		"organization_display_name",
		"orgName",
		"orgDisplayName",
		"workspaceName",
		"workspaceDisplayName",
	); name != "" {
		return name
	}

	return firstStringFieldInNamedJSONObject(
		credentials,
		[]string{"organization", "organizationInfo", "organization_info", "org", "workspace"},
		[]string{"name", "displayName", "display_name"},
	)
}

func firstStringFieldInNamedJSONObject(value any, objectNames []string, fieldNames []string) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, objectName := range objectNames {
			if nested, ok := typed[objectName]; ok {
				if found := firstStringField(nested, fieldNames); found != "" {
					return found
				}
			}
		}
		for _, nested := range typed {
			if found := firstStringFieldInNamedJSONObject(nested, objectNames, fieldNames); found != "" {
				return found
			}
		}
	case []any:
		for _, nested := range typed {
			if found := firstStringFieldInNamedJSONObject(nested, objectNames, fieldNames); found != "" {
				return found
			}
		}
	}
	return ""
}

func firstStringField(value any, fieldNames []string) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	for _, fieldName := range fieldNames {
		if found, ok := object[fieldName].(string); ok && found != "" {
			return found
		}
	}
	return ""
}

func regularFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	switch {
	case err == nil:
		if !info.Mode().IsRegular() {
			return false, fmt.Errorf("provider credential source %q is not a regular file", path)
		}
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		if wrapped := WrapStoragePermissionError("inspect provider credential source", path, err); wrapped != err {
			return false, fmt.Errorf("inspect provider credential source %q: %w", path, wrapped)
		}
		return false, fmt.Errorf("inspect provider credential source %q: %w", path, err)
	}
}

func copyOpaqueCredentialFile(source string, destination string, permissions StoragePermissions) error {
	if exists, err := pathExists(destination); err != nil {
		return err
	} else if exists {
		return nil
	}

	sourceFile, err := os.Open(source)
	if err != nil {
		if wrapped := WrapStoragePermissionError("open provider credential source", source, err); wrapped != err {
			return fmt.Errorf("open provider credential source %q: %w", source, wrapped)
		}
		return fmt.Errorf("open provider credential source %q: %w", source, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, permissions.FileMode())
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		if wrapped := WrapStoragePermissionError("create provider credential destination", destination, err); wrapped != err {
			return fmt.Errorf("create provider credential destination %q: %w", destination, wrapped)
		}
		return fmt.Errorf("create provider credential destination %q: %w", destination, err)
	}

	created := true
	defer func() {
		if created {
			_ = os.Remove(destination)
		}
	}()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		_ = destinationFile.Close()
		if wrapped := WrapStoragePermissionError("copy provider credential", destination, err); wrapped != err {
			return fmt.Errorf("copy provider credential to %q: %w", destination, wrapped)
		}
		return fmt.Errorf("copy provider credential to %q: %w", destination, err)
	}
	if err := destinationFile.Close(); err != nil {
		if wrapped := WrapStoragePermissionError("close provider credential destination", destination, err); wrapped != err {
			return fmt.Errorf("close provider credential destination %q: %w", destination, wrapped)
		}
		return fmt.Errorf("close provider credential destination %q: %w", destination, err)
	}
	if err := permissions.ApplyFile(destination); err != nil {
		return err
	}

	created = false
	return nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	switch {
	case err == nil:
		return true, nil
	case os.IsNotExist(err):
		return false, nil
	default:
		if wrapped := WrapStoragePermissionError("inspect provider credential destination", path, err); wrapped != err {
			return false, fmt.Errorf("inspect provider credential destination %q: %w", path, wrapped)
		}
		return false, fmt.Errorf("inspect provider credential destination %q: %w", path, err)
	}
}
