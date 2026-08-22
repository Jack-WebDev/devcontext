package launcher

// ConfidenceStatus identifies the backend-owned launch readiness state used by
// UI-facing launch confidence contracts.
type ConfidenceStatus string

const (
	// ConfidenceReady means everything required for safe launch is available.
	ConfidenceReady ConfidenceStatus = "ready"

	// ConfidenceNeedsAttention means launch may be possible, but the user should
	// review non-blocking issues first.
	ConfidenceNeedsAttention ConfidenceStatus = "needs_attention"

	// ConfidenceBlocked means Dev Context cannot guarantee the requested safe
	// launch, so the UI must not offer the launch action.
	ConfidenceBlocked ConfidenceStatus = "blocked"
)

// Valid reports whether status is one of the bounded launch confidence states.
func (s ConfidenceStatus) Valid() bool {
	switch s {
	case ConfidenceReady, ConfidenceNeedsAttention, ConfidenceBlocked:
		return true
	default:
		return false
	}
}

// ConfidenceCheckComponent identifies the system area evaluated by one launch
// confidence check.
type ConfidenceCheckComponent string

const (
	// ConfidenceCheckClaude identifies Claude provider readiness.
	ConfidenceCheckClaude ConfidenceCheckComponent = "claude"

	// ConfidenceCheckCodex identifies Codex provider readiness.
	ConfidenceCheckCodex ConfidenceCheckComponent = "codex"

	// ConfidenceCheckVSCode identifies VS Code launch readiness.
	ConfidenceCheckVSCode ConfidenceCheckComponent = "vscode"

	// ConfidenceCheckIsolation identifies context and environment isolation
	// readiness.
	ConfidenceCheckIsolation ConfidenceCheckComponent = "isolation"
)

// Valid reports whether component is one of the bounded confidence check
// components.
func (c ConfidenceCheckComponent) Valid() bool {
	switch c {
	case ConfidenceCheckClaude, ConfidenceCheckCodex, ConfidenceCheckVSCode, ConfidenceCheckIsolation:
		return true
	default:
		return false
	}
}

// ConfidenceCheck is a presentation-safe backend-owned launch readiness check.
// It gives the UI stable semantics without exposing raw provider strings,
// filesystem details, or environment internals.
type ConfidenceCheck struct {
	Component  ConfidenceCheckComponent `json:"component"`
	Severity   ConfidenceStatus         `json:"severity"`
	Label      string                   `json:"label"`
	Message    string                   `json:"message"`
	ActionHint string                   `json:"actionHint,omitempty"`
}

// Valid reports whether check has a known component, known severity, and the
// user-facing text required for display.
func (c ConfidenceCheck) Valid() bool {
	return c.Component.Valid() && c.Severity.Valid() && c.Label != "" && c.Message != ""
}
