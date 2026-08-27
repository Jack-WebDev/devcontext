package filesystem

import (
	"fmt"
	"io"
	"os"

	"devctx/packages/core/provider"
)

// ProviderStoragePath joins provider-owned storage path elements using the
// platform style already present in storageDir.
func ProviderStoragePath(storageDir string, elements ...string) string {
	path := storageDir
	for _, element := range elements {
		path = joinPlatformPath(path, element)
	}
	return path
}

// ProviderCredentialFileExists reports whether path exists as a regular file.
// Non-regular files are reported as errors.
func ProviderCredentialFileExists(path string) (bool, error) { return regularFileExists(path) }

// CopyOpaqueProviderCredentialFile copies a provider credential file without
// interpreting its contents. Existing destination files are left untouched.
func CopyOpaqueProviderCredentialFile(source string, destination string, permissions StoragePermissions) error {
	if permissions == nil {
		permissions = NewDefaultStoragePermissions()
	}
	return copyOpaqueCredentialFile(source, destination, permissions)
}

// NewProviderCredentialFileOperations returns generic credential file
// operations for provider-owned import implementations.
func NewProviderCredentialFileOperations(permissions StoragePermissions) provider.CredentialFileOperations {
	return providerCredentialFileOperations{permissions: permissions}
}

type providerCredentialFileOperations struct{ permissions StoragePermissions }

func (o providerCredentialFileOperations) FileExists(path string) (bool, error) {
	return ProviderCredentialFileExists(path)
}

func (o providerCredentialFileOperations) CopyOpaqueFile(source string, destination string) error {
	return CopyOpaqueProviderCredentialFile(source, destination, o.permissions)
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
		return fmt.Errorf("open provider credential source %q: %w", source, err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, permissions.FileMode())
	if err != nil {
		if os.IsExist(err) {
			return nil
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
		return fmt.Errorf("copy provider credential to %q: %w", destination, err)
	}
	if err := destinationFile.Close(); err != nil {
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
		return false, fmt.Errorf("inspect provider credential destination %q: %w", path, err)
	}
}
