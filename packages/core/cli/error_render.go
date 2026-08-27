package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

type renderedError struct {
	Title    string
	Why      string
	Recovery string
}

var sensitiveAssignmentPattern = regexp.MustCompile(`(?i)\b([A-Z0-9_]*(TOKEN|SECRET|PASSWORD|KEY)[A-Z0-9_]*)=([^\s]+)`)

// RenderError returns a user-facing CLI error with what failed, why it failed,
// and what to do next. Debug output includes sanitized internal details.
func RenderError(err error, debug bool) string {
	if err == nil {
		return ""
	}

	rendered := classifyError(err)
	var builder strings.Builder
	builder.WriteString(rendered.Title)
	builder.WriteString("\n\n")
	builder.WriteString(rendered.Why)
	builder.WriteString("\n\n")
	builder.WriteString("Next step:\n")
	builder.WriteString(rendered.Recovery)
	builder.WriteString("\n")

	if debug {
		builder.WriteString("\nDebug:\n")
		builder.WriteString(redactSensitiveValues(err.Error()))
		builder.WriteString("\n")
	}

	return builder.String()
}

func classifyError(err error) renderedError {
	var globalConfigFileError *config.GlobalConfigFileError
	var contextMismatchError *launcher.ContextMismatchError
	var storagePermission storagePermissionDetails
	var projectPathError *project.PathError
	var missingContextError *devcontext.MissingContextError
	var contextStorageError *filesystem.ContextStorageError
	var executableNotFound *codingtool.ExecutableNotFoundError
	var processLaunchError *launcher.ProcessLaunchError
	switch {
	case errors.As(err, &globalConfigFileError):
		return renderedError{
			Title:    "Unable to read Dev Context configuration",
			Why:      fmt.Sprintf("The global configuration file at %q is invalid: %s.", globalConfigFileError.Path, globalConfigFileError.Cause),
			Recovery: globalConfigFileError.Recovery + ".",
		}
	case errors.As(err, &contextMismatchError) && errors.Is(err, launcher.ErrContextMismatchRequiresConfirmation):
		return renderedError{
			Title: "Context mismatch requires confirmation",
			Why: fmt.Sprintf(
				"The project at %q is bound to context %q, but the request selected context %q.",
				contextMismatchError.ProjectPath,
				contextMismatchError.BoundContextID.String(),
				contextMismatchError.RequestedContextID.String(),
			),
			Recovery: "Confirm the mismatch intentionally or rerun with the bound context.",
		}
	case errors.As(err, &contextMismatchError) && errors.Is(err, launcher.ErrContextMismatchRejected):
		return renderedError{
			Title: "Command canceled",
			Why: fmt.Sprintf(
				"The context mismatch for project %q was not approved.",
				contextMismatchError.ProjectPath,
			),
			Recovery: "Run the command again when you are ready.",
		}
	case errors.Is(err, launcher.ErrLaunchSelectionRequired):
		return renderedError{
			Title:    "Context selection required",
			Why:      "No explicit context or trusted project binding selected a context.",
			Recovery: "Choose a context in the selector or rerun with `--context <id>`.",
		}
	case errors.As(err, &storagePermission):
		return renderedError{
			Title:    "Unable to access local storage",
			Why:      storagePermissionWhy(storagePermission),
			Recovery: "Check ownership and permissions for the affected path, then retry.",
		}
	case errors.Is(err, ErrInvalidCommand), errors.Is(err, ErrUnknownCommand):
		return renderedError{
			Title:    "Unable to parse command",
			Why:      "The command line does not match a supported Dev Context command shape.",
			Recovery: "Check the command and run it again.",
		}
	case errors.Is(err, devcontext.ErrInvalidID):
		return renderedError{
			Title:    "Unable to use context ID",
			Why:      "The context ID is not filesystem-safe.",
			Recovery: "Use lowercase letters, digits, and hyphens only, starting and ending with a letter or digit.",
		}
	case errors.As(err, &missingContextError):
		return renderedError{
			Title:    "Unable to find context",
			Why:      fmt.Sprintf("Context %q is not configured on this machine.", missingContextError.ContextID.String()),
			Recovery: missingContextRecovery(missingContextError.AvailableIDs),
		}
	case errors.Is(err, devcontext.ErrContextNotFound):
		return renderedError{
			Title:    "Unable to find context",
			Why:      "The requested context is not configured on this machine.",
			Recovery: "Run `devctx context list` to see available contexts, then retry with one of those IDs.",
		}
	case errors.As(err, &contextStorageError):
		return renderedError{
			Title:    "Context storage is incomplete",
			Why:      contextStorageWhy(contextStorageError),
			Recovery: "Repair or recreate the context before launching. Dev Context will not recreate incomplete context storage automatically.",
		}
	case errors.Is(err, devcontext.ErrUnreadableContextConfig), errors.Is(err, devcontext.ErrInvalidContextConfig), errors.Is(err, devcontext.ErrContextIDMismatch):
		return renderedError{
			Title:    "Unable to read context configuration",
			Why:      "A stored context configuration is missing, malformed, or does not match its directory.",
			Recovery: "Inspect the affected context.toml file or recreate the context.",
		}
	case errors.As(err, &projectPathError):
		return projectPathRenderedError(projectPathError)
	case errors.Is(err, project.ErrProjectDirectoryNotFound):
		return renderedError{
			Title:    "Unable to open project",
			Why:      "The project path does not exist.",
			Recovery: "Check the path and run the command from an existing project directory.",
		}
	case errors.Is(err, project.ErrProjectPathNotDirectory):
		return renderedError{
			Title:    "Unable to open project",
			Why:      "The project path points to a file, not a directory.",
			Recovery: "Pass a project directory or run the command from inside one.",
		}
	case errors.Is(err, project.ErrProjectDirectoryUnreadable):
		return renderedError{
			Title:    "Unable to open project",
			Why:      "Dev Context cannot read the project directory.",
			Recovery: "Check directory permissions and try again.",
		}
	case errors.Is(err, project.ErrInvalidProjectPath):
		return renderedError{
			Title:    "Unable to use project path",
			Why:      "The project path cannot be resolved to a valid absolute path.",
			Recovery: "Pass a valid project directory path.",
		}
	case errors.Is(err, project.ErrInvalidProjectBindings), errors.Is(err, project.ErrDuplicateProjectBinding):
		return renderedError{
			Title:    "Unable to read project bindings",
			Why:      "The project binding file is malformed or contains duplicate project entries.",
			Recovery: "Fix projects.toml or move it aside so Dev Context can create a fresh binding file.",
		}
	case errors.Is(err, config.ErrInvalidGlobalConfig), errors.Is(err, config.ErrUnsupportedSchemaVersion):
		return renderedError{
			Title:    "Unable to read Dev Context configuration",
			Why:      "The global configuration is malformed or uses an unsupported schema version.",
			Recovery: "Fix config.toml or move it aside so Dev Context can create a new default configuration.",
		}
	case errors.Is(err, filesystem.ErrUserHomeUnavailable):
		return renderedError{
			Title:    "Unable to locate Dev Context storage",
			Why:      "The current user's home directory could not be determined.",
			Recovery: "Check the user environment and try again.",
		}
	case errors.As(err, &executableNotFound) && executableNotFound.EditorID == codingtool.VSCodeID:
		return renderedError{
			Title:    "VS Code command not found",
			Why:      missingVSCodeWhy(executableNotFound.Candidates),
			Recovery: missingVSCodeRecovery(executableNotFound.Candidates),
		}
	case errors.Is(err, codingtool.ErrExecutableNotFound):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "Dev Context could not find the configured editor executable.",
			Recovery: "Install the editor command, add it to PATH, or configure a valid editor executable.",
		}
	case errors.Is(err, codingtool.ErrExecutableNotExecutable):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "The configured editor executable is not usable.",
			Recovery: "Configure an executable editor path or install the editor command on PATH.",
		}
	case errors.Is(err, codingtool.ErrMissingExecutable), errors.Is(err, launcher.ErrMissingProcessExecutable):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "No editor executable was resolved for the selected context.",
			Recovery: "Configure an editor executable or install the editor command on PATH.",
		}
	case errors.As(err, &processLaunchError) && errors.Is(err, launcher.ErrProcessExecutableNotFound):
		return renderedError{
			Title:    "VS Code command not found",
			Why:      processExecutableWhy(processLaunchError),
			Recovery: "Install the VS Code command line launcher, add it to PATH, or set codingtool.executable_override in the context to a valid VS Code CLI path.",
		}
	case errors.Is(err, launcher.ErrProcessExecutableNotFound):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "Dev Context could not find the configured editor executable.",
			Recovery: "Install the editor command, add it to PATH, or configure a valid editor executable.",
		}
	case errors.Is(err, launcher.ErrProcessPermissionDenied):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "The operating system denied permission to start the editor process.",
			Recovery: "Check executable, project, and Dev Context storage permissions, then retry.",
		}
	case errors.Is(err, launcher.ErrProcessWorkingDirectoryInvalid):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "The editor working directory is missing or is not a directory.",
			Recovery: "Check the project path and run Dev Context from an existing project directory.",
		}
	case errors.Is(err, launcher.ErrProcessStartFailed):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "The operating system could not start the editor process.",
			Recovery: "Check the editor command and project path, then retry.",
		}
	case isPermissionError(err):
		return renderedError{
			Title:    "Unable to access local storage",
			Why:      "The operating system denied permission to a required file or directory.",
			Recovery: "Check ownership and permissions for the project and ~/.devctx paths, then retry.",
		}
	case errors.Is(err, ErrCanceled):
		return renderedError{
			Title:    "Command canceled",
			Why:      "The operation was canceled before Dev Context made changes.",
			Recovery: "Run the command again when you are ready.",
		}
	case errors.Is(err, ErrLaunchFailed):
		return renderedError{
			Title:    "Unable to launch editor",
			Why:      "Dev Context could not start the configured editor process.",
			Recovery: "Check that the editor command is installed and available, then retry.",
		}
	default:
		return renderedError{
			Title:    "Dev Context command failed",
			Why:      "An unexpected error occurred.",
			Recovery: "Retry the command. If it keeps failing, rerun with debug output and include the sanitized details in a bug report.",
		}
	}
}

type storagePermissionDetails interface {
	StorageOperation() string
	StoragePath() string
}

func redactSensitiveValues(value string) string {
	return sensitiveAssignmentPattern.ReplaceAllString(value, "$1=<redacted>")
}

func storagePermissionWhy(err storagePermissionDetails) string {
	operation := strings.TrimSpace(err.StorageOperation())
	path := strings.TrimSpace(err.StoragePath())
	switch {
	case operation != "" && path != "":
		return fmt.Sprintf("The operating system denied permission to %s at %q.", operation, path)
	case path != "":
		return fmt.Sprintf("The operating system denied permission at %q.", path)
	default:
		return "The operating system denied permission to a required file or directory."
	}
}

func missingContextRecovery(availableIDs []devcontext.ID) string {
	if len(availableIDs) == 0 {
		return "Create the context or run `devctx context list` to see available contexts, then retry."
	}
	return fmt.Sprintf("Retry with one of these context IDs: %s.", strings.Join(contextIDStrings(availableIDs), ", "))
}

func contextStorageWhy(err *filesystem.ContextStorageError) string {
	if err == nil || len(err.Missing) == 0 {
		return "The selected context is missing required storage directories."
	}

	entries := make([]string, 0, len(err.Missing))
	for _, missing := range err.Missing {
		kind := missingDirectoryLabel(missing)
		if missing.Reason == "" {
			entries = append(entries, fmt.Sprintf("%s %q", kind, missing.Path))
			continue
		}
		entries = append(entries, fmt.Sprintf("%s %q (%s)", kind, missing.Path, missing.Reason))
	}
	return fmt.Sprintf("Context %q is missing required storage directories: %s.", err.ContextID.String(), strings.Join(entries, "; "))
}

func missingDirectoryLabel(missing filesystem.MissingContextDirectory) string {
	kind := string(missing.Kind)
	if missing.ProviderID == "" {
		return kind
	}
	kind += ":" + missing.ProviderID
	if missing.ProviderDisplayName != "" {
		kind += " (" + missing.ProviderDisplayName + ")"
	}
	return kind
}

func projectPathRenderedError(err *project.PathError) renderedError {
	path := strings.TrimSpace(err.Path)
	if path == "" {
		path = "the supplied path"
	}

	switch {
	case errors.Is(err, project.ErrProjectDirectoryNotFound):
		return renderedError{
			Title:    "Unable to open project",
			Why:      fmt.Sprintf("The project path %q does not exist.", path),
			Recovery: "Create the directory or pass an existing project directory.",
		}
	case errors.Is(err, project.ErrProjectPathNotDirectory):
		return renderedError{
			Title:    "Unable to open project",
			Why:      fmt.Sprintf("The project path %q points to a file, not a directory.", path),
			Recovery: "Pass a project directory or run the command from inside one.",
		}
	case errors.Is(err, project.ErrProjectDirectoryUnreadable):
		return renderedError{
			Title:    "Unable to open project",
			Why:      fmt.Sprintf("Dev Context cannot read the project directory %q.", path),
			Recovery: "Check directory permissions and try again.",
		}
	case errors.Is(err, project.ErrInvalidProjectPath):
		return renderedError{
			Title:    "Unable to use project path",
			Why:      fmt.Sprintf("The project path %q cannot be resolved to a valid absolute path.", path),
			Recovery: "Pass a valid project directory path.",
		}
	default:
		return renderedError{
			Title:    "Unable to use project path",
			Why:      fmt.Sprintf("Dev Context could not use project path %q.", path),
			Recovery: "Check the project path and try again.",
		}
	}
}

func missingVSCodeWhy(candidates []string) string {
	if len(candidates) == 0 {
		return "Dev Context expected the VS Code CLI command `code` but could not find it on PATH."
	}
	return fmt.Sprintf("Dev Context expected the VS Code CLI command on PATH and checked: `%s`.", strings.Join(candidates, "`, `"))
}

func missingVSCodeRecovery(candidates []string) string {
	expected := "`code`"
	if len(candidates) > 0 {
		expected = "`" + strings.Join(candidates, "`, `") + "`"
	}
	return fmt.Sprintf("Install the VS Code command line launcher so %s is on PATH, or set codingtool.executable_override in the context to a valid VS Code CLI path.", expected)
}

func processExecutableWhy(err *launcher.ProcessLaunchError) string {
	executable := strings.TrimSpace(string(err.Executable))
	if executable == "" {
		return "Dev Context could not find the configured VS Code command."
	}
	return fmt.Sprintf("Dev Context could not find the configured VS Code command %q.", executable)
}

func contextIDStrings(ids []devcontext.ID) []string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = id.String()
	}
	return values
}
