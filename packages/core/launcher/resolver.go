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

	lookup, err := r.Projects.LookupWithContextValidation(string(request.ProjectPath), request.ProjectPath, r.Contexts)
	if err != nil {
		return ResolutionResult{}, err
	}

	return ResolutionResult{
		Context:  &ctx,
		Source:   ResolutionSourceExplicit,
		Warnings: explicitContextWarnings(request, lookup),
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
		return ResolutionResult{
			Context: &ctx,
			Source:  ResolutionSourceProjectBinding,
		}, nil
	}

	warnings := danglingBindingWarnings(lookup)
	contexts, err := r.Contexts.List()
	if err != nil {
		return ResolutionResult{}, err
	}

	return ResolutionResult{
		Source:            ResolutionSourceUserSelection,
		Warnings:          warnings,
		SelectionRequired: true,
		AvailableContexts: contexts,
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

func explicitContextWarnings(request LaunchRequest, lookup project.BindingLookup) []ResolutionWarning {
	warnings := danglingBindingWarnings(lookup)
	if !lookup.Bound {
		return warnings
	}

	requestedContextID := *request.RequestedContext
	if lookup.Binding.ContextID == requestedContextID {
		return warnings
	}

	return append(warnings, ResolutionWarning{
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
	})
}
