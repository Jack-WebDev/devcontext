package cli

import (
	"errors"

	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

var (
	// ErrCanceled identifies a command canceled by the user.
	ErrCanceled = errors.New("CLI command canceled")

	// ErrLaunchFailed identifies a failed process launch.
	ErrLaunchFailed = errors.New("CLI launch failed")
)

// ExitCode is the process outcome contract for CLI callers.
type ExitCode int

const (
	// ExitSuccess identifies a completed command.
	ExitSuccess ExitCode = 0

	// ExitInternalError identifies an unexpected application failure.
	ExitInternalError ExitCode = 1

	// ExitUsageError identifies invalid command syntax or unsupported command
	// usage.
	ExitUsageError ExitCode = 2

	// ExitValidationError identifies invalid user or stored data.
	ExitValidationError ExitCode = 3

	// ExitCanceled identifies a user-canceled operation.
	ExitCanceled ExitCode = 4

	// ExitLaunchFailure identifies a failure to launch the target editor.
	ExitLaunchFailure ExitCode = 5
)

// ExitCodeForError maps typed CLI and application errors to stable outcomes.
func ExitCodeForError(err error) ExitCode {
	switch {
	case err == nil:
		return ExitSuccess
	case errors.Is(err, ErrCanceled), errors.Is(err, launcher.ErrContextMismatchRejected):
		return ExitCanceled
	case errors.Is(err, ErrLaunchFailed),
		errors.Is(err, editor.ErrExecutableNotFound),
		errors.Is(err, editor.ErrExecutableNotExecutable),
		errors.Is(err, editor.ErrMissingExecutable),
		errors.Is(err, launcher.ErrMissingProcessExecutable),
		errors.Is(err, launcher.ErrProcessExecutableNotFound),
		errors.Is(err, launcher.ErrProcessPermissionDenied),
		errors.Is(err, launcher.ErrProcessWorkingDirectoryInvalid),
		errors.Is(err, launcher.ErrProcessStartFailed):
		return ExitLaunchFailure
	case isValidationError(err):
		return ExitValidationError
	case errors.Is(err, ErrInvalidCommand), errors.Is(err, ErrUnknownCommand):
		return ExitUsageError
	default:
		return ExitInternalError
	}
}

func isValidationError(err error) bool {
	return errors.Is(err, devcontext.ErrInvalidID) ||
		errors.Is(err, devcontext.ErrContextNotFound) ||
		errors.Is(err, devcontext.ErrUnreadableContextConfig) ||
		errors.Is(err, devcontext.ErrInvalidContextConfig) ||
		errors.Is(err, devcontext.ErrContextIDMismatch) ||
		errors.Is(err, config.ErrInvalidGlobalConfig) ||
		errors.Is(err, config.ErrUnsupportedSchemaVersion) ||
		errors.Is(err, filesystem.ErrContextStorageIncomplete) ||
		errors.Is(err, filesystem.ErrStoragePermissionDenied) ||
		errors.Is(err, filesystem.ErrUserHomeUnavailable) ||
		errors.Is(err, launcher.ErrContextMismatchRequiresConfirmation) ||
		errors.Is(err, launcher.ErrLaunchSelectionRequired) ||
		errors.Is(err, project.ErrInvalidProjectPath) ||
		errors.Is(err, project.ErrProjectDirectoryNotFound) ||
		errors.Is(err, project.ErrProjectPathNotDirectory) ||
		errors.Is(err, project.ErrProjectDirectoryUnreadable) ||
		errors.Is(err, project.ErrInvalidProjectBindings) ||
		errors.Is(err, project.ErrDuplicateProjectBinding) ||
		isPermissionError(err)
}
