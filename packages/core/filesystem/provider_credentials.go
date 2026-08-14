package filesystem

import (
	"fmt"
	"io"
	"os"
)

const (
	globalCodexDirectoryName  = ".codex"
	globalClaudeDirectoryName = ".claude"

	codexAuthFileName         = "auth.json"
	claudeCredentialsFileName = ".credentials.json"
	claudeSettingsFileName    = "settings.json"
)

// ImportProviderCredentials copies supported global provider credential files
// into one context-owned provider tree. Credential files are treated as opaque
// bytes and existing isolated files are never overwritten.
func ImportProviderCredentials(paths PlatformPaths, contextPaths ContextPaths) error {
	return ImportProviderCredentialsWithPermissions(paths, contextPaths, NewDefaultStoragePermissions())
}

// ImportProviderCredentialsWithPermissions copies supported global provider
// credential files using the supplied storage permission policy.
func ImportProviderCredentialsWithPermissions(paths PlatformPaths, contextPaths ContextPaths, permissions StoragePermissions) error {
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

	if err := importCodexCredentials(homeDir, contextPaths, permissions); err != nil {
		return err
	}
	if err := importClaudeCredentials(homeDir, contextPaths, permissions); err != nil {
		return err
	}
	return nil
}

func importCodexCredentials(homeDir string, paths ContextPaths, permissions StoragePermissions) error {
	source := joinPlatformPath(joinPlatformPath(homeDir, globalCodexDirectoryName), codexAuthFileName)
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
	globalClaudeDir := joinPlatformPath(homeDir, globalClaudeDirectoryName)
	credentialsSource := joinPlatformPath(globalClaudeDir, claudeCredentialsFileName)
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

	settingsSource := joinPlatformPath(globalClaudeDir, claudeSettingsFileName)
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
