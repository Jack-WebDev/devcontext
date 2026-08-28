package application

import (
	"path/filepath"

	devlog "devctx/packages/core/logging"
)

func (s *Service) getHistory() (HistoryState, error) {
	homeDir, err := s.dependencies.Paths.DevContextHomeDir()
	if err != nil {
		return HistoryState{}, err
	}
	events, err := devlog.ReadLocalEvents(filepath.Join(homeDir, "logs"))
	if err != nil {
		return HistoryState{}, err
	}
	entries := make([]HistoryEntry, 0, len(events))
	for _, event := range events {
		entries = append(entries, HistoryEntry{
			Event: string(event.Name), Timestamp: event.Timestamp.UTC(), ProjectPath: event.ProjectPath,
			ContextID: event.ContextID, ToolID: event.ToolID, Message: historyEventMessage(event),
		})
	}
	return HistoryState{Entries: entries}, nil
}

func historyEventMessage(event devlog.Event) string {
	switch event.Name {
	case devlog.EventLaunchSucceeded:
		return "Launch succeeded."
	case devlog.EventLaunchMissingEditor:
		return "Launch could not find the selected coding tool."
	case devlog.EventLaunchProviderMissing:
		return "Launch could not prepare an enabled provider."
	case devlog.EventLaunchProcessFailure:
		return "Launch could not start the selected coding tool."
	case devlog.EventLaunchConfigError:
		return "Launch configuration needs attention."
	default:
		return "Launch context was resolved."
	}
}
