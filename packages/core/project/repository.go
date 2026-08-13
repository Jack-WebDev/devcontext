package project

import (
	"fmt"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
)

// BindingLookup is the result of looking up one canonical project binding.
type BindingLookup struct {
	ProjectPath Path
	Binding     Binding
	Bound       bool
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
