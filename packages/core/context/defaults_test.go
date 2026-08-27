package context_test

import (
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

func TestDefaultContextSeeds(t *testing.T) {
	createdAt := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name        string
		context     devcontext.Context
		wantID      devcontext.ID
		wantName    string
		wantCreated time.Time
	}{
		{
			name:        "personal",
			context:     devcontext.DefaultPersonalContext(createdAt),
			wantID:      devcontext.MustID("personal"),
			wantName:    "Personal",
			wantCreated: createdAt,
		},
		{
			name:        "company",
			context:     devcontext.DefaultCompanyContext(createdAt),
			wantID:      devcontext.MustID("company"),
			wantName:    "Company",
			wantCreated: createdAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.context.ID != tt.wantID {
				t.Fatalf("ID = %q, want %q", tt.context.ID, tt.wantID)
			}
			if tt.context.Name != tt.wantName {
				t.Fatalf("name = %q, want %q", tt.context.Name, tt.wantName)
			}
			if !reflect.DeepEqual(tt.context.Tool, codingtool.DefaultLaunchTarget()) {
				t.Fatalf("launch target = %#v, want %#v", tt.context.Tool, codingtool.DefaultLaunchTarget())
			}
			assertEnabledProvider(t, tt.context.Providers, "claude")
			assertEnabledProvider(t, tt.context.Providers, "codex")
			if len(tt.context.Providers) != 2 {
				t.Fatalf("provider count = %d, want 2", len(tt.context.Providers))
			}
			if !tt.context.CreatedAt.Equal(tt.wantCreated) {
				t.Fatalf("created at = %s, want %s", tt.context.CreatedAt, tt.wantCreated)
			}
		})
	}
}

func TestDefaultContextSeedsUseRegistryDefaults(t *testing.T) {
	registry := provider.MustNewRegistry([]provider.Provider{defaultProvider{id: "future", name: "Future Provider"}}, "future")
	ctx := devcontext.DefaultPersonalContextWithProviderRegistry(time.Now(), registry)
	if len(ctx.Providers) != 1 {
		t.Fatalf("provider count = %d, want one", len(ctx.Providers))
	}
	assertEnabledProvider(t, ctx.Providers, "future")
}

type defaultProvider struct {
	id   provider.ID
	name string
}

func (p defaultProvider) ID() provider.ID { return p.id }

func (p defaultProvider) DisplayName() string { return p.name }

func (defaultProvider) BuildEnvironment(provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return provider.EnvironmentContribution{}, nil
}

func (defaultProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}

func assertEnabledProvider(t *testing.T, providers provider.Configs, providerID provider.ID) {
	t.Helper()

	config, ok := providers[providerID]
	if !ok {
		t.Fatalf("provider %q is missing", providerID)
	}
	if !config.Enabled {
		t.Fatalf("provider %q enabled = false, want true", providerID)
	}
}
