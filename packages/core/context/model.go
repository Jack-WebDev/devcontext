package context

import (
	"time"

	"devctx/packages/core/editor"
	"devctx/packages/core/provider"
)

// Metadata stores non-sensitive context annotations.
type Metadata map[string]string

// Context represents a named development identity.
type Context struct {
	ID ID

	// Name is the user-facing display name.
	Name string

	Editor    editor.Config
	Providers provider.Configs
	Metadata  Metadata
	CreatedAt time.Time
}
