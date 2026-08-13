package context_test

import (
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
)

func TestContextModelConstructsNamedDevelopmentIdentities(t *testing.T) {
	createdAt := time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)

	tests := []struct {
		name              string
		context           devcontext.Context
		wantID            devcontext.ID
		wantDisplayName   string
		wantProviderState map[devcontext.ProviderID]bool
		wantMetadata      devcontext.Metadata
	}{
		{
			name: "personal",
			context: devcontext.Context{
				ID:   devcontext.MustID("personal"),
				Name: "Personal",
				Editor: devcontext.EditorConfig{
					Type: "vscode",
				},
				Providers: map[devcontext.ProviderID]devcontext.ProviderConfig{
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
			wantProviderState: map[devcontext.ProviderID]bool{
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
				Editor: devcontext.EditorConfig{
					Type: "vscode",
				},
				Providers: map[devcontext.ProviderID]devcontext.ProviderConfig{
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
			wantProviderState: map[devcontext.ProviderID]bool{
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
				Editor: devcontext.EditorConfig{
					Type: "vscode",
				},
				Providers: map[devcontext.ProviderID]devcontext.ProviderConfig{
					"claude":          {Enabled: true},
					"codex":           {Enabled: false},
					"future-provider": {Enabled: true},
				},
				Metadata: devcontext.Metadata{
					"owner": "client-a",
				},
				CreatedAt: createdAt,
			},
			wantID:          devcontext.MustID("client-a"),
			wantDisplayName: "Client A",
			wantProviderState: map[devcontext.ProviderID]bool{
				"claude":          true,
				"codex":           false,
				"future-provider": true,
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
			if tt.context.Editor.Type != "vscode" {
				t.Fatalf("editor type = %q, want %q", tt.context.Editor.Type, "vscode")
			}
			if len(tt.context.Providers) != len(tt.wantProviderState) {
				t.Fatalf("provider count = %d, want %d", len(tt.context.Providers), len(tt.wantProviderState))
			}
			for providerID, wantEnabled := range tt.wantProviderState {
				provider, ok := tt.context.Providers[providerID]
				if !ok {
					t.Fatalf("provider %q is missing", providerID)
				}
				if provider.Enabled != wantEnabled {
					t.Fatalf("provider %q enabled = %t, want %t", providerID, provider.Enabled, wantEnabled)
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
