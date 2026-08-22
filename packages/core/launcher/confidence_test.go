package launcher_test

import (
	"encoding/json"
	"testing"

	"devctx/packages/core/launcher"
)

func TestConfidenceStatusVariantsSerialize(t *testing.T) {
	tests := []struct {
		name   string
		status launcher.ConfidenceStatus
		want   string
	}{
		{
			name:   "ready",
			status: launcher.ConfidenceReady,
			want:   `{"status":"ready"}`,
		},
		{
			name:   "needs attention",
			status: launcher.ConfidenceNeedsAttention,
			want:   `{"status":"needs_attention"}`,
		},
		{
			name:   "blocked",
			status: launcher.ConfidenceBlocked,
			want:   `{"status":"blocked"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.status.Valid() {
				t.Fatalf("status %q is not valid", tt.status)
			}

			data, err := json.Marshal(struct {
				Status launcher.ConfidenceStatus `json:"status"`
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

func TestConfidenceStatusRejectsUnknownValues(t *testing.T) {
	if launcher.ConfidenceStatus("configured").Valid() {
		t.Fatal("provider status value is valid as launch confidence")
	}
	if launcher.ConfidenceStatus("healthy").Valid() {
		t.Fatal("unapproved UI synonym is valid as launch confidence")
	}
}

func TestConfidenceCheckComponentsValidate(t *testing.T) {
	tests := []launcher.ConfidenceCheckComponent{
		launcher.ConfidenceCheckClaude,
		launcher.ConfidenceCheckCodex,
		launcher.ConfidenceCheckVSCode,
		launcher.ConfidenceCheckIsolation,
	}

	for _, component := range tests {
		if !component.Valid() {
			t.Fatalf("component %q is not valid", component)
		}
	}

	if launcher.ConfidenceCheckComponent("editor").Valid() {
		t.Fatal("unknown confidence check component is valid")
	}
}

func TestConfidenceCheckSerializesWithOptionalActionHint(t *testing.T) {
	check := launcher.ConfidenceCheck{
		Component:  launcher.ConfidenceCheckCodex,
		Severity:   launcher.ConfidenceNeedsAttention,
		Label:      "Codex",
		Message:    "Codex is not authenticated for this context.",
		ActionHint: "Sign in to Codex for this context.",
	}

	if !check.Valid() {
		t.Fatalf("check is not valid: %#v", check)
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal check: %v", err)
	}

	want := `{"component":"codex","severity":"needs_attention","label":"Codex","message":"Codex is not authenticated for this context.","actionHint":"Sign in to Codex for this context."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
}

func TestConfidenceCheckOmitsEmptyActionHint(t *testing.T) {
	check := launcher.ConfidenceCheck{
		Component: launcher.ConfidenceCheckIsolation,
		Severity:  launcher.ConfidenceReady,
		Label:     "Isolation",
		Message:   "The context isolation directories are ready.",
	}

	if !check.Valid() {
		t.Fatalf("check is not valid: %#v", check)
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal check: %v", err)
	}

	want := `{"component":"isolation","severity":"ready","label":"Isolation","message":"The context isolation directories are ready."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
}

func TestConfidenceCheckRejectsIncompleteValues(t *testing.T) {
	tests := []struct {
		name  string
		check launcher.ConfidenceCheck
	}{
		{
			name: "unknown component",
			check: launcher.ConfidenceCheck{
				Component: "editor",
				Severity:  launcher.ConfidenceReady,
				Label:     "VS Code",
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "unknown severity",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  "configured",
				Label:     "VS Code",
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "missing label",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  launcher.ConfidenceReady,
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "missing message",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  launcher.ConfidenceReady,
				Label:     "VS Code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.check.Valid() {
				t.Fatalf("check is valid: %#v", tt.check)
			}
		})
	}
}
