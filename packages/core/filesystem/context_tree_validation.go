package filesystem

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	devcontext "devctx/packages/core/context"
)

// ErrContextStorageIncomplete identifies a context whose expected storage
// directories are missing or unusable.
var ErrContextStorageIncomplete = errors.New("context storage is incomplete")

// ContextDirectoryKind identifies one expected directory in a context tree.
type ContextDirectoryKind string

const (
	ContextDirectoryRoot     ContextDirectoryKind = "context"
	ContextDirectoryProvider ContextDirectoryKind = "provider"
)

// MissingContextDirectory describes one absent or non-directory context path.
type MissingContextDirectory struct {
	Kind       ContextDirectoryKind
	ProviderID string
	Path       string
	Reason     string
}

// ContextStorageError reports incomplete storage for one context.
type ContextStorageError struct {
	ContextID devcontext.ID
	Missing   []MissingContextDirectory
}

func (e *ContextStorageError) Error() string {
	if e == nil {
		return ""
	}

	details := make([]string, 0, len(e.Missing))
	for _, missing := range e.Missing {
		kind := string(missing.Kind)
		if missing.ProviderID != "" {
			kind += ":" + missing.ProviderID
		}
		if missing.Reason == "" {
			details = append(details, fmt.Sprintf("%s %q", kind, missing.Path))
			continue
		}
		details = append(details, fmt.Sprintf("%s %q (%s)", kind, missing.Path, missing.Reason))
	}
	if len(details) == 0 {
		return fmt.Sprintf("%v: context %q", ErrContextStorageIncomplete, e.ContextID.String())
	}
	return fmt.Sprintf("%v: context %q missing %s", ErrContextStorageIncomplete, e.ContextID.String(), strings.Join(details, ", "))
}

func (e *ContextStorageError) Unwrap() error {
	return ErrContextStorageIncomplete
}

// ValidateContextDirectoryTree verifies that all expected context-owned
// directories exist. It reports incomplete storage instead of recreating paths.
func ValidateContextDirectoryTree(paths ContextPaths) error {
	missing := make([]MissingContextDirectory, 0)
	for _, expected := range expectedContextDirectories(paths) {
		info, err := os.Stat(expected.Path)
		if err != nil {
			if isMissingContextDirectoryError(err) {
				missing = append(missing, MissingContextDirectory{
					Kind:       expected.Kind,
					ProviderID: expected.ProviderID,
					Path:       expected.Path,
					Reason:     "missing",
				})
				continue
			}
			if wrapped := WrapStoragePermissionError("inspect directory", expected.Path, err); wrapped != err {
				return wrapped
			}
			return fmt.Errorf("inspect context directory %q: %w", expected.Path, err)
		}
		if !info.IsDir() {
			missing = append(missing, MissingContextDirectory{
				Kind:       expected.Kind,
				ProviderID: expected.ProviderID,
				Path:       expected.Path,
				Reason:     "not a directory",
			})
		}
	}

	if len(missing) > 0 {
		return &ContextStorageError{
			ContextID: paths.ContextID,
			Missing:   missing,
		}
	}
	return nil
}

type expectedContextDirectory struct {
	Kind       ContextDirectoryKind
	ProviderID string
	Path       string
}

func expectedContextDirectories(paths ContextPaths) []expectedContextDirectory {
	expected := []expectedContextDirectory{
		{Kind: ContextDirectoryRoot, Path: paths.RootDir},
	}
	for _, providerID := range sortedProviderStorageIDs(paths.ProviderStorageDirs) {
		expected = append(expected, expectedContextDirectory{
			Kind:       ContextDirectoryProvider,
			ProviderID: string(providerID),
			Path:       paths.ProviderStorageDirs[providerID],
		})
	}
	return expected
}

func isMissingContextDirectoryError(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR)
}
