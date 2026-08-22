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
