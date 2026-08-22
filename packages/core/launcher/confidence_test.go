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
