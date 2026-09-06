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
			Category: historyEventCategory(event.Name), Timestamp: event.Timestamp.UTC(), ProjectPath: event.ProjectPath,
			ContextID: event.ContextID, ToolID: event.ToolID, Message: historyEventMessage(event),
		})
	}
	return HistoryState{Entries: entries}, nil
}

func historyEventCategory(name devlog.EventName) HistoryCategory {
	switch name {
	case devlog.EventContextResolution, devlog.EventLaunchSucceeded:
		return HistoryCategoryLaunch
	case devlog.EventLaunchMissingEditor, devlog.EventLaunchConfigError, devlog.EventLaunchProviderMissing, devlog.EventLaunchProcessFailure:
		return HistoryCategoryWarning
	case devlog.EventProjectBound, devlog.EventProjectUnbound, "project_binding_changed":
		return HistoryCategoryBinding
	case devlog.EventRepairCompleted, devlog.EventProviderReset:
		return HistoryCategoryRepair
	case devlog.EventProviderAuthenticated, "provider_connected":
		return HistoryCategoryAuthentication
	case devlog.EventWorkspaceStopped, "environment_stopped":
		return HistoryCategoryWorkspace
	case devlog.EventContextOverrideAccepted:
		return HistoryCategoryOverride
	default:
		return HistoryCategoryContext
	}
}

func historyEventMessage(event devlog.Event) string {
	switch event.Name {
	case devlog.EventContextResolution:
		return "Launch context resolved."
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
	case devlog.EventContextCreated:
		return "Context created."
	case devlog.EventContextUpdated:
		return "Context updated."
	case devlog.EventContextArchived:
		return "Context archived."
	case devlog.EventContextRestored:
		return "Context restored."
	case devlog.EventContextDeleted:
		return "Context deleted."
	case devlog.EventProviderAuthenticated, "provider_connected":
		return "Provider authenticated."
	case devlog.EventProviderReset:
		return "Provider storage reset."
	case devlog.EventRepairCompleted:
		return "Repair completed."
	case devlog.EventProjectBound:
		return "Project bound to context."
	case devlog.EventProjectUnbound:
		return "Project binding removed."
	case "project_binding_changed":
		return "Project context binding changed."
	case devlog.EventWorkspaceStopped, "environment_stopped":
		return "Workspace stopped."
	case devlog.EventContextOverrideAccepted:
		return "Context override accepted."
	default:
		return "Activity recorded."
	}
}
