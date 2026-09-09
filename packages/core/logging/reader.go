package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"devctx/packages/core/filesystem"
)

// ReadLocalEvents reads the allowlisted local event log in newest-first order.
// A missing log is a valid empty history.
func ReadLocalEvents(logsDir string) ([]Event, error) {
	events, err := readEvents(filepath.Join(logsDir, DefaultFileName))
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp.After(events[j].Timestamp) })
	return events, nil
}

// RemoveEventsForProject deletes local activity records for one forgotten
// project. It does not affect records for other projects.
func RemoveEventsForProject(logsDir, projectPath string, permissions filesystem.StoragePermissions) error {
	if projectPath == "" {
		return fmt.Errorf("remove project events: project path is required")
	}

	path := filepath.Join(logsDir, DefaultFileName)
	return filesystem.WithExclusiveFileLock(path, func() error {
		events, err := readEvents(path)
		if err != nil {
			return err
		}
		remaining := events[:0]
		for _, event := range events {
			if event.ProjectPath != projectPath {
				remaining = append(remaining, event)
			}
		}
		if len(remaining) == len(events) {
			return nil
		}
		return writeEvents(path, remaining, permissions)
	})
}

func readEvents(path string) ([]Event, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("open event log: %w", err)
	}
	defer file.Close()

	events := make([]Event, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, fmt.Errorf("decode event log: %w", err)
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read event log: %w", err)
	}
	return events, nil
}
