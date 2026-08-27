package launcher

import (
	"errors"
	"os"
	"sort"
	"strings"

	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

// ProviderConfidenceCheck derives a UI-safe confidence check for one registered
// provider status.
func ProviderConfidenceCheck(providerID provider.ID, displayName string, status provider.Status) (ConfidenceCheck, bool) {
	if strings.TrimSpace(string(providerID)) == "" {
		return ConfidenceCheck{}, false
	}

	name := strings.TrimSpace(displayName)
	if name == "" {
		name = string(providerID)
	}

	check := ConfidenceCheck{
		Component:  ConfidenceCheckProvider,
		ProviderID: string(providerID),
		Label:      name,
	}

	switch status.State {
	case provider.StatusConfigured:
		check.Severity = ConfidenceReady
		check.Message = name + " is ready for this context."
	case provider.StatusNotConfigured:
		check.Severity = ConfidenceNeedsAttention
		check.Message = confidenceMessage(status.Explanation, name+" is not configured for this context.")
		check.ActionHint = "Open and configure " + name + " for this context."
	case provider.StatusDirectoryMissing:
		check.Severity = ConfidenceBlocked
		check.Message = confidenceMessage(status.Explanation, name+" isolated provider directory is missing.")
		check.ActionHint = "Run diagnostics to repair context storage."
	case provider.StatusUnavailable:
		check.Severity = ConfidenceNeedsAttention
		check.Message = confidenceMessage(status.Explanation, name+" readiness could not be determined.")
		check.ActionHint = "Run diagnostics to inspect " + name + "."
	default:
		check.Severity = ConfidenceNeedsAttention
		check.Message = name + " readiness could not be determined."
		check.ActionHint = "Run diagnostics to inspect " + name + "."
	}

	return check, true
}

// VSCodeConfidenceCheck derives the VS Code readiness check from executable
// detection output.
func VSCodeConfidenceCheck(executable editor.Executable, err error) ConfidenceCheck {
	check := ConfidenceCheck{
		Component: ConfidenceCheckVSCode,
		Label:     "VS Code",
	}

	if err == nil && strings.TrimSpace(string(executable)) != "" {
		check.Severity = ConfidenceReady
		check.Message = "VS Code is available for launch."
		return check
	}

	check.Severity = ConfidenceBlocked
	check.ActionHint = "Install the VS Code command line launcher or configure the VS Code executable."

	switch {
	case err == nil:
		check.Message = "Dev Context could not find a VS Code command to launch."
	case errors.Is(err, editor.ErrExecutableNotFound):
		check.Message = "Dev Context could not find a VS Code command to launch."
	case errors.Is(err, editor.ErrExecutableNotExecutable):
		check.Message = "The configured VS Code command cannot be run."
	default:
		check.Message = "VS Code readiness could not be checked."
	}

	return check
}

// IsolationConfidenceChecks derives readiness checks for the context-owned
// isolation storage required by launch and editor profile isolation.
func IsolationConfidenceChecks(paths filesystem.ContextPaths) []ConfidenceCheck {
	return []ConfidenceCheck{
		directoryConfidenceCheck(ConfidenceCheck{
			Component: ConfidenceCheckIsolation,
			Label:     "Context storage",
			Severity:  ConfidenceReady,
			Message:   "Context storage is ready.",
		}, paths.RootDir, "Context storage is not ready."),
		multiDirectoryConfidenceCheck(ConfidenceCheck{
			Component: ConfidenceCheckIsolation,
			Label:     "Provider isolation",
			Severity:  ConfidenceReady,
			Message:   "Provider isolation directories are ready.",
		}, providerStorageDirectories(paths), "Provider isolation storage is incomplete."),
		multiDirectoryConfidenceCheck(ConfidenceCheck{
			Component: ConfidenceCheckIsolation,
			Label:     "VS Code profile",
			Severity:  ConfidenceReady,
			Message:   "VS Code profile isolation is ready.",
		}, []string{paths.VSCodeDir, paths.VSCodeUserDataDir}, "VS Code profile isolation is not ready."),
	}
}

func providerStorageDirectories(paths filesystem.ContextPaths) []string {
	providerIDs := make([]provider.ID, 0, len(paths.ProviderStorageDirs))
	for providerID := range paths.ProviderStorageDirs {
		providerIDs = append(providerIDs, providerID)
	}
	sort.Slice(providerIDs, func(i int, j int) bool {
		return providerIDs[i] < providerIDs[j]
	})

	dirs := make([]string, 0, len(providerIDs))
	for _, providerID := range providerIDs {
		dirs = append(dirs, paths.ProviderStorageDirs[providerID])
	}
	return dirs
}

func confidenceMessage(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func multiDirectoryConfidenceCheck(ready ConfidenceCheck, paths []string, blockedMessage string) ConfidenceCheck {
	for _, path := range paths {
		if !directoryReady(path) {
			return blockedIsolationCheck(ready.Label, blockedMessage)
		}
	}
	return ready
}

func directoryConfidenceCheck(ready ConfidenceCheck, path string, blockedMessage string) ConfidenceCheck {
	if !directoryReady(path) {
		return blockedIsolationCheck(ready.Label, blockedMessage)
	}
	return ready
}

func blockedIsolationCheck(label string, message string) ConfidenceCheck {
	return ConfidenceCheck{
		Component:  ConfidenceCheckIsolation,
		Severity:   ConfidenceBlocked,
		Label:      label,
		Message:    message,
		ActionHint: "Run diagnostics to repair context storage.",
	}
}

func directoryReady(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
