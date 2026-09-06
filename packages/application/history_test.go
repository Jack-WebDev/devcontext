package application

import (
	"testing"

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
