package provider_test

import (
	"reflect"
	"testing"

	"devctx/packages/core/provider"
)

func TestRegistryPreservesProviderOrderAndLookup(t *testing.T) {
	first := registryFakeProvider{id: "first", displayName: "First Provider"}
	second := registryFakeProvider{id: "second", displayName: "Second Provider"}

	registry, err := provider.NewRegistry([]provider.Provider{first, second}, "second")
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	if !reflect.DeepEqual(registry.All(), []provider.Provider{first, second}) {
		t.Fatalf("providers = %#v, want input order", registry.All())
	}
	if got, ok := registry.Get("first"); !ok || got.DisplayName() != "First Provider" {
		t.Fatalf("lookup first = %#v, %t", got, ok)
	}
	if _, ok := registry.Get("missing"); ok {
		t.Fatal("missing provider was found")
	}
	if got := registry.DisplayName("second"); got != "Second Provider" {
		t.Fatalf("display name = %q, want Second Provider", got)
	}
	if got := registry.DisplayName("missing"); got != "missing" {
		t.Fatalf("missing display name = %q, want raw ID", got)
	}
}

func TestRegistryBuildsDefaultConfigs(t *testing.T) {
	registry, err := provider.NewRegistry(
		[]provider.Provider{
			registryFakeProvider{id: "first"},
			registryFakeProvider{id: "second"},
			registryFakeProvider{id: "third"},
		},
		"third",
		"first",
	)
	if err != nil {
		t.Fatalf("create registry: %v", err)
	}

	if got := registry.DefaultEnabledIDs(); !reflect.DeepEqual(got, []provider.ID{"first", "third"}) {
		t.Fatalf("default IDs = %#v, want registry order", got)
	}
	if got := registry.DefaultConfigs(); !reflect.DeepEqual(got, provider.Configs{
		"first": {Enabled: true},
		"third": {Enabled: true},
	}) {
		t.Fatalf("default configs = %#v", got)
	}
}

func TestRegistryRejectsInvalidProviders(t *testing.T) {
	tests := []struct {
		name      string
		providers []provider.Provider
		defaults  []provider.ID
	}{
		{
			name:      "nil provider",
			providers: []provider.Provider{nil},
		},
		{
			name:      "empty id",
			providers: []provider.Provider{registryFakeProvider{}},
		},
		{
			name: "duplicate id",
			providers: []provider.Provider{
				registryFakeProvider{id: "duplicate"},
				registryFakeProvider{id: "duplicate"},
			},
		},
		{
			name:      "unknown default",
			providers: []provider.Provider{registryFakeProvider{id: "known"}},
			defaults:  []provider.ID{"missing"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := provider.NewRegistry(tt.providers, tt.defaults...); err == nil {
				t.Fatal("error = nil, want validation error")
			}
		})
	}
}

func TestDefaultRegistryContainsDefaultProviders(t *testing.T) {
	registry := provider.DefaultRegistry()

	if _, ok := registry.Get(provider.ClaudeID); !ok {
		t.Fatal("Claude provider is not registered")
	}
	if _, ok := registry.Get(provider.CodexID); !ok {
		t.Fatal("Codex provider is not registered")
	}
	if got := registry.DefaultConfigs(); !reflect.DeepEqual(got, provider.Configs{
		provider.ClaudeID: {Enabled: true},
		provider.CodexID:  {Enabled: true},
	}) {
		t.Fatalf("builtin default configs = %#v", got)
	}
}

type registryFakeProvider struct {
	id          provider.ID
	displayName string
}

func (p registryFakeProvider) ID() provider.ID {
	return p.id
}

func (p registryFakeProvider) DisplayName() string {
	if p.displayName != "" {
		return p.displayName
	}
	return string(p.id)
}

func (p registryFakeProvider) BuildEnvironment(provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return nil, nil
}

func (p registryFakeProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}
