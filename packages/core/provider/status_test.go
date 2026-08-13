package provider_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"devctx/packages/core/provider"
)

func TestProviderStatusVariantsSerialize(t *testing.T) {
	tests := []struct {
		name   string
		status provider.Status
		want   string
	}{
		{
			name:   "ready",
			status: provider.ReadyStatus(),
			want:   `{"state":"ready"}`,
		},
		{
			name:   "not configured",
			status: provider.NotConfiguredStatus("run the provider setup locally"),
			want:   `{"state":"not_configured","explanation":"run the provider setup locally"}`,
		},
		{
			name:   "directory missing",
			status: provider.DirectoryMissingStatus("context provider directory is missing"),
			want:   `{"state":"directory_missing","explanation":"context provider directory is missing"}`,
		},
		{
			name:   "unavailable",
			status: provider.UnavailableStatus("provider command was not found"),
			want:   `{"state":"unavailable","explanation":"provider command was not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.status.State.Valid() {
				t.Fatalf("state %q is not valid", tt.status.State)
			}

			data, err := json.Marshal(tt.status)
			if err != nil {
				t.Fatalf("marshal status: %v", err)
			}
			if string(data) != tt.want {
				t.Fatalf("json = %s, want %s", data, tt.want)
			}

			var decoded provider.Status
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal status: %v", err)
			}
			if !reflect.DeepEqual(decoded, tt.status) {
				t.Fatalf("decoded status = %#v, want %#v", decoded, tt.status)
			}
		})
	}
}

func TestProviderStatusStateRejectsUnknownValues(t *testing.T) {
	if provider.StatusState("expired_subscription").Valid() {
		t.Fatal("unknown provider status state is valid")
	}
}
