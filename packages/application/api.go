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
	Confidence                 *LaunchConfidenceState           `json:"confidence,omitempty"`
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
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	Tool           ToolState             `json:"tool"`
	AvailableTools []ToolOption          `json:"availableTools"`
	Providers      []ProviderState       `json:"providers"`
	Confidence     LaunchConfidenceState `json:"confidence"`
	Metadata       map[string]string     `json:"metadata,omitempty"`
}

// ToolState describes the coding tool selected by a context, including
// presentation-safe readiness and recovery guidance.
type ToolState struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Status     LaunchConfidenceStatus `json:"status"`
	Message    string                 `json:"message"`
	ActionHint string                 `json:"actionHint,omitempty"`
}

// ToolOption describes a registered coding tool that can be selected for a
// context without exposing its executable or storage configuration.
type ToolOption struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ProviderState describes one provider's participation and local readiness for
// a context.
type ProviderState struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Enabled     bool                   `json:"enabled"`
	State       ProviderReadinessState `json:"state"`
	Explanation string                 `json:"explanation,omitempty"`
	ActionHint  string                 `json:"actionHint,omitempty"`
	Identity    ProviderIdentityState  `json:"identity"`
}

// ProviderReadinessState is the UI-facing provider readiness vocabulary.
// It intentionally maps core provider.StatusConfigured to ready so API
// consumers do not need to understand storage-oriented provider wording.
type ProviderReadinessState string

const (
	// ProviderReadinessReady means the provider has local state for this
	// context.
	ProviderReadinessReady ProviderReadinessState = "ready"

	// ProviderReadinessNotConfigured means provider storage exists but does not
	// appear initialized.
	ProviderReadinessNotConfigured ProviderReadinessState = "not_configured"

	// ProviderReadinessDirectoryMissing means context-owned provider storage is
	// absent.
	ProviderReadinessDirectoryMissing ProviderReadinessState = "directory_missing"

	// ProviderReadinessUnavailable means local readiness could not be
	// determined.
	ProviderReadinessUnavailable ProviderReadinessState = "unavailable"
)

// Valid reports whether state is one of the bounded API provider readiness
// states.
func (s ProviderReadinessState) Valid() bool {
	switch s {
	case ProviderReadinessReady, ProviderReadinessNotConfigured, ProviderReadinessDirectoryMissing, ProviderReadinessUnavailable:
		return true
	default:
		return false
	}
}

// ProviderIdentityStatus identifies whether provider account identity is safe
// to display.
type ProviderIdentityStatus string

const (
	// ProviderIdentityVerified means identity metadata was verified from local
	// provider state and is safe to display.
	ProviderIdentityVerified ProviderIdentityStatus = "verified"

	// ProviderIdentityUnavailable means provider state exists, but account
	// identity could not be verified.
	ProviderIdentityUnavailable ProviderIdentityStatus = "unavailable"

	// ProviderIdentityNone means no provider account identity is present for
	// this context.
	ProviderIdentityNone ProviderIdentityStatus = "none"

	// ProviderIdentityMismatchEvidence means Dev Context has explicit evidence
	// that the provider account may not match the intended context identity.
	ProviderIdentityMismatchEvidence ProviderIdentityStatus = "mismatch_evidence"
)

// Valid reports whether status is one of the bounded API provider identity
// states.
func (s ProviderIdentityStatus) Valid() bool {
	switch s {
	case ProviderIdentityVerified, ProviderIdentityUnavailable, ProviderIdentityNone, ProviderIdentityMismatchEvidence:
		return true
	default:
		return false
	}
}

// ProviderIdentityState contains only safe provider account identity metadata.
// Verified provider-specific fields are added by provider extraction phases;
// until then, the status tells the UI not to guess.
type ProviderIdentityState struct {
	Status  ProviderIdentityStatus  `json:"status"`
	Message string                  `json:"message,omitempty"`
	Fields  []ProviderMetadataField `json:"fields,omitempty"`
}

// ProviderMetadataField contains one safe provider-supplied value for display.
type ProviderMetadataField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// PreflightLaunchProjectRequest asks the service to check launch readiness for
// one project and context without starting the editor process.
type PreflightLaunchProjectRequest struct {
	ProjectPath            string `json:"projectPath,omitempty"`
	ContextID              string `json:"contextId"`
	ConfirmContextMismatch bool   `json:"confirmContextMismatch"`
}

// PreflightLaunchProjectResult describes launch readiness before an editor
// process is started.
type PreflightLaunchProjectResult struct {
	Project    ProjectState          `json:"project"`
	Context    ContextState          `json:"context"`
	Confidence LaunchConfidenceState `json:"confidence"`
	Warnings   []ResolutionWarning   `json:"warnings,omitempty"`
}

// LaunchConfidenceState summarizes backend-owned launch readiness for the
// selected or recommended context.
type LaunchConfidenceState struct {
	ContextID string                  `json:"contextId"`
	Status    LaunchConfidenceStatus  `json:"status"`
	Checks    []LaunchConfidenceCheck `json:"checks"`
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
	ProviderID        string                  `json:"providerId"`
	Name              string                  `json:"name"`
	MetadataAvailable bool                    `json:"metadataAvailable"`
	Fields            []ProviderMetadataField `json:"fields,omitempty"`
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
