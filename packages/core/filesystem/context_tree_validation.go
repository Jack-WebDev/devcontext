package filesystem

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

// ErrContextStorageIncomplete identifies a context whose expected storage
// directories are missing or unusable.
var ErrContextStorageIncomplete = errors.New("context storage is incomplete")

// ContextDirectoryKind identifies one expected directory in a context tree.
type ContextDirectoryKind string

const (
	ContextDirectoryRoot     ContextDirectoryKind = "context"
	ContextDirectoryTool     ContextDirectoryKind = "tool"
	ContextDirectoryProvider ContextDirectoryKind = "provider"
)

// MissingContextDirectory describes one absent or non-directory context path.
type MissingContextDirectory struct {
	Kind                ContextDirectoryKind
	ProviderID          string
	ProviderDisplayName string
	Path                string
	Reason              string
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
			if missing.ProviderDisplayName != "" {
				kind += " (" + missing.ProviderDisplayName + ")"
			}
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
	return validateContextDirectoryTree(paths, provider.Registry{})
}

// ValidateContextDirectoryTreeWithProviderRegistry verifies that all expected
// context-owned directories exist for registered providers enabled in ctx. It
// reports incomplete storage instead of recreating paths.
func ValidateContextDirectoryTreeWithProviderRegistry(paths ContextPaths, ctx devcontext.Context, registry provider.Registry) error {
	return ValidateContextDirectoryTreeWithRegistries(paths, ctx, registry, codingtool.BuiltInRegistry())
}

// ValidateContextDirectoryTreeWithRegistries verifies all storage required by
// registered providers enabled in ctx and its registered selected coding tool.
func ValidateContextDirectoryTreeWithRegistries(paths ContextPaths, ctx devcontext.Context, providerRegistry provider.Registry, toolRegistry codingtool.Registry) error {
	if providerRegistry.IsZero() {
		providerRegistry = provider.BuiltInRegistry()
	}
	if toolRegistry.IsZero() {
		toolRegistry = codingtool.BuiltInRegistry()
	}
	paths = paths.WithProviderStorageDirs(registeredEnabledProviderIDs(ctx, providerRegistry))
	paths = paths.WithToolStorageDirs(registeredSelectedToolIDs(ctx, toolRegistry))
	return validateContextDirectoryTree(paths, providerRegistry)
}

func validateContextDirectoryTree(paths ContextPaths, registry provider.Registry) error {
	missing := make([]MissingContextDirectory, 0)
	for _, expected := range expectedContextDirectories(paths, registry) {
		info, err := os.Stat(expected.Path)
		if err != nil {
			if isMissingContextDirectoryError(err) {
				missing = append(missing, MissingContextDirectory{
					Kind:                expected.Kind,
					ProviderID:          expected.ProviderID,
					ProviderDisplayName: expected.ProviderDisplayName,
					Path:                expected.Path,
					Reason:              "missing",
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
				Kind:                expected.Kind,
				ProviderID:          expected.ProviderID,
				ProviderDisplayName: expected.ProviderDisplayName,
				Path:                expected.Path,
				Reason:              "not a directory",
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
	Kind                ContextDirectoryKind
	ProviderID          string
	ProviderDisplayName string
	Path                string
}

func expectedContextDirectories(paths ContextPaths, registry provider.Registry) []expectedContextDirectory {
	expected := []expectedContextDirectory{
		{Kind: ContextDirectoryRoot, Path: paths.RootDir},
		{Kind: ContextDirectoryTool, Path: paths.ToolStorageRootDir},
	}
	for _, toolID := range sortedToolStorageIDs(paths.ToolStorageDirs) {
		expected = append(expected, expectedContextDirectory{Kind: ContextDirectoryTool, Path: paths.ToolStorageDirs[toolID]})
	}
	for _, providerID := range sortedProviderStorageIDs(paths.ProviderStorageDirs) {
		expected = append(expected, expectedContextDirectory{
			Kind:                ContextDirectoryProvider,
			ProviderID:          string(providerID),
			ProviderDisplayName: providerDisplayName(providerID, registry),
			Path:                paths.ProviderStorageDirs[providerID],
		})
	}
	return expected
}

func providerDisplayName(providerID provider.ID, registry provider.Registry) string {
	if registry.IsZero() {
		return ""
	}
	integration, ok := registry.Get(providerID)
	if !ok {
		return ""
	}
	return integration.DisplayName()
}

func isMissingContextDirectoryError(err error) bool {
	return os.IsNotExist(err) || errors.Is(err, syscall.ENOTDIR)
}
