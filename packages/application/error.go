package application

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

// ErrorCode identifies a presentation-safe application failure category.
type ErrorCode string

const (
	// ErrorCodeCanceled identifies a user-canceled application operation.
	ErrorCodeCanceled ErrorCode = "canceled"

	// ErrorCodeValidation identifies invalid user input or invalid stored data.
	ErrorCodeValidation ErrorCode = "validation_error"

	// ErrorCodeContextMismatch identifies a launch that requires explicit
	// confirmation because the selected context conflicts with the project
	// binding.
	ErrorCodeContextMismatch ErrorCode = "context_mismatch_requires_confirmation"

	// ErrorCodeLaunch identifies a failure to start the configured editor.
	ErrorCodeLaunch ErrorCode = "launch_error"

	// ErrorCodeInternal identifies an unexpected application failure.
	ErrorCodeInternal ErrorCode = "internal_error"
)

// Error is the typed error shape returned by application use cases.
type Error struct {
	Code     ErrorCode `json:"code"`
	Message  string    `json:"message"`
	Recovery string    `json:"recovery"`

	ContextMismatch *ContextMismatch `json:"contextMismatch,omitempty"`

	cause error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// NewError converts a lower-level failure into the stable application error
// shape used by framework adapters.
func NewError(err error) *Error {
	if err == nil {
		return nil
	}

	var contextMismatchError *launcher.ContextMismatchError
	var storagePermission storagePermissionDetails
	var projectPathError *project.PathError
	var missingContextError *devcontext.MissingContextError
	var contextStorageError *filesystem.ContextStorageError
	var executableNotFound *editor.ExecutableNotFoundError
	var processLaunchError *launcher.ProcessLaunchError
	switch {
	case errors.As(err, &contextMismatchError) && errors.Is(err, launcher.ErrContextMismatchRequiresConfirmation):
		return &Error{
			Code:     ErrorCodeContextMismatch,
			Message:  "Context mismatch requires confirmation.",
			Recovery: "Confirm the mismatch intentionally or choose the bound context.",
			ContextMismatch: &ContextMismatch{
				ProjectPath:        string(contextMismatchError.ProjectPath),
				BoundContextID:     contextMismatchError.BoundContextID.String(),
				RequestedContextID: contextMismatchError.RequestedContextID.String(),
			},
			cause: err,
		}
	case errors.Is(err, launcher.ErrContextMismatchRejected):
		return applicationError(ErrorCodeCanceled, "Command canceled.", "Run the action again when you are ready.", err)
	case errors.As(err, &storagePermission):
		return applicationError(
			ErrorCodeValidation,
			"Unable to access local storage.",
			storagePermissionRecovery(storagePermission),
			err,
		)
	case errors.As(err, &projectPathError):
		message, recovery := projectPathMessageAndRecovery(projectPathError)
		return applicationError(ErrorCodeValidation, message, recovery, err)
	case errors.As(err, &missingContextError):
		return applicationError(
			ErrorCodeValidation,
			fmt.Sprintf("Context %q does not exist.", missingContextError.ContextID.String()),
			missingContextRecovery(missingContextError.AvailableIDs),
			err,
		)
	case errors.As(err, &contextStorageError):
		return applicationError(
			ErrorCodeValidation,
			"Context storage is incomplete.",
			contextStorageRecovery(contextStorageError),
			err,
		)
	case errors.As(err, &executableNotFound) && executableNotFound.EditorID == editor.VSCodeID:
		return applicationError(
			ErrorCodeLaunch,
			"VS Code command not found.",
			missingVSCodeRecovery(executableNotFound.Candidates),
			err,
		)
	case errors.As(err, &processLaunchError) && errors.Is(err, launcher.ErrProcessExecutableNotFound):
		return applicationError(
			ErrorCodeLaunch,
			"VS Code command not found.",
			processExecutableRecovery(processLaunchError),
			err,
		)
	case isLaunchError(err):
		return applicationError(ErrorCodeLaunch, "Unable to launch editor.", "Check the editor command, project path, and permissions, then retry.", err)
	case isValidationError(err):
		return applicationError(ErrorCodeValidation, "Unable to complete request.", "Check the selected project and context, then retry.", err)
	default:
		return applicationError(ErrorCodeInternal, "Dev Context failed unexpectedly.", "Retry the action. If it keeps failing, include debug details in a bug report.", err)
	}
}

type storagePermissionDetails interface {
	StorageOperation() string
	StoragePath() string
}

// ContextMismatch contains presentation-safe details for a conflicting launch
// request.
type ContextMismatch struct {
	ProjectPath        string `json:"projectPath"`
	BoundContextID     string `json:"boundContextId"`
	RequestedContextID string `json:"requestedContextId"`
}

func applicationError(code ErrorCode, message string, recovery string, cause error) *Error {
	return &Error{
		Code:     code,
		Message:  message,
		Recovery: recovery,
		cause:    cause,
	}
}

func isLaunchError(err error) bool {
	return errors.Is(err, editor.ErrExecutableNotFound) ||
		errors.Is(err, editor.ErrExecutableNotExecutable) ||
		errors.Is(err, editor.ErrMissingExecutable) ||
		errors.Is(err, launcher.ErrMissingProcessExecutable) ||
		errors.Is(err, launcher.ErrProcessExecutableNotFound) ||
		errors.Is(err, launcher.ErrProcessPermissionDenied) ||
		errors.Is(err, launcher.ErrProcessWorkingDirectoryInvalid) ||
		errors.Is(err, launcher.ErrProcessStartFailed)
}

func isValidationError(err error) bool {
	return errors.Is(err, devcontext.ErrInvalidID) ||
		errors.Is(err, devcontext.ErrContextNotFound) ||
		errors.Is(err, devcontext.ErrContextAlreadyExists) ||
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
		errors.Is(err, os.ErrPermission)
}

func storagePermissionRecovery(err storagePermissionDetails) string {
	operation := strings.TrimSpace(err.StorageOperation())
	path := strings.TrimSpace(err.StoragePath())
	switch {
	case operation != "" && path != "":
		return fmt.Sprintf("Dev Context was denied permission to %s at %q. Check ownership and permissions, then retry.", operation, path)
	case path != "":
		return fmt.Sprintf("Dev Context was denied permission at %q. Check ownership and permissions, then retry.", path)
	default:
		return "Check ownership and permissions for ~/.devctx and the selected project, then retry."
	}
}

func projectPathMessageAndRecovery(err *project.PathError) (string, string) {
	path := strings.TrimSpace(err.Path)
	if path == "" {
		path = "the supplied path"
	}

	switch {
	case errors.Is(err, project.ErrProjectDirectoryNotFound):
		return "Project path does not exist.", fmt.Sprintf("Check %q and choose an existing project directory.", path)
	case errors.Is(err, project.ErrProjectPathNotDirectory):
		return "Project path is not a directory.", fmt.Sprintf("Choose a project directory instead of %q.", path)
	case errors.Is(err, project.ErrProjectDirectoryUnreadable):
		return "Project directory is not readable.", fmt.Sprintf("Check permissions for %q, then retry.", path)
	case errors.Is(err, project.ErrInvalidProjectPath):
		return "Project path is invalid.", fmt.Sprintf("Choose a valid project directory path instead of %q.", path)
	default:
		return "Unable to use project path.", fmt.Sprintf("Check %q and retry.", path)
	}
}

func missingContextRecovery(availableIDs []devcontext.ID) string {
	if len(availableIDs) == 0 {
		return "Create the context, then retry. Dev Context will not launch under a different context automatically."
	}
	return fmt.Sprintf(
		"Choose one of these configured contexts: %s. Dev Context will not launch under a different context automatically.",
		strings.Join(contextIDStrings(availableIDs), ", "),
	)
}

func contextStorageRecovery(err *filesystem.ContextStorageError) string {
	if err == nil || len(err.Missing) == 0 {
		return "Repair or recreate the selected context before launching."
	}

	paths := make([]string, 0, len(err.Missing))
	for _, missing := range err.Missing {
		paths = append(paths, fmt.Sprintf("%s: %s", missing.Kind, missing.Path))
	}
	return fmt.Sprintf(
		"Repair or recreate context %q. Missing paths: %s.",
		err.ContextID.String(),
		strings.Join(paths, "; "),
	)
}

func missingVSCodeRecovery(candidates []string) string {
	expected := "`code`"
	if len(candidates) > 0 {
		expected = "`" + strings.Join(candidates, "`, `") + "`"
	}
	return fmt.Sprintf(
		"Install the VS Code command line launcher so %s is on PATH, or set editor.executable_override in the context to a valid VS Code CLI path.",
		expected,
	)
}

func processExecutableRecovery(err *launcher.ProcessLaunchError) string {
	executable := strings.TrimSpace(string(err.Executable))
	if executable == "" {
		return missingVSCodeRecovery([]string{"code"})
	}
	return fmt.Sprintf(
		"The configured VS Code command %q was not found. Install the VS Code command line launcher, add it to PATH, or set editor.executable_override in the context.",
		executable,
	)
}

func contextIDStrings(ids []devcontext.ID) []string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = id.String()
	}
	return values
}
