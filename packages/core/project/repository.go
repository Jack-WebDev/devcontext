package project

import (
	"errors"
	"fmt"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
)

// DanglingProjectBindingRecovery is the user-facing recovery guidance for a
// binding whose target context no longer exists.
const DanglingProjectBindingRecovery = "rebind the project to an existing context or unbind the project"

// BindingLookup is the result of looking up one canonical project binding.
type BindingLookup struct {
	ProjectPath      Path
	Binding          Binding
	Bound            bool
	Dangling         bool
	MissingContextID devcontext.ID
	Recovery         string
}

// UnbindResult is the result of removing a project binding.
type UnbindResult struct {
	ProjectPath Path
	Binding     Binding
	Removed     bool
}

// Repository stores project bindings in one projects.toml file.
type Repository struct {
	path          string
	platformPaths filesystem.PlatformPaths
}

// NewRepository creates a project binding repository.
func NewRepository(path string, platformPaths filesystem.PlatformPaths) Repository {
	return Repository{
		path:          path,
		platformPaths: platformPaths,
	}
}

// List returns all stored project bindings.
func (r Repository) List() ([]Binding, error) {
	return ReadProjectBindingsFile(r.path)
}

// Lookup returns the binding for a project path or an explicit unbound result.
func (r Repository) Lookup(projectPath string, baseDir Path) (BindingLookup, error) {
	canonicalPath, err := CanonicalizePath(r.platformPaths, projectPath, baseDir)
	if err != nil {
		return BindingLookup{}, err
	}

	bindings, err := r.List()
	if err != nil {
		return BindingLookup{}, err
	}

	for _, binding := range bindings {
		if binding.ProjectPath == canonicalPath {
			return BindingLookup{
				ProjectPath: canonicalPath,
				Binding:     binding,
				Bound:       true,
			}, nil
		}
	}

	return BindingLookup{ProjectPath: canonicalPath}, nil
}

// LookupWithContextValidation returns a binding lookup and reports dangling
// bindings whose target context no longer exists.
func (r Repository) LookupWithContextValidation(projectPath string, baseDir Path, contexts devcontext.Repository) (BindingLookup, error) {
	lookup, err := r.Lookup(projectPath, baseDir)
	if err != nil {
		return BindingLookup{}, err
	}
	if !lookup.Bound {
		return lookup, nil
	}

	if _, err := contexts.Get(lookup.Binding.ContextID); err != nil {
		if errors.Is(err, devcontext.ErrContextNotFound) {
			lookup.Bound = false
			lookup.Dangling = true
			lookup.MissingContextID = lookup.Binding.ContextID
			lookup.Recovery = DanglingProjectBindingRecovery
			return lookup, nil
		}
		return BindingLookup{}, fmt.Errorf("validate binding context %q: %w", lookup.Binding.ContextID.String(), err)
	}

	return lookup, nil
}

// Bind validates and persists one project-to-context association. Existing
// bindings for the same project path are intentionally replaced.
func (r Repository) Bind(projectPath string, baseDir Path, contextID devcontext.ID, contexts devcontext.Repository, createdAt time.Time) (Binding, error) {
	canonicalPath, err := CanonicalizePath(r.platformPaths, projectPath, baseDir)
	if err != nil {
		return Binding{}, err
	}
	if err := ValidateProjectDirectory(canonicalPath); err != nil {
		return Binding{}, err
	}
	if _, err := contexts.Get(contextID); err != nil {
		return Binding{}, fmt.Errorf("validate target context %q: %w", contextID.String(), err)
	}

	bindings, err := r.List()
	if err != nil {
		return Binding{}, err
	}

	binding := Binding{
		ProjectPath: canonicalPath,
		ContextID:   contextID,
		CreatedAt:   createdAt.UTC(),
	}

	replaced := false
	for i := range bindings {
		if bindings[i].ProjectPath == canonicalPath {
			bindings[i] = binding
			replaced = true
			break
		}
	}
	if !replaced {
		bindings = append(bindings, binding)
	}

	if err := WriteProjectBindingsFile(r.path, bindings); err != nil {
		return Binding{}, err
	}
	return binding, nil
}

// Unbind removes the stored association for one project path. Removing an
// already-unbound project is idempotent.
func (r Repository) Unbind(projectPath string, baseDir Path) (UnbindResult, error) {
	canonicalPath, err := CanonicalizePath(r.platformPaths, projectPath, baseDir)
	if err != nil {
		return UnbindResult{}, err
	}

	bindings, err := r.List()
	if err != nil {
		return UnbindResult{}, err
	}

	remaining := make([]Binding, 0, len(bindings))
	var removed Binding
	for _, binding := range bindings {
		if binding.ProjectPath == canonicalPath {
			removed = binding
			continue
		}
		remaining = append(remaining, binding)
	}

	result := UnbindResult{
		ProjectPath: canonicalPath,
		Binding:     removed,
		Removed:     removed.ProjectPath != "",
	}
	if !result.Removed {
		return result, nil
	}

	if err := WriteProjectBindingsFile(r.path, remaining); err != nil {
		return UnbindResult{}, err
	}
	return result, nil
}

// UnbindContext removes every project binding for a context in one atomic
// project-bindings write and returns the affected bindings.
func (r Repository) UnbindContext(contextID devcontext.ID) ([]Binding, error) {
	bindings, err := r.List()
	if err != nil {
		return nil, err
	}
	remaining := make([]Binding, 0, len(bindings))
	removed := make([]Binding, 0)
	for _, binding := range bindings {
		if binding.ContextID == contextID {
			removed = append(removed, binding)
			continue
		}
		remaining = append(remaining, binding)
	}
	if len(removed) == 0 {
		return removed, nil
	}
	if err := WriteProjectBindingsFile(r.path, remaining); err != nil {
		return nil, err
	}
	return removed, nil
}
