package provider

import "fmt"

// Registry owns the set of providers Dev Context knows how to run.
//
// The registry preserves provider order for stable UI and CLI presentation,
// provides lookup by persisted provider ID, and identifies which providers are
// enabled in newly seeded contexts.
type Registry struct {
	providers         []Provider
	providersByID     map[ID]Provider
	defaultEnabledIDs map[ID]struct{}
}

// NewRegistry creates a provider registry from an ordered provider list.
func NewRegistry(providers []Provider, defaultEnabledIDs ...ID) (Registry, error) {
	ordered := make([]Provider, 0, len(providers))
	byID := make(map[ID]Provider, len(providers))

	for i, integration := range providers {
		if integration == nil {
			return Registry{}, fmt.Errorf("provider registry contains nil provider at index %d", i)
		}
		id := integration.ID()
		if id == "" {
			return Registry{}, fmt.Errorf("provider registry contains provider with empty ID at index %d", i)
		}
		if _, exists := byID[id]; exists {
			return Registry{}, fmt.Errorf("provider registry contains duplicate provider ID %q", id)
		}
		ordered = append(ordered, integration)
		byID[id] = integration
	}

	defaults := make(map[ID]struct{}, len(defaultEnabledIDs))
	for _, id := range defaultEnabledIDs {
		if id == "" {
			return Registry{}, fmt.Errorf("provider registry contains empty default provider ID")
		}
		if _, ok := byID[id]; !ok {
			return Registry{}, fmt.Errorf("provider registry default provider %q is not registered", id)
		}
		defaults[id] = struct{}{}
	}

	return Registry{
		providers:         ordered,
		providersByID:     byID,
		defaultEnabledIDs: defaults,
	}, nil
}

// MustNewRegistry creates a provider registry and panics if the registry is
// invalid. It is intended for package-level or default application wiring.
func MustNewRegistry(providers []Provider, defaultEnabledIDs ...ID) Registry {
	registry, err := NewRegistry(providers, defaultEnabledIDs...)
	if err != nil {
		panic(err)
	}
	return registry
}

// BuiltInRegistry returns the built-in providers enabled by default in Dev
// Context.
func BuiltInRegistry() Registry {
	return MustNewRegistry(
		[]Provider{
			ClaudeProvider{},
			CodexProvider{},
		},
		ClaudeID,
		CodexID,
	)
}

// IsZero reports whether the registry has not been initialized.
func (r Registry) IsZero() bool {
	return r.providers == nil && r.providersByID == nil && r.defaultEnabledIDs == nil
}

// All returns registered providers in stable presentation order.
func (r Registry) All() []Provider {
	return append([]Provider(nil), r.providers...)
}

// Get returns the provider registered for id.
func (r Registry) Get(id ID) (Provider, bool) {
	integration, ok := r.providersByID[id]
	return integration, ok
}

// DefaultEnabledIDs returns default-enabled providers in registry order.
func (r Registry) DefaultEnabledIDs() []ID {
	ids := make([]ID, 0, len(r.defaultEnabledIDs))
	for _, integration := range r.providers {
		id := integration.ID()
		if _, ok := r.defaultEnabledIDs[id]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// DefaultConfigs returns provider configs for a newly seeded context.
func (r Registry) DefaultConfigs() Configs {
	ids := r.DefaultEnabledIDs()
	configs := make(Configs, len(ids))
	for _, id := range ids {
		configs[id] = Config{Enabled: true}
	}
	return configs
}

// DisplayName returns the registered provider display name, or the raw ID when
// the provider is not registered.
func (r Registry) DisplayName(id ID) string {
	integration, ok := r.Get(id)
	if !ok {
		return string(id)
	}
	return integration.DisplayName()
}
