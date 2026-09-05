package launcher

import (
	"fmt"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
)

// Resolver selects the context for a launch request without starting any
// processes.
type Resolver struct {
	Contexts devcontext.Repository
	Projects project.Repository
}

// NewResolver creates a context resolver backed by the supplied repositories.
func NewResolver(contexts devcontext.Repository, projects project.Repository) Resolver {
	return Resolver{
		Contexts: contexts,
		Projects: projects,
	}
}

// Resolve chooses a context from explicit user input or a trusted project
// binding. When neither source can choose safely, it returns the available
// contexts and marks user selection as required.
func (r Resolver) Resolve(request LaunchRequest) (ResolutionResult, error) {
	if request.RequestedContext != nil {
		return r.resolveExplicitContext(request)
	}

	return r.resolveProjectBindingOrSelection(request)
}

func (r Resolver) resolveExplicitContext(request LaunchRequest) (ResolutionResult, error) {
	ctx, err := r.Contexts.Get(*request.RequestedContext)
	if err != nil {
		return ResolutionResult{}, err
	}
	if ctx.IsArchived() {
		return ResolutionResult{}, fmt.Errorf("%w: %s", devcontext.ErrContextArchived, ctx.ID.String())
	}

	lookup, err := r.Projects.LookupWithContextValidation(string(request.ProjectPath), request.ProjectPath, r.Contexts)
	if err != nil {
		return ResolutionResult{}, err
	}

	warnings, err := explicitContextWarnings(request, lookup)
	if err != nil {
		return ResolutionResult{}, err
	}

	return ResolutionResult{
		Context:  &ctx,
		Source:   ResolutionSourceExplicit,
		Warnings: warnings,
	}, nil
}

func (r Resolver) resolveProjectBindingOrSelection(request LaunchRequest) (ResolutionResult, error) {
	lookup, err := r.Projects.LookupWithContextValidation(string(request.ProjectPath), request.ProjectPath, r.Contexts)
	if err != nil {
		return ResolutionResult{}, err
	}

	if lookup.Bound {
		ctx, err := r.Contexts.Get(lookup.Binding.ContextID)
		if err != nil {
			return ResolutionResult{}, fmt.Errorf("load bound context %q: %w", lookup.Binding.ContextID.String(), err)
		}
		if ctx.IsArchived() {
			return r.selectionRequired(nil)
		}
		return ResolutionResult{
			Context: &ctx,
			Source:  ResolutionSourceProjectBinding,
		}, nil
	}

	warnings := danglingBindingWarnings(lookup)
	return r.selectionRequired(warnings)
}

func (r Resolver) selectionRequired(warnings []ResolutionWarning) (ResolutionResult, error) {
	contexts, err := r.Contexts.List()
	if err != nil {
		return ResolutionResult{}, err
	}

	available := make([]devcontext.Context, 0, len(contexts))
	for _, ctx := range contexts {
		if !ctx.IsArchived() {
			available = append(available, ctx)
		}
	}
	return ResolutionResult{
		Source:            ResolutionSourceUserSelection,
		Warnings:          warnings,
		SelectionRequired: true,
		AvailableContexts: available,
	}, nil
}

func danglingBindingWarnings(lookup project.BindingLookup) []ResolutionWarning {
	if !lookup.Dangling {
		return nil
	}

	return []ResolutionWarning{
		{
			Code:           WarningDanglingProjectBinding,
			ProjectPath:    lookup.ProjectPath,
			BoundContextID: lookup.MissingContextID,
			Message: fmt.Sprintf(
				"Project binding points to missing context %q; %s.",
				lookup.MissingContextID.String(),
				lookup.Recovery,
			),
		},
	}
}

func explicitContextWarnings(request LaunchRequest, lookup project.BindingLookup) ([]ResolutionWarning, error) {
	warnings := danglingBindingWarnings(lookup)
	if !lookup.Bound {
		return warnings, nil
	}

	requestedContextID := *request.RequestedContext
	if lookup.Binding.ContextID == requestedContextID {
		return warnings, nil
	}

	warning := ResolutionWarning{
		Code:               WarningContextMismatch,
		ProjectPath:        lookup.ProjectPath,
		BoundContextID:     lookup.Binding.ContextID,
		RequestedContextID: requestedContextID,
		Message: fmt.Sprintf(
			"Requested context %q differs from project binding %q for %q.",
			requestedContextID.String(),
			lookup.Binding.ContextID.String(),
			lookup.ProjectPath,
		),
	}

	switch request.MismatchConfirmation {
	case ContextMismatchAccepted:
		return append(warnings, warning), nil
	case ContextMismatchRejected:
		return nil, contextMismatchError(ErrContextMismatchRejected, warning)
	default:
		return nil, contextMismatchError(ErrContextMismatchRequiresConfirmation, warning)
	}
}

func contextMismatchError(err error, warning ResolutionWarning) error {
	return &ContextMismatchError{
		ProjectPath:        warning.ProjectPath,
		BoundContextID:     warning.BoundContextID,
		RequestedContextID: warning.RequestedContextID,
		Err:                err,
	}
}
