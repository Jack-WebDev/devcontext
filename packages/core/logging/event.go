package logging

import (
	"errors"
	"os"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

// EventName identifies one approved troubleshooting event.
type EventName string

const (
	EventContextResolution     EventName = "context_resolution"
	EventLaunchSucceeded       EventName = "launch_succeeded"
	EventLaunchMissingEditor   EventName = "launch_missing_editor"
	EventLaunchConfigError     EventName = "launch_configuration_error"
	EventLaunchProviderMissing EventName = "launch_provider_missing"
	EventLaunchProcessFailure  EventName = "launch_process_failure"
	EventContextCreated        EventName = "context_created"
	EventProviderConnected     EventName = "provider_connected"
	EventProviderReset         EventName = "provider_reset"
	EventRepairCompleted       EventName = "repair_completed"
	EventProjectBindingChanged EventName = "project_binding_changed"
	EventEnvironmentStopped    EventName = "environment_stopped"
)

// ErrorCategory identifies a bounded failure class without requiring callers to
// persist raw errors.
type ErrorCategory string

const (
	ErrorCategoryConfiguration ErrorCategory = "configuration"
	ErrorCategoryContext       ErrorCategory = "context"
	ErrorCategoryTool          ErrorCategory = "editor"
	ErrorCategoryPermission    ErrorCategory = "permission"
	ErrorCategoryProcess       ErrorCategory = "process"
	ErrorCategoryProject       ErrorCategory = "project"
	ErrorCategoryProvider      ErrorCategory = "provider"
	ErrorCategorySelection     ErrorCategory = "selection_required"
	ErrorCategoryUnknown       ErrorCategory = "unknown"
)

// Event is the full allowlisted local log record. Add fields here deliberately;
// callers cannot attach arbitrary structured data.
type Event struct {
	Name             EventName     `json:"event"`
	Timestamp        time.Time     `json:"timestamp"`
	ProjectPath      string        `json:"project_path,omitempty"`
	ContextID        string        `json:"context_id,omitempty"`
	ToolID           string        `json:"editor_id,omitempty"`
	ResolutionSource string        `json:"resolution_source,omitempty"`
	ErrorCategory    ErrorCategory `json:"error_category,omitempty"`
	Error            string        `json:"error,omitempty"`
}

// EventInput stores approved event construction inputs.
type EventInput struct {
	Name             EventName
	Timestamp        time.Time
	ProjectPath      string
	ContextID        string
	ToolID           string
	ResolutionSource string
	Err              error
	KnownEnvironment []string
}

// NewEvent creates an allowlisted event and sanitizes any diagnostic error text
// before it is stored on the event.
func NewEvent(input EventInput) Event {
	event := Event{
		Name:             input.Name,
		Timestamp:        input.Timestamp,
		ProjectPath:      input.ProjectPath,
		ContextID:        input.ContextID,
		ToolID:           input.ToolID,
		ResolutionSource: input.ResolutionSource,
	}

	if input.Err != nil {
		event.ErrorCategory = CategoryForError(input.Err)
		event.Error = SanitizeError(input.Err, input.KnownEnvironment)
	}

	return event
}

// CategoryForError maps internal errors to stable diagnostic categories.
func CategoryForError(err error) ErrorCategory {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, config.ErrInvalidGlobalConfig), errors.Is(err, config.ErrUnsupportedSchemaVersion):
		return ErrorCategoryConfiguration
	case errors.Is(err, filesystem.ErrStoragePermissionDenied):
		return ErrorCategoryPermission
	case errors.Is(err, devcontext.ErrInvalidID), errors.Is(err, devcontext.ErrContextNotFound),
		errors.Is(err, devcontext.ErrUnreadableContextConfig), errors.Is(err, devcontext.ErrInvalidContextConfig),
		errors.Is(err, devcontext.ErrContextIDMismatch), errors.Is(err, filesystem.ErrContextStorageIncomplete),
		errors.Is(err, launcher.ErrLaunchSelectionRequired):
		if errors.Is(err, launcher.ErrLaunchSelectionRequired) {
			return ErrorCategorySelection
		}
		return ErrorCategoryContext
	case errors.Is(err, project.ErrProjectDirectoryNotFound), errors.Is(err, project.ErrProjectPathNotDirectory),
		errors.Is(err, project.ErrProjectDirectoryUnreadable), errors.Is(err, project.ErrInvalidProjectPath),
		errors.Is(err, project.ErrInvalidProjectBindings), errors.Is(err, project.ErrDuplicateProjectBinding):
		return ErrorCategoryProject
	case errors.Is(err, codingtool.ErrExecutableNotFound), errors.Is(err, codingtool.ErrExecutableNotExecutable),
		errors.Is(err, codingtool.ErrMissingExecutable), errors.Is(err, launcher.ErrMissingProcessExecutable):
		return ErrorCategoryTool
	case errors.Is(err, launcher.ErrProcessExecutableNotFound), errors.Is(err, launcher.ErrProcessStartFailed),
		errors.Is(err, launcher.ErrProcessWorkingDirectoryInvalid):
		return ErrorCategoryProcess
	case errors.Is(err, launcher.ErrProcessPermissionDenied), isPermissionError(err):
		return ErrorCategoryPermission
	default:
		return ErrorCategoryUnknown
	}
}

// LaunchEventNameForError maps a launch failure to one approved event name.
func LaunchEventNameForError(err error) EventName {
	switch {
	case errors.Is(err, codingtool.ErrExecutableNotFound), errors.Is(err, codingtool.ErrExecutableNotExecutable),
		errors.Is(err, codingtool.ErrMissingExecutable), errors.Is(err, launcher.ErrMissingProcessExecutable):
		return EventLaunchMissingEditor
	case errors.Is(err, launcher.ErrProcessExecutableNotFound), errors.Is(err, launcher.ErrProcessPermissionDenied),
		errors.Is(err, launcher.ErrProcessWorkingDirectoryInvalid), errors.Is(err, launcher.ErrProcessStartFailed):
		return EventLaunchProcessFailure
	case CategoryForError(err) == ErrorCategoryConfiguration:
		return EventLaunchConfigError
	default:
		return EventContextResolution
	}
}

func isPermissionError(err error) bool {
	return errors.Is(err, filesystem.ErrUserHomeUnavailable) ||
		os.IsPermission(err) ||
		errors.Is(err, os.ErrPermission)
}
