package logging_test

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"devctx/packages/core/filesystem"
	devlog "devctx/packages/core/logging"
)

func TestLocalLoggerAppendsParseableRecordsWithRestrictedPermissions(t *testing.T) {
	logsDir := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)
	logger := devlog.NewLocalLogger(logsDir, filesystem.NewDefaultStoragePermissions(), func() time.Time {
		return now
	})

	events := []devlog.Event{
		{Name: devlog.EventContextResolution, ProjectPath: "/work/app", ContextID: "personal"},
		{Name: devlog.EventLaunchSucceeded, ProjectPath: "/work/app", ContextID: "personal", EditorID: "vscode"},
	}
	for _, event := range events {
		if err := logger.Record(event); err != nil {
			t.Fatalf("record event: %v", err)
		}
	}

	path := filepath.Join(logsDir, devlog.DefaultFileName)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var records []devlog.Event
	for scanner.Scan() {
		var record devlog.Event
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			t.Fatalf("decode record %q: %v", scanner.Text(), err)
		}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan log file: %v", err)
	}
	if len(records) != len(events) {
		t.Fatalf("record count = %d, want %d", len(records), len(events))
	}
	for _, record := range records {
		if !record.Timestamp.Equal(now) {
			t.Fatalf("timestamp = %s, want %s", record.Timestamp, now)
		}
	}

	if runtime.GOOS != "windows" {
		assertMode(t, logsDir, filesystem.RestrictedDirectoryMode)
		assertMode(t, path, filesystem.RestrictedFileMode)
	}
}

func TestNoopLoggerDropsEvents(t *testing.T) {
	if err := (devlog.NoopLogger{}).Record(devlog.Event{Name: devlog.EventLaunchSucceeded}); err != nil {
		t.Fatalf("record noop event: %v", err)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %q: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("mode for %q = %o, want %o", path, got, want)
	}
}
