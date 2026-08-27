package context_test

import (
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

func TestContextModelConstructsNamedDevelopmentIdentities(t *testing.T) {
	createdAt := time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name              string
		context           devcontext.Context
		wantID            devcontext.ID
		wantDisplayName   string
		wantProviderState map[provider.ID]bool
		wantProviderOpts  map[provider.ID]provider.Options
		wantMetadata      devcontext.Metadata
	}{
		{
			name: "personal",
			context: devcontext.Context{
				ID:   devcontext.MustID("personal"),
				Name: "Personal",
				Tool: codingtool.DefaultConfig(),
				Providers: provider.Configs{
					"claude": {Enabled: true},
					"codex":  {Enabled: true},
				},
				Metadata: devcontext.Metadata{
					"kind": "default",
				},
				CreatedAt: createdAt,
			},
			wantID:          devcontext.MustID("personal"),
			wantDisplayName: "Personal",
			wantProviderState: map[provider.ID]bool{
				"claude": true,
				"codex":  true,
			},
			wantMetadata: devcontext.Metadata{
				"kind": "default",
			},
		},
		{
			name: "company",
			context: devcontext.Context{
				ID:   devcontext.MustID("company"),
				Name: "Company",
				Tool: codingtool.DefaultConfig(),
				Providers: provider.Configs{
					"claude": {Enabled: true},
					"codex":  {Enabled: true},
				},
				Metadata: devcontext.Metadata{
					"kind": "default",
				},
				CreatedAt: createdAt,
			},
			wantID:          devcontext.MustID("company"),
			wantDisplayName: "Company",
			wantProviderState: map[provider.ID]bool{
				"claude": true,
				"codex":  true,
			},
			wantMetadata: devcontext.Metadata{
				"kind": "default",
			},
		},
		{
			name: "client",
			context: devcontext.Context{
				ID:   devcontext.MustID("client-a"),
				Name: "Client A",
				Tool: codingtool.DefaultConfig(),
				Providers: provider.Configs{
					"claude": {
						Enabled: true,
						Options: provider.Options{
							"profile": "client-a",
						},
					},
					"codex": {Enabled: false},
					"future-provider": {
						Enabled: true,
						Options: provider.Options{
							"mode": "sandbox",
						},
					},
				},
				Metadata: devcontext.Metadata{
					"owner": "client-a",
				},
				CreatedAt: createdAt,
			},
			wantID:          devcontext.MustID("client-a"),
			wantDisplayName: "Client A",
			wantProviderState: map[provider.ID]bool{
				"claude":          true,
				"codex":           false,
				"future-provider": true,
			},
			wantProviderOpts: map[provider.ID]provider.Options{
				"claude": {
					"profile": "client-a",
				},
				"future-provider": {
					"mode": "sandbox",
				},
			},
			wantMetadata: devcontext.Metadata{
				"owner": "client-a",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.context.ID != tt.wantID {
				t.Fatalf("context ID = %q, want %q", tt.context.ID, tt.wantID)
			}
			if tt.context.Name != tt.wantDisplayName {
				t.Fatalf("context name = %q, want %q", tt.context.Name, tt.wantDisplayName)
			}
			if tt.context.Tool.Type != codingtool.TypeVSCode {
				t.Fatalf("editor type = %q, want %q", tt.context.Tool.Type, codingtool.TypeVSCode)
			}
			if len(tt.context.Providers) != len(tt.wantProviderState) {
				t.Fatalf("provider count = %d, want %d", len(tt.context.Providers), len(tt.wantProviderState))
			}
			for providerID, wantEnabled := range tt.wantProviderState {
				providerConfig, ok := tt.context.Providers[providerID]
				if !ok {
					t.Fatalf("provider %q is missing", providerID)
				}
				if providerConfig.Enabled != wantEnabled {
					t.Fatalf("provider %q enabled = %t, want %t", providerID, providerConfig.Enabled, wantEnabled)
				}
			}
			for providerID, wantOptions := range tt.wantProviderOpts {
				providerConfig := tt.context.Providers[providerID]
				if len(providerConfig.Options) != len(wantOptions) {
					t.Fatalf("provider %q option count = %d, want %d", providerID, len(providerConfig.Options), len(wantOptions))
				}
				for key, wantValue := range wantOptions {
					if providerConfig.Options[key] != wantValue {
						t.Fatalf("provider %q option %q = %q, want %q", providerID, key, providerConfig.Options[key], wantValue)
					}
				}
			}
			if len(tt.context.Metadata) != len(tt.wantMetadata) {
				t.Fatalf("metadata count = %d, want %d", len(tt.context.Metadata), len(tt.wantMetadata))
			}
			for key, wantValue := range tt.wantMetadata {
				if tt.context.Metadata[key] != wantValue {
					t.Fatalf("metadata %q = %q, want %q", key, tt.context.Metadata[key], wantValue)
				}
			}
			if !tt.context.CreatedAt.Equal(createdAt) {
				t.Fatalf("created at = %s, want %s", tt.context.CreatedAt, createdAt)
			}
		})
	}
}
