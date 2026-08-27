package launcher

import (
	"errors"
	"os"
	"strings"

	codingtool "devctx/packages/core/codingtool"
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

// ToolConfidenceCheck derives the selected coding tool's readiness check from
// executable detection output.
func ToolConfidenceCheck(toolID codingtool.ID, displayName string, executable codingtool.Executable, err error) ConfidenceCheck {
	name := strings.TrimSpace(displayName)
	if name == "" {
		name = string(toolID)
	}
	check := ConfidenceCheck{
		Component: ConfidenceCheckTool,
		ToolID:    string(toolID),
		Label:     name,
	}

	if err == nil && strings.TrimSpace(string(executable)) != "" {
		check.Severity = ConfidenceReady
		check.Message = name + " is available for launch."
		return check
	}

	check.Severity = ConfidenceBlocked
	check.ActionHint = "Install " + name + " or configure its executable."

	switch {
	case err == nil:
		check.Message = "Dev Context could not find a " + name + " command to launch."
	case errors.Is(err, codingtool.ErrExecutableNotFound):
		check.Message = "Dev Context could not find a " + name + " command to launch."
	case errors.Is(err, codingtool.ErrExecutableNotExecutable):
		check.Message = "The configured " + name + " command cannot be run."
	default:
		check.Message = name + " readiness could not be checked."
	}

	return check
}

// IsolationConfidenceChecks derives readiness checks for the context-owned
// isolation storage required by launch. Provider checks are generated from the
// enabled registered providers passed by the application layer.
func IsolationConfidenceChecks(paths filesystem.ContextPaths, providers []provider.Provider, toolID codingtool.ID, toolName string) []ConfidenceCheck {
	checks := []ConfidenceCheck{
		directoryConfidenceCheck(ConfidenceCheck{
			Component: ConfidenceCheckIsolation,
			Label:     "Context storage",
			Severity:  ConfidenceReady,
			Message:   "Context storage is ready.",
		}, paths.RootDir, "Context storage is not ready."),
	}
	for _, integration := range providers {
		if integration == nil {
			continue
		}
		name := strings.TrimSpace(integration.DisplayName())
		if name == "" {
			name = string(integration.ID())
		}
		checks = append(checks, directoryConfidenceCheck(ConfidenceCheck{
			Component: ConfidenceCheckIsolation,
			Label:     name + " isolation",
			Severity:  ConfidenceReady,
			Message:   name + " isolation storage is ready.",
		}, paths.ProviderStorageDir(integration.ID()), name+" isolation storage is not ready."))
	}
	name := strings.TrimSpace(toolName)
	if name == "" {
		name = string(toolID)
	}
	checks = append(checks, multiDirectoryConfidenceCheck(ConfidenceCheck{
		Component: ConfidenceCheckIsolation,
		ToolID:    string(toolID),
		Label:     name + " isolation",
		Severity:  ConfidenceReady,
		Message:   name + " isolation storage is ready.",
	}, []string{paths.ToolStorageRootDir, paths.ToolStorageDir(toolID)}, name+" isolation storage is not ready."))
	return checks
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
			return blockedIsolationCheck(ready, blockedMessage)
		}
	}
	return ready
}

func directoryConfidenceCheck(ready ConfidenceCheck, path string, blockedMessage string) ConfidenceCheck {
	if !directoryReady(path) {
		return blockedIsolationCheck(ready, blockedMessage)
	}
	return ready
}

func blockedIsolationCheck(ready ConfidenceCheck, message string) ConfidenceCheck {
	return ConfidenceCheck{
		Component:  ConfidenceCheckIsolation,
		ToolID:     ready.ToolID,
		Severity:   ConfidenceBlocked,
		Label:      ready.Label,
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
