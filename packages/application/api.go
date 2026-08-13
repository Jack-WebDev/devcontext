package application

// GetLaunchStateRequest identifies the project the GUI is rendering.
type GetLaunchStateRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
}

// LaunchState contains everything the GUI needs to render the selector for one
// project.
type LaunchState struct {
	Project           ProjectState        `json:"project"`
	Contexts          []ContextState      `json:"contexts"`
	Binding           ProjectBindingState `json:"binding"`
	SelectedContextID string              `json:"selectedContextId,omitempty"`
	SelectionRequired bool                `json:"selectionRequired"`
	ResolutionSource  string              `json:"resolutionSource,omitempty"`
	Warnings          []ResolutionWarning `json:"warnings,omitempty"`
	FirstRun          bool                `json:"firstRun"`
}

// ProjectState is the presentation-safe identity of one project.
type ProjectState struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ContextState is the presentation-safe identity and readiness summary for one
// configured context.
type ContextState struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Editor    EditorState       `json:"editor"`
	Providers []ProviderState   `json:"providers"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// EditorState describes the editor selected by a context without exposing
// execution details.
type EditorState struct {
	Type string `json:"type"`
}

// ProviderState describes one provider's participation and local readiness for
// a context.
type ProviderState struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	State       string `json:"state"`
	Explanation string `json:"explanation,omitempty"`
}

// ProjectBindingState describes the current remembered context for a project.
type ProjectBindingState struct {
	ProjectPath      string `json:"projectPath"`
	Bound            bool   `json:"bound"`
	ContextID        string `json:"contextId,omitempty"`
	Dangling         bool   `json:"dangling"`
	MissingContextID string `json:"missingContextId,omitempty"`
	Recovery         string `json:"recovery,omitempty"`
}

// ResolutionWarning is a presentation-safe launch warning.
type ResolutionWarning struct {
	Code               string `json:"code"`
	Message            string `json:"message"`
	ProjectPath        string `json:"projectPath,omitempty"`
	BoundContextID     string `json:"boundContextId,omitempty"`
	RequestedContextID string `json:"requestedContextId,omitempty"`
}

// LaunchProjectRequest asks the service to open one project with one context.
type LaunchProjectRequest struct {
	ProjectPath            string `json:"projectPath,omitempty"`
	ContextID              string `json:"contextId"`
	ConfirmContextMismatch bool   `json:"confirmContextMismatch"`
}

// LaunchProjectResult describes a completed editor launch.
type LaunchProjectResult struct {
	Project  ProjectState        `json:"project"`
	Context  ContextState        `json:"context"`
	Warnings []ResolutionWarning `json:"warnings,omitempty"`
}

// BindProjectRequest asks the service to remember a project-context
// association.
type BindProjectRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
	ContextID   string `json:"contextId"`
}

// UnbindProjectRequest asks the service to remove a remembered project-context
// association.
type UnbindProjectRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
}
