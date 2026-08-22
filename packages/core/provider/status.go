package provider

// StatusState identifies one local provider readiness state.
type StatusState string

const (
	// StatusConfigured means the provider appears locally configured for the
	// context.
	StatusConfigured StatusState = "configured"

	// StatusReady is kept as a source compatibility alias for older call sites.
	StatusReady StatusState = StatusConfigured

	// StatusNotConfigured means provider storage exists but does not appear
	// initialized.
	StatusNotConfigured StatusState = "not_configured"

	// StatusDirectoryMissing means the context-owned provider directory is
	// absent.
	StatusDirectoryMissing StatusState = "directory_missing"

	// StatusUnavailable means readiness could not be determined locally.
	StatusUnavailable StatusState = "unavailable"
)

// Status is the provider's non-sensitive local readiness summary.
type Status struct {
	State       StatusState `json:"state"`
	Explanation string      `json:"explanation,omitempty"`
}

// ReadyStatus reports a locally ready provider.
func ReadyStatus() Status {
	return ConfiguredStatus()
}

// ConfiguredStatus reports a locally configured provider.
func ConfiguredStatus() Status {
	return Status{State: StatusConfigured}
}

// NotConfiguredStatus reports provider storage that is present but
// uninitialized.
func NotConfiguredStatus(explanation string) Status {
	return Status{State: StatusNotConfigured, Explanation: explanation}
}

// DirectoryMissingStatus reports missing context-owned provider storage.
func DirectoryMissingStatus(explanation string) Status {
	return Status{State: StatusDirectoryMissing, Explanation: explanation}
}

// UnavailableStatus reports that local readiness could not be determined.
func UnavailableStatus(explanation string) Status {
	return Status{State: StatusUnavailable, Explanation: explanation}
}

// Valid reports whether state is one of the bounded provider readiness states.
func (s StatusState) Valid() bool {
	switch s {
	case StatusConfigured, StatusNotConfigured, StatusDirectoryMissing, StatusUnavailable:
		return true
	default:
		return false
	}
}
