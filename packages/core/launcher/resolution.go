package launcher

import devcontext "devctx/packages/core/context"

// ResolutionSource explains why a context was selected.
type ResolutionSource string

const (
	// ResolutionSourceExplicit identifies a context requested directly by the user.
	ResolutionSourceExplicit ResolutionSource = "explicit"

	// ResolutionSourceProjectBinding identifies a context selected from a stored
	// project binding.
	ResolutionSourceProjectBinding ResolutionSource = "project_binding"

	// ResolutionSourceUserSelection identifies an interactive user-selection
	// path, including results that still require the user to choose.
	ResolutionSourceUserSelection ResolutionSource = "user_selection"
)

// WarningCode identifies a context-resolution warning.
type WarningCode string

const (
	// WarningDanglingProjectBinding identifies a project binding whose stored
	// context no longer exists.
	WarningDanglingProjectBinding WarningCode = "dangling_project_binding"
)

// ResolutionWarning is a non-fatal issue discovered during context resolution.
type ResolutionWarning struct {
	Code    WarningCode
	Message string
}

// ResolutionResult represents the outcome of context resolution.
//
// Context is nil when SelectionRequired is true. AvailableContexts contains the
// choices the caller can present to the user in that state.
type ResolutionResult struct {
	Context           *devcontext.Context
	Source            ResolutionSource
	Warnings          []ResolutionWarning
	SelectionRequired bool
	AvailableContexts []devcontext.Context
}
