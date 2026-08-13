package application

import (
	"errors"
	"os"

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
	case isLaunchError(err):
		return applicationError(ErrorCodeLaunch, "Unable to launch editor.", "Check the editor command, project path, and permissions, then retry.", err)
	case isValidationError(err):
		return applicationError(ErrorCodeValidation, "Unable to complete request.", "Check the selected project and context, then retry.", err)
	default:
		return applicationError(ErrorCodeInternal, "Dev Context failed unexpectedly.", "Retry the action. If it keeps failing, include debug details in a bug report.", err)
	}
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
