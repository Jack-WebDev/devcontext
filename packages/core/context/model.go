package context

import (
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/provider"
)

// Metadata stores non-sensitive context annotations.
type Metadata map[string]string

// Context represents a named development identity.
type Context struct {
	ID ID

	// Name is the user-facing display name.
	Name string

	Tool       codingtool.LaunchTarget
	Providers  provider.Configs
	Metadata   Metadata
	CreatedAt  time.Time
	ArchivedAt *time.Time
}

// IsArchived reports whether this context is retired from routine use.
func (c Context) IsArchived() bool { return c.ArchivedAt != nil }
