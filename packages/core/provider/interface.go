package provider

// Provider is the small contract implemented by local AI/tool integrations.
//
// Implementations should only describe their own environment variables and
// local readiness. They must not perform remote authentication checks or load
// runtime plugins.
type Provider interface {
	ID() ID
	DisplayName() string
	BuildEnvironment(RuntimeContext) (EnvironmentContribution, error)
	Status(RuntimeContext) (Status, error)
}

// RuntimeContext is the provider-owned view of a selected Dev Context.
//
// It intentionally stores plain values instead of depending on the context or
// filesystem packages, keeping provider implementations usable without import
// cycles.
type RuntimeContext struct {
	ContextID string
	Config    Config
	Paths     ContextPaths
}

// ContextPaths contains the context-owned storage locations a provider may use.
type ContextPaths struct {
	RootDir           string
	StorageDir        string
	VSCodeDir         string
	VSCodeUserDataDir string
}

// EnvironmentContribution stores environment variables owned by one provider.
type EnvironmentContribution map[string]string
