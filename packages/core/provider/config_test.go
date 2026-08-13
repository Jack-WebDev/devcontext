package provider_test

import (
	"testing"

	"devctx/packages/core/provider"
)

func TestProviderConfigsStoreEnabledStateAndOptionsByProviderID(t *testing.T) {
	configs := provider.Configs{
		"claude": {
			Enabled: true,
			Options: provider.Options{
				"profile": "personal",
			},
		},
		"codex": {
			Enabled: true,
		},
		"future-provider": {
			Enabled: false,
			Options: provider.Options{
				"mode": "sandbox",
			},
		},
	}

	tests := []struct {
		id          provider.ID
		wantEnabled bool
		optionKey   string
		optionValue string
	}{
		{
			id:          "claude",
			wantEnabled: true,
			optionKey:   "profile",
			optionValue: "personal",
		},
		{
			id:          "codex",
			wantEnabled: true,
		},
		{
			id:          "future-provider",
			wantEnabled: false,
			optionKey:   "mode",
			optionValue: "sandbox",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.id), func(t *testing.T) {
			config, ok := configs[tt.id]
			if !ok {
				t.Fatalf("provider %q is missing", tt.id)
			}
			if config.Enabled != tt.wantEnabled {
				t.Fatalf("enabled = %t, want %t", config.Enabled, tt.wantEnabled)
			}
			if tt.optionKey == "" {
				return
			}
			if config.Options[tt.optionKey] != tt.optionValue {
				t.Fatalf("option %q = %q, want %q", tt.optionKey, config.Options[tt.optionKey], tt.optionValue)
			}
		})
	}
}
