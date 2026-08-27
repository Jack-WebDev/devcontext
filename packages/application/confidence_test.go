package application_test

import (
	"encoding/json"
	"testing"

	"devctx/packages/application"
)

func TestLaunchConfidenceStatusVariantsSerialize(t *testing.T) {
	tests := []struct {
		name   string
		status application.LaunchConfidenceStatus
		want   string
	}{
		{
			name:   "ready",
			status: application.LaunchConfidenceReady,
			want:   `{"status":"ready"}`,
		},
		{
			name:   "needs attention",
			status: application.LaunchConfidenceNeedsAttention,
			want:   `{"status":"needs_attention"}`,
		},
		{
			name:   "blocked",
			status: application.LaunchConfidenceBlocked,
			want:   `{"status":"blocked"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.status.Valid() {
				t.Fatalf("status %q is not valid", tt.status)
			}

			data, err := json.Marshal(struct {
				Status application.LaunchConfidenceStatus `json:"status"`
			}{Status: tt.status})
			if err != nil {
				t.Fatalf("marshal status: %v", err)
			}
			if string(data) != tt.want {
				t.Fatalf("json = %s, want %s", data, tt.want)
			}
		})
	}
}

func TestLaunchConfidenceCheckAliasesCoreContract(t *testing.T) {
	check := application.LaunchConfidenceCheck{
		Component:  application.LaunchConfidenceCheckProvider,
		ProviderID: "claude",
		Severity:   application.LaunchConfidenceBlocked,
		Label:      "Claude",
		Message:    "Claude cannot be checked for this context.",
		ActionHint: "Open diagnostics.",
	}

	if !check.Valid() {
		t.Fatalf("check is not valid: %#v", check)
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal check: %v", err)
	}

	want := `{"component":"provider","providerId":"claude","severity":"blocked","label":"Claude","message":"Claude cannot be checked for this context.","actionHint":"Open diagnostics."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
}

func TestProviderReadinessStateVariantsSerialize(t *testing.T) {
	tests := []struct {
		name  string
		state application.ProviderReadinessState
		want  string
	}{
		{
			name:  "ready",
			state: application.ProviderReadinessReady,
			want:  `{"state":"ready"}`,
		},
		{
			name:  "not configured",
			state: application.ProviderReadinessNotConfigured,
			want:  `{"state":"not_configured"}`,
		},
		{
			name:  "directory missing",
			state: application.ProviderReadinessDirectoryMissing,
			want:  `{"state":"directory_missing"}`,
		},
		{
			name:  "unavailable",
			state: application.ProviderReadinessUnavailable,
			want:  `{"state":"unavailable"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.state.Valid() {
				t.Fatalf("state %q is not valid", tt.state)
			}

			data, err := json.Marshal(struct {
				State application.ProviderReadinessState `json:"state"`
			}{State: tt.state})
			if err != nil {
				t.Fatalf("marshal state: %v", err)
			}
			if string(data) != tt.want {
				t.Fatalf("json = %s, want %s", data, tt.want)
			}
		})
	}

	if application.ProviderReadinessState("configured").Valid() {
		t.Fatal("storage-oriented configured state is valid as an API provider readiness state")
	}
}

func TestProviderSetupActionSerializesBoundedState(t *testing.T) {
	action := application.ProviderSetupAction{
		State:   application.ProviderSetupOpenAndConfigure,
		Label:   "Open and configure",
		Message: "Codex needs to be configured for this context.",
	}
	if !action.State.Valid() {
		t.Fatalf("setup state %q is invalid", action.State)
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("marshal setup action: %v", err)
	}
	want := `{"state":"open_and_configure","label":"Open and configure","message":"Codex needs to be configured for this context."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
	if application.ProviderSetupState("sign_in").Valid() {
		t.Fatal("unknown provider setup state is valid")
	}
}

func TestProviderIdentityStateVariantsSerialize(t *testing.T) {
	tests := []struct {
		name     string
		identity application.ProviderIdentityState
		want     string
	}{
		{
			name: "verified identity fields",
			identity: application.ProviderIdentityState{
				Status: application.ProviderIdentityVerified,
				Fields: []application.ProviderMetadataField{{Label: "Email", Value: "user@example.com"}, {Label: "Plan", Value: "Business"}, {Label: "Account", Value: "acct_123"}},
			},
			want: `{"status":"verified","fields":[{"label":"Email","value":"user@example.com"},{"label":"Plan","value":"Business"},{"label":"Account","value":"acct_123"}]}`,
		},
		{
			name: "unavailable",
			identity: application.ProviderIdentityState{
				Status:  application.ProviderIdentityUnavailable,
				Message: "Account identity unavailable.",
			},
			want: `{"status":"unavailable","message":"Account identity unavailable."}`,
		},
		{
			name:     "none",
			identity: application.ProviderIdentityState{Status: application.ProviderIdentityNone},
			want:     `{"status":"none"}`,
		},
		{
			name: "mismatch evidence",
			identity: application.ProviderIdentityState{
				Status:  application.ProviderIdentityMismatchEvidence,
				Message: "Different account identity detected.",
			},
			want: `{"status":"mismatch_evidence","message":"Different account identity detected."}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.identity.Status.Valid() {
				t.Fatalf("identity status %q is not valid", tt.identity.Status)
			}

			data, err := json.Marshal(tt.identity)
			if err != nil {
				t.Fatalf("marshal identity: %v", err)
			}
			if string(data) != tt.want {
				t.Fatalf("json = %s, want %s", data, tt.want)
			}
		})
	}

	if application.ProviderIdentityStatus("configured").Valid() {
		t.Fatal("provider readiness value is valid as an identity status")
	}
}
