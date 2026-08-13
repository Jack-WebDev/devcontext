package launcher

import (
	"errors"
	"fmt"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
)

var (
	// ErrContextMismatchRequiresConfirmation identifies a conflicting explicit
	// context that cannot continue until the user makes an override decision.
	ErrContextMismatchRequiresConfirmation = errors.New("context mismatch requires confirmation")

	// ErrContextMismatchRejected identifies a conflicting explicit context that
	// the user declined to continue with.
	ErrContextMismatchRejected = errors.New("context mismatch rejected")
)

// ContextMismatchError describes a conflict between an explicit context request
// and the project's stored binding.
type ContextMismatchError struct {
	ProjectPath        project.Path
	BoundContextID     devcontext.ID
	RequestedContextID devcontext.ID
	Err                error
}

func (e *ContextMismatchError) Error() string {
	return fmt.Sprintf(
		"%v: project %q is bound to context %q, requested context %q",
		e.Err,
		e.ProjectPath,
		e.BoundContextID.String(),
		e.RequestedContextID.String(),
	)
}

func (e *ContextMismatchError) Unwrap() error {
	return e.Err
}
