package provider

// ID identifies a provider configuration.
type ID string

// Options stores provider-specific, non-secret configuration values.
type Options map[string]string

// Config describes whether a provider should participate in a context.
type Config struct {
	Enabled bool
	Options Options
}

// Configs stores provider intent keyed by provider identifier.
type Configs map[ID]Config
