package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	case errors.Is(err, devcontext.ErrContextNotFound):
		return renderedError{
			Title:    "Unable to find context",
			Why:      "The requested context is not configured on this machine.",
			Recovery: "Run `devctx context list` to see available contexts, then retry with one of those IDs.",
		}
	case errors.Is(err, devcontext.ErrUnreadableContextConfig), errors.Is(err, devcontext.ErrInvalidContextConfig), errors.Is(err, devcontext.ErrContextIDMismatch):
		return renderedError{
			Title:    "Unable to read context configuration",
			Why:      "A stored context configuration is missing, malformed, or does not match its directory.",
			Recovery: "Inspect the affected context.toml file or recreate the context.",
		}
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

func redactSensitiveValues(value string) string {
	return sensitiveAssignmentPattern.ReplaceAllString(value, "$1=<redacted>")
}
