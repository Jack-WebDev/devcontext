package logging

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ReadLocalEvents reads the allowlisted local event log in newest-first order.
// A missing log is a valid empty history.
func ReadLocalEvents(logsDir string) ([]Event, error) {
	file, err := os.Open(filepath.Join(logsDir, DefaultFileName))
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
	sort.SliceStable(events, func(i, j int) bool { return events[i].Timestamp.After(events[j].Timestamp) })
	return events, nil
}
