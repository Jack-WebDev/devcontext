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

// LaunchFailureDetails contains presentation-safe diagnostics for a failed
// process start. It intentionally excludes command arguments, environment
// values, and context storage paths.
type LaunchFailureDetails struct {
	Executable string    `json:"executable"`
	ExitCode   *int      `json:"exitCode,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
	Logs       string    `json:"logs"`
}

// TrayState describes presentation-safe system-tray content. The desktop host
// owns rendering and platform availability; this contract has no OS details.
type TrayState struct {
	Enabled        bool                    `json:"enabled"`
	Indicator      string                  `json:"indicator"`
	Environments   []TrayEnvironmentItem   `json:"environments"`
	RecentProjects []TrayRecentProjectItem `json:"recentProjects"`
}
type TrayEnvironmentItem struct {
	ID          string `json:"id"`
	ProjectName string `json:"projectName"`
	ContextName string `json:"contextName"`
	ToolName    string `json:"toolName"`
}
type TrayRecentProjectItem struct {
	ProjectName string `json:"projectName"`
	ContextName string `json:"contextName"`
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

// UpdateContextDetailsRequest changes only the human-readable identity fields.
// The internal ID, tool configuration, providers, and project bindings remain
// owned by their respective contracts.
type UpdateContextDetailsRequest struct {
	ContextID   string `json:"contextId"`
	Name        string `json:"name"`
	Purpose     string `json:"purpose,omitempty"`
	Description string `json:"description,omitempty"`
}

// UpdateContextAppearanceRequest changes only presentation metadata.
type UpdateContextAppearanceRequest struct {
	ContextID string `json:"contextId"`
	Icon      string `json:"icon,omitempty"`
	Accent    string `json:"accent,omitempty"`
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

// TrustCenterState contains factual local protection and integration-boundary
// information for the Trust Center. It intentionally contains no credentials,
// commands, environment values, or context storage paths.
type TrustCenterState struct {
	Contexts              []TrustContextProtection      `json:"contexts"`
	ProjectMappings       []TrustProjectMapping         `json:"projectMappings"`
	CredentialSync        TrustCredentialSyncProtection `json:"credentialSync"`
	IntegrationBoundaries []TrustIntegrationBoundary    `json:"integrationBoundaries"`
}

// TrustContextProtection reports the actual isolation readiness for one
// configured development identity.
type TrustContextProtection struct {
	ID        string                    `json:"id"`
	Name      string                    `json:"name"`
	Providers []TrustProviderProtection `json:"providers"`
	Tool      TrustCodingToolProtection `json:"tool"`
}

// TrustProviderProtection reports one enabled provider's isolated storage
// readiness without disclosing its storage location.
type TrustProviderProtection struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Isolation TrustIsolationProtection `json:"isolation"`
}

// TrustCodingToolProtection reports the selected coding tool's isolated
// storage readiness without exposing commands or host paths.
type TrustCodingToolProtection struct {
	ID        string                   `json:"id"`
	Name      string                   `json:"name"`
	Isolation TrustIsolationProtection `json:"isolation"`
}

// TrustIsolationProtection contains backend-derived isolation readiness.
type TrustIsolationProtection struct {
	Status  LaunchConfidenceStatus `json:"status"`
	Message string                 `json:"message"`
}

// TrustProjectMapping describes an explicit remembered project-to-context
// relationship.
type TrustProjectMapping struct {
	Project     ProjectState `json:"project"`
	ContextID   string       `json:"contextId"`
	ContextName string       `json:"contextName"`
}

// TrustCredentialSyncProtection reports Dev Context's actual credential-sync
// boundary. Credentials remain in local provider-owned context storage.
type TrustCredentialSyncProtection struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
}

// TrustIntegrationBoundary describes the data Dev Context can provide to one
// selected coding-tool integration. It never grants credentials or commands.
type TrustIntegrationBoundary struct {
	ToolID              string `json:"toolId"`
	ToolName            string `json:"toolName"`
	StatusDataAvailable bool   `json:"statusDataAvailable"`
	Message             string `json:"message"`
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

// ValidateProjectDirectoryRequest asks the application to canonicalize and
// validate a project folder before it is used by a later operation.
type ValidateProjectDirectoryRequest struct {
	ProjectPath string `json:"projectPath"`
}

// ContextState is the presentation-safe identity and readiness summary for one
// configured context.
type ContextState struct {
	ID               string                       `json:"id"`
	Name             string                       `json:"name"`
	Purpose          string                       `json:"purpose,omitempty"`
	Description      string                       `json:"description,omitempty"`
	Tool             ToolState                    `json:"tool"`
	AvailableTools   []ToolOption                 `json:"availableTools"`
	Providers        []ProviderState              `json:"providers"`
	DevelopmentTools []DevelopmentToolIntegration `json:"developmentTools"`
	Confidence       LaunchConfidenceState        `json:"confidence"`
	Metadata         map[string]string            `json:"metadata,omitempty"`
}

// DevelopmentToolIntegration is a generic, presentation-safe development
// adapter. It lets UI surfaces render registered integrations by category
// without knowing their provider or coding-tool implementation.
type DevelopmentToolIntegration struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	Category     DevelopmentToolCategory `json:"category"`
	Status       DevelopmentToolStatus   `json:"status"`
	Message      string                  `json:"message"`
	RecoveryHint string                  `json:"recoveryHint,omitempty"`
	Enabled      bool                    `json:"enabled"`
}

type DevelopmentToolCategory string

const (
	DevelopmentToolCategoryCoding          DevelopmentToolCategory = "coding"
	DevelopmentToolCategoryAI              DevelopmentToolCategory = "ai"
	DevelopmentToolCategoryVersionControl  DevelopmentToolCategory = "version-control"
	DevelopmentToolCategorySourceHosting   DevelopmentToolCategory = "source-hosting"
	DevelopmentToolCategoryCloudRegistries DevelopmentToolCategory = "cloud-registries"
	DevelopmentToolCategoryOther           DevelopmentToolCategory = "other"
)

// DevelopmentToolStatus is the bounded, user-facing readiness vocabulary for
// any registered development integration.
type DevelopmentToolStatus string

const (
	DevelopmentToolAvailable     DevelopmentToolStatus = "available"
	DevelopmentToolConnected     DevelopmentToolStatus = "connected"
	DevelopmentToolNeedsSignIn   DevelopmentToolStatus = "needs_sign_in"
	DevelopmentToolNotConfigured DevelopmentToolStatus = "not_configured"
	DevelopmentToolNotFound      DevelopmentToolStatus = "not_found"
	DevelopmentToolUnavailable   DevelopmentToolStatus = "unavailable"
	DevelopmentToolError         DevelopmentToolStatus = "error"
)

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
	Groups                     []PreflightGroup            `json:"groups"`
	VerificationSteps          []LaunchVerificationStep    `json:"verificationSteps,omitempty"`
	Warnings                   []ResolutionWarning         `json:"warnings,omitempty"`
	RunningEnvironmentConflict *RunningEnvironmentConflict `json:"runningEnvironmentConflict,omitempty"`
}

// PreflightGroupID identifies one product-level area checked before launch.
type PreflightGroupID string

const (
	PreflightGroupProject   PreflightGroupID = "project"
	PreflightGroupContext   PreflightGroupID = "context"
	PreflightGroupIsolation PreflightGroupID = "isolation"
	PreflightGroupTools     PreflightGroupID = "tools"
	PreflightGroupWorkspace PreflightGroupID = "workspace"
)

// PreflightGroup is an aggregate preflight result for one product concept.
// Checks contain presentation-safe evidence for an optional detail disclosure.
type PreflightGroup struct {
	ID       PreflightGroupID       `json:"id"`
	Label    string                 `json:"label"`
	Status   LaunchConfidenceStatus `json:"status"`
	Blocking bool                   `json:"blocking"`
	Message  string                 `json:"message"`
	Checks   []PreflightCheck       `json:"checks"`
}

// PreflightCheck is one presentation-safe piece of evidence within a group.
// Blocking is the application-owned launch policy for this check: blocking
// checks cannot continue without remediation, while needs-attention checks
// require a deliberate choice in the UI.
type PreflightCheck struct {
	ID         string                 `json:"id"`
	Label      string                 `json:"label"`
	Status     LaunchConfidenceStatus `json:"status"`
	Blocking   bool                   `json:"blocking"`
	Message    string                 `json:"message"`
	ActionHint string                 `json:"actionHint,omitempty"`
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

// LaunchVerificationStep describes a presentation-safe launch stage. A
// preflight result contains only pending stages because the work begins later
// in LaunchProject; it never represents work already completed by preflight.
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
	// ContextID is retained for imports and API callers that need to preserve an
	// existing identifier. New UI flows should omit it; the service derives a
	// stable filesystem-safe ID from Name.
	ContextID          string   `json:"contextId"`
	TemplateID         string   `json:"templateId,omitempty"`
	Name               string   `json:"name,omitempty"`
	Purpose            string   `json:"purpose,omitempty"`
	Description        string   `json:"description,omitempty"`
	Icon               string   `json:"icon,omitempty"`
	Accent             string   `json:"accent,omitempty"`
	EnabledProviderIDs []string `json:"enabledProviderIds,omitempty"`
	// EnabledDevelopmentToolIDs is the generic creation-flow representation of
	// selected registered integrations. It supersedes the UI's need to split
	// selections by adapter type; the legacy fields above remain supported for
	// existing callers.
	EnabledDevelopmentToolIDs []string `json:"enabledDevelopmentToolIds,omitempty"`
	ToolID                    string   `json:"toolId,omitempty"`
	ImportProviderIDs         []string `json:"importProviderIds,omitempty"`
}

// CreateContextResult describes a newly created context.
type CreateContextResult struct {
	Context ContextState `json:"context"`
}

// ContextTemplateState is a safe set of defaults for creating a context.
// Templates do not include credentials, provider settings, or tool settings.
type ContextTemplateState struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon,omitempty"`
	Accent      string `json:"accent"`
}

// ContextTemplatesState contains the built-in templates available to the
// create-context flow.
type ContextTemplatesState struct {
	Templates []ContextTemplateState `json:"templates"`
}

// DuplicateContextRequest copies a context's safe configuration into a new
// isolated context. Credentials are never copied.
type DuplicateContextRequest struct {
	SourceContextID string `json:"sourceContextId"`
	ContextID       string `json:"contextId"`
	Name            string `json:"name,omitempty"`
}

// DuplicateContextResult describes the newly duplicated context.
type DuplicateContextResult struct {
	Context ContextState `json:"context"`
}

// ContextTransferVersion identifies the current safe context metadata format.
// The format intentionally excludes credentials and integration-owned storage.
const ContextTransferVersion = 1

// ExportContextMetadataRequest identifies the context whose portable safe
// configuration should be exported.
type ExportContextMetadataRequest struct {
	ContextID string `json:"contextId"`
}

// ContextMetadataExport is a versioned, portable context configuration. It is
// constructed only from metadata and non-secret provider and coding-tool
// settings; it is never a context-directory archive.
type ContextMetadataExport struct {
	Version int                     `json:"version"`
	Context ContextTransferMetadata `json:"context"`
}

// ContextTransferMetadata contains the portable fields for one development
// identity. It deliberately has no context ID, timestamps, paths, account
// identities, credentials, or runtime state.
type ContextTransferMetadata struct {
	Name         string                      `json:"name"`
	Metadata     map[string]string           `json:"metadata,omitempty"`
	Providers    []ContextTransferProvider   `json:"providers"`
	LaunchTarget ContextTransferLaunchTarget `json:"launchTarget"`
}

// ContextTransferProvider contains non-secret configuration for a registered
// provider. Provider IDs are checked against the receiving registry on import.
type ContextTransferProvider struct {
	ID      string            `json:"id"`
	Enabled bool              `json:"enabled"`
	Options map[string]string `json:"options,omitempty"`
}

// ContextTransferLaunchTarget contains the portable selected-tool
// configuration. Executable overrides are host-specific and are excluded.
type ContextTransferLaunchTarget struct {
	DefaultTool string                `json:"defaultTool"`
	Tools       []ContextTransferTool `json:"tools"`
}

// ContextTransferTool contains non-secret options for one registered coding
// tool. The receiving registry validates its ID during import.
type ContextTransferTool struct {
	ID      string            `json:"id"`
	Options map[string]string `json:"options,omitempty"`
}

// ImportContextMetadataRequest creates a new context from a safe metadata
// export. ContextID is always supplied by the receiving user and is never
// taken from an export document.
type ImportContextMetadataRequest struct {
	ContextID string                `json:"contextId"`
	Export    ContextMetadataExport `json:"export"`
}

// ImportContextMetadataResult describes the fresh isolated context created
// from an imported metadata document.
type ImportContextMetadataResult struct {
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
	ProviderID string `json:"providerId"`
	Name       string `json:"name"`
	// Discovered confirms that this session was found by the registered
	// integration. The frontend must not offer assignment for unverified data.
	Discovered        bool                    `json:"discovered"`
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
