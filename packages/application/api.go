package application

import "time"

// GetLaunchStateRequest identifies the project the GUI is rendering.
type GetLaunchStateRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
}

// SettingsState contains user-configurable application behavior.
type SettingsState struct {
	CloseAfterLaunch   bool `json:"closeAfterLaunch"`
	LaunchVerification bool `json:"launchVerification"`
	RememberProjects   bool `json:"rememberProjects"`
	TrayEnabled        bool `json:"trayEnabled"`
}

// UpdateSettingsRequest replaces the supported application preferences.
type UpdateSettingsRequest SettingsState

// GetHomeDashboardRequest identifies the project represented by the Home
// dashboard.
type GetHomeDashboardRequest struct {
	ProjectPath string `json:"projectPath,omitempty"`
}

// HomeDashboardState contains backend-owned data for the Home screen. Recent
// projects, running environments, and activity are intentionally empty until
// their dedicated persistence contracts are introduced.
type HomeDashboardState struct {
	Project        ProjectState             `json:"project"`
	CurrentContext *HomeCurrentContextState `json:"currentContext,omitempty"`
	RecentProjects []RecentProjectState     `json:"recentProjects"`
	Running        HomeRunningSummary       `json:"running"`
	Activity       HomeActivitySummary      `json:"activity"`
}

// HomeCurrentContextState is the selected context summary for one project.
type HomeCurrentContextState struct {
	ID         string                `json:"id"`
	Name       string                `json:"name"`
	Tool       ToolState             `json:"tool"`
	Confidence LaunchConfidenceState `json:"confidence"`
}

// RecentProjectsState contains recent successful launches for presentation.
// Entries are ordered from most to least recently launched.
type RecentProjectsState struct {
	Projects []RecentProjectState `json:"projects"`
}

// RecentProjectState is the presentation-safe record of one successful
// project launch. Context metadata is retained even when a project has no
// remembered project binding.
type RecentProjectState struct {
	Project        ProjectState `json:"project"`
	ContextID      string       `json:"contextId"`
	ContextName    string       `json:"contextName,omitempty"`
	LastLaunchedAt time.Time    `json:"lastLaunchedAt"`
}

// ContextListState contains every configured context and the aggregate data
// needed to present an identity list.
type ContextListState struct {
	Contexts []ContextListItem `json:"contexts"`
}

// ContextListItem combines a context's backend-owned readiness state with its
// project-binding count and most recent successful launch. EnabledProviders is
// intentionally pre-filtered so clients do not need to infer the summary.
type ContextListItem struct {
	Context          ContextState    `json:"context"`
	EnabledProviders []ProviderState `json:"enabledProviders"`
	ProjectCount     int             `json:"projectCount"`
	LastUsedAt       *time.Time      `json:"lastUsedAt,omitempty"`
}

// GetContextDetailsRequest identifies one configured context.
type GetContextDetailsRequest struct {
	ContextID string `json:"contextId"`
}

// ContextDetailsState contains the backend-owned data for one context's
// detail view. It extends the list summary with its storage location and
// creation time.
type ContextDetailsState struct {
	Context          ContextState    `json:"context"`
	Location         string          `json:"location"`
	CreatedAt        time.Time       `json:"createdAt"`
	ProjectCount     int             `json:"projectCount"`
	LastUsedAt       *time.Time      `json:"lastUsedAt,omitempty"`
	EnabledProviders []ProviderState `json:"enabledProviders"`
}

// HomeRunningSummary reserves aggregate running-environment data for later
// running-environment tracking phases.
type HomeRunningSummary struct {
	Count              int                       `json:"count"`
	ContextCounts      []HomeRunningContextCount `json:"contextCounts"`
	IsolationProtected bool                      `json:"isolationProtected"`
}

type HomeRunningContextCount struct {
	ContextID   string `json:"contextId"`
	ContextName string `json:"contextName"`
	Count       int    `json:"count"`
}

// HomeActivitySummary reserves aggregate activity data for later history
// phases.
type HomeActivitySummary struct {
	Count int `json:"count"`
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
	Description    string                `json:"description,omitempty"`
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
	SetupAction *ProviderSetupAction   `json:"setupAction,omitempty"`
	Identity    ProviderIdentityState  `json:"identity"`
}

// ProviderSetupAction describes the next backend-derived setup state for an
// enabled provider. It contains presentation-safe text only; provider-specific
// setup mechanics remain owned by the provider integration.
type ProviderSetupAction struct {
	State   ProviderSetupState `json:"state"`
	Label   string             `json:"label"`
	Message string             `json:"message"`
}

// ProviderSetupState is the UI-facing provider setup vocabulary.
type ProviderSetupState string

const (
	ProviderSetupOpenAndConfigure ProviderSetupState = "open_and_configure"
	ProviderSetupWaitingForSignIn ProviderSetupState = "waiting_for_sign_in"
	ProviderSetupVerified         ProviderSetupState = "verified"
)

// Valid reports whether state is one of the bounded API provider setup states.
func (s ProviderSetupState) Valid() bool {
	switch s {
	case ProviderSetupOpenAndConfigure, ProviderSetupWaitingForSignIn, ProviderSetupVerified:
		return true
	default:
		return false
	}
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
	Project                    ProjectState                `json:"project"`
	Context                    ContextState                `json:"context"`
	Confidence                 LaunchConfidenceState       `json:"confidence"`
	VerificationSteps          []LaunchVerificationStep    `json:"verificationSteps,omitempty"`
	Warnings                   []ResolutionWarning         `json:"warnings,omitempty"`
	RunningEnvironmentConflict *RunningEnvironmentConflict `json:"runningEnvironmentConflict,omitempty"`
}

// RunningEnvironmentConflict identifies an active environment for the same project.
type RunningEnvironmentConflict struct {
	Kind        string                  `json:"kind"`
	Environment RunningEnvironmentState `json:"environment"`
}

// LaunchVerificationStepStatus identifies the current state of one
// presentation-safe launch verification step.
type LaunchVerificationStepStatus string

const (
	LaunchVerificationStepPending        LaunchVerificationStepStatus = "pending"
	LaunchVerificationStepReady          LaunchVerificationStepStatus = "ready"
	LaunchVerificationStepNeedsAttention LaunchVerificationStepStatus = "needs_attention"
	LaunchVerificationStepBlocked        LaunchVerificationStepStatus = "blocked"
)

// LaunchVerificationStep describes one stage of a safe launch. It does not
// expose runtime paths, commands, or environment variables.
type LaunchVerificationStep struct {
	ID      string                       `json:"id"`
	Label   string                       `json:"label"`
	Status  LaunchVerificationStepStatus `json:"status"`
	Message string                       `json:"message"`
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
	ContextID          string   `json:"contextId"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	Icon               string   `json:"icon,omitempty"`
	Accent             string   `json:"accent,omitempty"`
	EnabledProviderIDs []string `json:"enabledProviderIds,omitempty"`
	ToolID             string   `json:"toolId,omitempty"`
	ImportProviderIDs  []string `json:"importProviderIds,omitempty"`
}

// CreateContextResult describes a newly created context.
type CreateContextResult struct {
	Context ContextState `json:"context"`
}

// ProjectsState contains all known projects from remembered bindings and
// successful launch history.
type ProjectsState struct {
	Projects []ProjectListItem `json:"projects"`
}
type ProjectListItem struct {
	Project        ProjectState `json:"project"`
	ContextID      string       `json:"contextId,omitempty"`
	ContextName    string       `json:"contextName,omitempty"`
	LastLaunchedAt *time.Time   `json:"lastLaunchedAt,omitempty"`
	Running        bool         `json:"running"`
}

// GetDiagnosticsRequest identifies the context that diagnostics should inspect.
// An empty ContextID reserves application-wide diagnostics for a later phase.
type GetDiagnosticsRequest struct {
	ContextID string `json:"contextId,omitempty"`
}

// DiagnosticsState contains structured, presentation-safe diagnostics. Checks
// are added by their owning integration phases; this contract deliberately does
// not expose raw environment variables, credentials, or filesystem paths.
type DiagnosticsState struct {
	Groups []DiagnosticGroup `json:"groups"`
}

// DiagnosticGroup collects related checks for one diagnostics surface.
type DiagnosticGroup struct {
	ID     string            `json:"id"`
	Label  string            `json:"label"`
	Checks []DiagnosticCheck `json:"checks"`
}

// DiagnosticCheck describes one backend-derived diagnostic result.
type DiagnosticCheck struct {
	ID       string             `json:"id"`
	Severity DiagnosticSeverity `json:"severity"`
	Label    string             `json:"label"`
	Message  string             `json:"message"`
	Details  []DiagnosticDetail `json:"details,omitempty"`
}

// DiagnosticSeverity uses the shared UI status vocabulary for diagnostic
// presentation without requiring frontend severity inference.
type DiagnosticSeverity string

const (
	DiagnosticSeverityReady          DiagnosticSeverity = "ready"
	DiagnosticSeverityNeedsAttention DiagnosticSeverity = "needs_attention"
	DiagnosticSeverityBlocked        DiagnosticSeverity = "blocked"
)

// DiagnosticDetail is a safe backend-provided detail. When IsPath is true,
// the frontend must keep Value behind its path disclosure control.
type DiagnosticDetail struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	IsPath bool   `json:"isPath,omitempty"`
}

// GetRepairActionsRequest identifies the context whose repair options should
// be evaluated.
type GetRepairActionsRequest struct {
	ContextID string `json:"contextId"`
}

// RepairActionsState contains the repair actions available for one context.
type RepairActionsState struct {
	Actions []RepairAction `json:"actions"`
}

// RepairAction is a backend-owned repair operation. Targets preview the exact
// integration-owned paths affected by the operation.
type RepairAction struct {
	ID                   string         `json:"id"`
	Label                string         `json:"label"`
	Description          string         `json:"description"`
	Destructive          bool           `json:"destructive"`
	RequiresConfirmation bool           `json:"requiresConfirmation"`
	Targets              []RepairTarget `json:"targets"`
}

// RepairTarget describes one path affected by a repair action. Paths are
// presentation-safe only when revealed through the frontend's path disclosure.
type RepairTarget struct {
	Label string `json:"label"`
	Path  string `json:"path"`
	Kind  string `json:"kind"`
}

// RunRepairActionRequest asks the service to execute one advertised action.
// ConfirmDestructive must be true for any destructive action.
type RunRepairActionRequest struct {
	ContextID          string `json:"contextId"`
	ActionID           string `json:"actionId"`
	ConfirmDestructive bool   `json:"confirmDestructive"`
}

// RunRepairActionResult returns refreshed diagnostics after one action.
type RunRepairActionResult struct {
	ActionID    string           `json:"actionId"`
	Diagnostics DiagnosticsState `json:"diagnostics"`
}

// HistoryState contains local launch events for later history presentation.
type HistoryState struct {
	Entries []HistoryEntry `json:"entries"`
}

// HistoryEntry is a presentation-safe local activity record.
type HistoryEntry struct {
	Event       string          `json:"event"`
	Category    HistoryCategory `json:"category"`
	Timestamp   time.Time       `json:"timestamp"`
	ProjectPath string          `json:"projectPath,omitempty"`
	ContextID   string          `json:"contextId,omitempty"`
	ToolID      string          `json:"toolId,omitempty"`
	Message     string          `json:"message"`
}

// HistoryCategory groups activity entries for user-facing history filters.
type HistoryCategory string

const (
	HistoryCategoryLaunch        HistoryCategory = "launch"
	HistoryCategoryConfiguration HistoryCategory = "configuration"
	HistoryCategoryWarning       HistoryCategory = "warning"
)

// RunningEnvironmentState is the presentation-safe model for one immutable
// coding-tool environment. Population and status refresh follow in later
// running-environment phases.
type RunningEnvironmentState struct {
	ID        string                         `json:"id"`
	Project   ProjectState                   `json:"project"`
	Context   RunningEnvironmentContextState `json:"context"`
	Tool      ToolOption                     `json:"tool"`
	StartedAt time.Time                      `json:"startedAt"`
	Process   RunningEnvironmentProcessState `json:"process"`
	Session   RunningEnvironmentSessionState `json:"session"`
	Launch    RunningEnvironmentLaunchState  `json:"launch"`
}

// RunningEnvironmentContextState identifies the immutable context selected at launch.
type RunningEnvironmentContextState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// RunningEnvironmentProcessState reports process state without requiring a PID.
type RunningEnvironmentProcessState struct {
	State string `json:"state"`
	PID   *int   `json:"pid,omitempty"`
}

// RunningEnvironmentSessionState reports coding-tool session state when available.
type RunningEnvironmentSessionState struct {
	ID    string `json:"id,omitempty"`
	State string `json:"state"`
}

// RunningEnvironmentLaunchState preserves the safe identity of the launch.
type RunningEnvironmentLaunchState struct {
	Source           string `json:"source"`
	ResolutionSource string `json:"resolutionSource"`
}

// RunningEnvironmentsState contains active coding-tool environments.
type RunningEnvironmentsState struct {
	Environments []RunningEnvironmentState `json:"environments"`
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
