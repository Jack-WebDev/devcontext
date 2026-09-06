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
	if errors.Is(err, codingtool.ErrExecutableDetectionTimedOut) {
		check.Severity = ConfidenceNeedsAttention
		check.Message = name + " is taking longer than expected to check."
		check.ActionHint = "Check again, continue without detection, or review diagnostics."
		check.Retryable = true
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

// ToolConfidenceCheckWithDetection adds platform-specific recovery copy when
// an adapter can safely explain how it resolved a launchable executable.
func ToolConfidenceCheckWithDetection(toolID codingtool.ID, displayName string, detection codingtool.ExecutableDetection, err error) ConfidenceCheck {
	check := ToolConfidenceCheck(toolID, displayName, detection.Executable, err)
	switch detection.Platform {
	case "windows":
		return windowsToolConfidenceCheck(check, detection, err)
	case "linux":
		return linuxToolConfidenceCheck(check, detection, err)
	case "darwin":
		return macOSToolConfidenceCheck(check, detection, err)
	default:
		return check
	}
}

func windowsToolConfidenceCheck(check ConfidenceCheck, detection codingtool.ExecutableDetection, err error) ConfidenceCheck {
	if err != nil {
		if detection.Source == codingtool.ExecutableDetectionConfigured {
			check.Message = "The executable selected for " + check.Label + " cannot be used."
			check.ActionHint = "Select a valid " + check.Label + " executable for this context."
		} else if errors.Is(err, codingtool.ErrExecutableNotFound) {
			check.Message = check.Label + " is not installed in a standard location and its command is not on PATH."
			check.ActionHint = "Install " + check.Label + ", add its command to PATH, or select its executable for this context."
		}
		return check
	}

	switch detection.Source {
	case codingtool.ExecutableDetectionInstalled:
		check.Message = check.Label + " is installed, but its command is not on PATH. Dev Context will open the installed application."
		check.ActionHint = "Add the " + check.Label + " command to PATH if you also want to launch it from a terminal."
	case codingtool.ExecutableDetectionConfigured:
		check.Message = check.Label + " will use the executable selected for this context."
	}
	return check
}

func linuxToolConfidenceCheck(check ConfidenceCheck, detection codingtool.ExecutableDetection, err error) ConfidenceCheck {
	if err != nil {
		switch {
		case errors.Is(err, codingtool.ErrExecutableNotExecutable):
			check.Message = "Dev Context does not have permission to run " + check.Label + "."
			check.ActionHint = "Check the executable permissions or select a different " + check.Label + " executable for this context."
		case detection.Source == codingtool.ExecutableDetectionConfigured:
			check.Message = "The executable selected for " + check.Label + " was not found."
			check.ActionHint = "Select a valid " + check.Label + " executable for this context."
		case errors.Is(err, codingtool.ErrExecutableNotFound):
			check.Message = check.Label + " was not found in a standard location or on PATH."
			check.ActionHint = "Install " + check.Label + ", add its command to PATH, or select its executable for this context."
		}
		return check
	}

	switch detection.Source {
	case codingtool.ExecutableDetectionInstalled:
		check.Message = check.Label + " is installed, but its command is not on PATH. Dev Context will use the installed executable."
		check.ActionHint = "Add the " + check.Label + " command to PATH if you also want to launch it from a terminal."
	case codingtool.ExecutableDetectionConfigured:
		check.Message = check.Label + " will use the executable selected for this context."
	}
	return check
}

func macOSToolConfidenceCheck(check ConfidenceCheck, detection codingtool.ExecutableDetection, err error) ConfidenceCheck {
	if err != nil {
		switch {
		case errors.Is(err, codingtool.ErrExecutableNotExecutable):
			check.Message = "macOS could not run " + check.Label + "."
			check.ActionHint = "Check the app permissions and macOS security settings, or select a different " + check.Label + " executable for this context."
		case detection.Source == codingtool.ExecutableDetectionConfigured:
			check.Message = "The executable selected for " + check.Label + " was not found."
			check.ActionHint = "Select a valid " + check.Label + " executable for this context."
		case errors.Is(err, codingtool.ErrExecutableNotFound):
			check.Message = check.Label + " was not found in Applications or on PATH."
			check.ActionHint = "Install " + check.Label + ", add its command to PATH, or select its executable for this context."
		}
		return check
	}

	switch detection.Source {
	case codingtool.ExecutableDetectionInstalled:
		check.Message = check.Label + " is installed, but its command is not on PATH. Dev Context will use the application bundle."
		check.ActionHint = "Install the " + check.Label + " shell command if you also want to launch it from a terminal."
	case codingtool.ExecutableDetectionConfigured:
		check.Message = check.Label + " will use the executable selected for this context."
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
			Component:  ConfidenceCheckIsolation,
			ProviderID: string(integration.ID()),
			Label:      name + " isolation",
			Severity:   ConfidenceReady,
			Message:    name + " isolation storage is ready.",
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
		ProviderID: ready.ProviderID,
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
