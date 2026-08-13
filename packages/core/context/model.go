package context

import "time"

// ID identifies a development context.
//
// Validation of safe filesystem identifiers is handled by the context
// validation boundary; this type exists so domain models do not pass raw
// strings for context identity.
type ID string

// EditorType identifies the editor a context intends to launch.
type EditorType string

// EditorConfig describes editor intent for a context.
type EditorConfig struct {
	Type EditorType
}

// ProviderID identifies a provider configuration within a context.
type ProviderID string

// ProviderConfig describes whether a provider should participate in a context.
type ProviderConfig struct {
	Enabled bool
}

// ProviderConfigs stores provider intent keyed by provider identifier.
type ProviderConfigs map[ProviderID]ProviderConfig

// Metadata stores non-sensitive context annotations.
type Metadata map[string]string

// Context represents a named development identity.
type Context struct {
	ID ID

	// Name is the user-facing display name.
	Name string

	Editor    EditorConfig
	Providers ProviderConfigs
	Metadata  Metadata
	CreatedAt time.Time
}
