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
		{name: "context configuration", event: devlog.EventContextCreated, want: HistoryCategoryConfiguration},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := historyEventCategory(tt.event); got != tt.want {
				t.Fatalf("history event category = %q, want %q", got, tt.want)
			}
		})
	}
}
