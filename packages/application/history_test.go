package application

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	devlog "devctx/packages/core/logging"
)

func TestHistoryEventCategoryUsesBackendOwnedFilters(t *testing.T) {
	tests := []struct {
		name  string
		event devlog.EventName
		want  HistoryCategory
	}{
		{name: "successful launch", event: devlog.EventLaunchSucceeded, want: HistoryCategoryLaunch},
		{name: "launch warning", event: devlog.EventLaunchProviderMissing, want: HistoryCategoryWarning},
		{name: "context change", event: devlog.EventContextCreated, want: HistoryCategoryContext},
		{name: "project binding", event: devlog.EventProjectBound, want: HistoryCategoryBinding},
		{name: "repair", event: devlog.EventRepairCompleted, want: HistoryCategoryRepair},
		{name: "authentication", event: devlog.EventProviderAuthenticated, want: HistoryCategoryAuthentication},
		{name: "workspace", event: devlog.EventWorkspaceStopped, want: HistoryCategoryWorkspace},
		{name: "override", event: devlog.EventContextOverrideAccepted, want: HistoryCategoryOverride},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := historyEventCategory(tt.event); got != tt.want {
				t.Fatalf("history event category = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetHistoryExcludesCredentialsAndEnvironmentValues(t *testing.T) {
	fixture := newApplicationFixture(t)
	secret := "history-private-token"
	homeDir, err := fixture.paths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	logsDir := filepath.Join(homeDir, "logs")
	logger := devlog.NewLocalLogger(logsDir, fixture.storagePermissions, func() time.Time {
		return fixture.now
	})
	if err := logger.Record(devlog.NewEvent(devlog.EventInput{
		Name:        devlog.EventLaunchProcessFailure,
		Timestamp:   fixture.now,
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
		ToolID:      "fake-editor",
		Err:         fmt.Errorf("launch failed: API_TOKEN=%s", secret),
		KnownEnvironment: []string{
			"API_TOKEN=" + secret,
			"PRIVATE_WORKSPACE_VALUE=" + secret,
		},
	})); err != nil {
		t.Fatalf("record history event: %v", err)
	}

	history, appErr := fixture.service().GetHistory()
	if appErr != nil {
		t.Fatalf("get history: %v", appErr)
	}
	data, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}
	for _, privateValue := range []string{secret, "API_TOKEN", "PRIVATE_WORKSPACE_VALUE"} {
		if strings.Contains(string(data), privateValue) {
			t.Fatalf("history leaked %q: %s", privateValue, data)
		}
	}
	if len(history.Entries) != 1 || history.Entries[0].Message != "Launch could not start the selected coding tool." {
		t.Fatalf("history entries = %#v, want safe launch outcome", history.Entries)
	}
}

func TestHistoryEventMessageDescribesKnownUserOutcomes(t *testing.T) {
	tests := []struct {
		event devlog.EventName
		want  string
	}{
		{event: devlog.EventContextResolution, want: "Launch context resolved."},
		{event: devlog.EventProjectBound, want: "Project bound to context."},
		{event: devlog.EventProjectUnbound, want: "Project binding removed."},
		{event: devlog.EventProviderAuthenticated, want: "Provider authenticated."},
		{event: devlog.EventWorkspaceStopped, want: "Workspace stopped."},
		{event: devlog.EventContextOverrideAccepted, want: "Context override accepted."},
		{event: "project_binding_changed", want: "Project context binding changed."},
	}

	for _, tt := range tests {
		if got := historyEventMessage(devlog.Event{Name: tt.event}); got != tt.want {
			t.Fatalf("history message for %q = %q, want %q", tt.event, got, tt.want)
		}
	}
}
