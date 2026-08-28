// Package running defines the durable identity and observed state of a coding-tool environment.
package running

import (
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

// ID identifies one running environment record.
type ID string

// Environment captures the immutable launch identity and observable runtime
// state of one coding-tool environment.
type Environment struct {
	ID        ID
	Project   ProjectIdentity
	Context   ContextIdentity
	Tool      ToolIdentity
	StartedAt time.Time
	Process   Process
	Session   Session
	Launch    LaunchIdentity
}

// ProjectIdentity identifies the project opened by an environment.
type ProjectIdentity struct {
	Path project.Path
	Name string
}

// ContextIdentity identifies the development context selected at launch.
type ContextIdentity struct {
	ID   devcontext.ID
	Name string
}

// ToolIdentity identifies the coding tool selected at launch.
type ToolIdentity struct {
	ID   codingtool.ID
	Name string
}

// ProcessState describes the operating-system process when it can be observed.
type ProcessState string

const (
	ProcessStateUnknown ProcessState = "unknown"
	ProcessStateRunning ProcessState = "running"
	ProcessStateStopped ProcessState = "stopped"
)

// Process contains optional process identity and its observed state.
type Process struct {
	PID   *int
	State ProcessState
}

// SessionState describes the coding-tool session when the integration exposes one.
type SessionState string

const (
	SessionStateUnknown SessionState = "unknown"
	SessionStateActive  SessionState = "active"
	SessionStateEnded   SessionState = "ended"
)

// Session contains optional coding-tool session identity and state.
type Session struct {
	ID    string
	State SessionState
}

// LaunchIdentity retains safe launch provenance without persisting command-line
// arguments or environment values.
type LaunchIdentity struct {
	Source           launcher.InvocationSource
	ResolutionSource launcher.ResolutionSource
}
