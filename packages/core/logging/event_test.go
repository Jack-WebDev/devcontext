package logging_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	devlog "devctx/packages/core/logging"
)

func TestEventSerializesOnlyApprovedFields(t *testing.T) {
	event := devlog.NewEvent(devlog.EventInput{
		Name:             devlog.EventLaunchProcessFailure,
		Timestamp:        time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
		ProjectPath:      "/work/app",
		ContextID:        "personal",
		EditorID:         "vscode",
		ResolutionSource: "explicit",
		Err:              fmt.Errorf("process failed: API_TOKEN=secret-token"),
		KnownEnvironment: []string{"API_TOKEN=secret-token"},
	})

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	if strings.Contains(string(data), "secret-token") {
		t.Fatalf("event leaked secret: %s", data)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("decode event: %v", err)
	}

	allowed := map[string]bool{
		"event":             true,
		"timestamp":         true,
		"project_path":      true,
		"context_id":        true,
		"editor_id":         true,
		"resolution_source": true,
		"error_category":    true,
		"error":             true,
	}
	for field := range fields {
		if !allowed[field] {
			t.Fatalf("unexpected serialized field %q in %s", field, data)
		}
	}
	if _, ok := fields["environment"]; ok {
		t.Fatalf("event includes environment field: %s", data)
	}
}

func TestSanitizeErrorRemovesSeededSecrets(t *testing.T) {
	seeds := []string{
		"known-env-token",
		"assignment-secret",
		"bearer-secret",
		"cookie-secret",
		"oauth-code-secret",
		"oauth-assignment-secret",
	}
	err := fmt.Errorf(
		"failed with %s API_TOKEN=%s authorization: Bearer %s oauth_code=%s cookie: sid=%s https://example.test/callback?code=%s",
		seeds[0],
		seeds[1],
		seeds[2],
		seeds[5],
		seeds[3],
		seeds[4],
	)

	got := devlog.SanitizeError(err, []string{"DEVCTX_API_TOKEN=" + seeds[0]})
	for _, seed := range seeds {
		if strings.Contains(got, seed) {
			t.Fatalf("sanitized error leaked %q in %q", seed, got)
		}
	}
	if !strings.Contains(got, devlog.RedactedValue) {
		t.Fatalf("sanitized error = %q, want redaction marker", got)
	}
}
