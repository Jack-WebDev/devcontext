package application

// GetLaunchStateRequest identifies the project the GUI is rendering.
type GetLaunchStateRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
}

// LaunchState contains everything the GUI needs to render the selector for one
// project.
type LaunchState struct {
	Project                    ProjectState                     `json:"project"`
	Contexts                   []ContextState                   `json:"contexts"`
	Binding                    ProjectBindingState              `json:"binding"`
	SelectedContextID          string                           `json:"selectedContextId,omitempty"`
	SelectionRequired          bool                             `json:"selectionRequired"`
	ResolutionSource           string                           `json:"resolutionSource,omitempty"`
	Warnings                   []ResolutionWarning              `json:"warnings,omitempty"`
	FirstRun                   bool                             `json:"firstRun"`
	ProviderCredentialSessions []ProviderCredentialSessionState `json:"providerCredentialSessions,omitempty"`
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

// CreateContextRequest asks the service to create one default context.
type CreateContextRequest struct {
	ContextID         string   `json:"contextId"`
	ImportProviderIDs []string `json:"importProviderIds,omitempty"`
}

// CreateContextResult describes a newly created context.
type CreateContextResult struct {
	Context ContextState `json:"context"`
}

// ProviderCredentialSessionState describes a detected global provider session
// using only non-secret metadata that helps the user classify the session.
type ProviderCredentialSessionState struct {
	ProviderID        string                        `json:"providerId"`
	Name              string                        `json:"name"`
	MetadataAvailable bool                          `json:"metadataAvailable"`
	Codex             *CodexCredentialSessionState  `json:"codex,omitempty"`
	Claude            *ClaudeCredentialSessionState `json:"claude,omitempty"`
}

// CodexCredentialSessionState is safe Codex identity metadata decoded from the
// id_token payload.
type CodexCredentialSessionState struct {
	Email            string `json:"email,omitempty"`
	ChatGPTPlanType  string `json:"chatgptPlanType,omitempty"`
	ChatGPTAccountID string `json:"chatgptAccountId,omitempty"`
}

// ClaudeCredentialSessionState is safe Claude identity metadata decoded from
// the global credentials file.
type ClaudeCredentialSessionState struct {
	SubscriptionType string `json:"subscriptionType,omitempty"`
	OrganizationUUID string `json:"organizationUuid,omitempty"`
	OrganizationName string `json:"organizationName,omitempty"`
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
