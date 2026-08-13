package context

import (
	"time"

	"devctx/packages/core/provider"
)

// EditorType identifies the editor a context intends to launch.
type EditorType string

// EditorConfig describes editor intent for a context.
type EditorConfig struct {
	Type EditorType
}

// Metadata stores non-sensitive context annotations.
type Metadata map[string]string

// Context represents a named development identity.
type Context struct {
	ID ID

	// Name is the user-facing display name.
	Name string

	Editor    EditorConfig
	Providers provider.Configs
	Metadata  Metadata
	CreatedAt time.Time
}
