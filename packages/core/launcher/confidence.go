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
