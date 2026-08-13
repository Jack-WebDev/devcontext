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

	// ResolutionSourceUserSelection identifies a context selected interactively
	// by the user.
	ResolutionSourceUserSelection ResolutionSource = "user_selection"
)

// WarningCode identifies a context-resolution warning.
type WarningCode string

// ResolutionWarning is a non-fatal issue discovered during context resolution.
type ResolutionWarning struct {
	Code    WarningCode
	Message string
}

// ResolutionResult represents the outcome of context resolution.
//
// Context is nil when SelectionRequired is true.
type ResolutionResult struct {
	Context           *devcontext.Context
	Source            ResolutionSource
	Warnings          []ResolutionWarning
	SelectionRequired bool
}
