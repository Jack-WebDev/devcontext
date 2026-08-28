package application

import "devctx/packages/core/launcher"

// LaunchConfidenceStatus is the API-facing launch readiness state. It aliases
// the core launcher status so application DTOs and core derivation rules cannot
// drift.
type LaunchConfidenceStatus = launcher.ConfidenceStatus

// LaunchConfidenceCheckComponent identifies the API-facing system area
// evaluated by one launch confidence check.
type LaunchConfidenceCheckComponent = launcher.ConfidenceCheckComponent

// LaunchConfidenceCheck is the API-facing representation of one backend-owned
// launch readiness check.
type LaunchConfidenceCheck = launcher.ConfidenceCheck

const (
	// LaunchConfidenceReady means everything required for safe launch is
	// available.
	LaunchConfidenceReady LaunchConfidenceStatus = launcher.ConfidenceReady

	// LaunchConfidenceNeedsAttention means launch may be possible, but the user
	// should review non-blocking issues first.
	LaunchConfidenceNeedsAttention LaunchConfidenceStatus = launcher.ConfidenceNeedsAttention

	// LaunchConfidenceBlocked means Dev Context cannot guarantee the requested
	// safe launch, so the UI must not offer the launch action.
	LaunchConfidenceBlocked LaunchConfidenceStatus = launcher.ConfidenceBlocked

	// LaunchConfidenceCheckProvider identifies registered provider readiness.
	LaunchConfidenceCheckProvider LaunchConfidenceCheckComponent = launcher.ConfidenceCheckProvider

	// LaunchConfidenceCheckTool identifies selected coding-tool readiness.
	LaunchConfidenceCheckTool LaunchConfidenceCheckComponent = launcher.ConfidenceCheckTool

	// LaunchConfidenceCheckIsolation identifies context and environment
	// isolation readiness.
	LaunchConfidenceCheckIsolation LaunchConfidenceCheckComponent = launcher.ConfidenceCheckIsolation

	// LaunchConfidenceCheckIdentity identifies meaningful conflicting verified
	// account identity evidence across enabled providers.
	LaunchConfidenceCheckIdentity LaunchConfidenceCheckComponent = launcher.ConfidenceCheckIdentity
)
