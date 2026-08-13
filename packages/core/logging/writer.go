package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devctx/packages/core/filesystem"
)

const (
	// DefaultFileName is the JSONL launch troubleshooting log file.
	DefaultFileName = "launch.jsonl"
)

// Logger records approved local troubleshooting events.
type Logger interface {
	Record(Event) error
}

// NoopLogger intentionally drops events.
type NoopLogger struct{}

var _ Logger = NoopLogger{}

// Record implements Logger.
func (NoopLogger) Record(Event) error {
	return nil
}

// LocalLogger appends one JSON event per line in the Dev Context logs directory.
type LocalLogger struct {
	LogsDir     string
	FileName    string
	Permissions filesystem.StoragePermissions
	Now         func() time.Time
	Disabled    bool
}

var _ Logger = LocalLogger{}

// NewLocalLogger creates the default local launch event logger.
func NewLocalLogger(logsDir string, permissions filesystem.StoragePermissions, now func() time.Time) LocalLogger {
	return LocalLogger{
		LogsDir:     logsDir,
		FileName:    DefaultFileName,
		Permissions: permissions,
		Now:         now,
	}
}

// Record appends one schema-approved event. The caller decides whether logging
// failures should affect the user-facing operation.
func (l LocalLogger) Record(event Event) error {
	if l.Disabled {
		return nil
	}
	if l.LogsDir == "" {
		return fmt.Errorf("missing logs directory")
	}

	permissions := l.Permissions
	if permissions == nil {
		permissions = filesystem.NewDefaultStoragePermissions()
	}
	if err := os.MkdirAll(l.LogsDir, permissions.DirectoryMode()); err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("create directory", l.LogsDir, err); wrapped != err {
			return fmt.Errorf("create logs directory %q: %w", l.LogsDir, wrapped)
		}
		return fmt.Errorf("create logs directory %q: %w", l.LogsDir, err)
	}
	if err := permissions.ApplyDirectory(l.LogsDir); err != nil {
		return err
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = l.now()
	}
	event.Error = SanitizeText(event.Error, nil)

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("encode log event: %w", err)
	}

	path := filepath.Join(l.LogsDir, l.fileName())
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, permissions.FileMode())
	if err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("open file", path, err); wrapped != err {
			return fmt.Errorf("open log file %q: %w", path, wrapped)
		}
		return fmt.Errorf("open log file %q: %w", path, err)
	}
	defer file.Close()

	if err := permissions.ApplyFile(path); err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		if wrapped := filesystem.WrapStoragePermissionError("append file", path, err); wrapped != err {
			return fmt.Errorf("append log event to %q: %w", path, wrapped)
		}
		return fmt.Errorf("append log event to %q: %w", path, err)
	}
	return nil
}

func (l LocalLogger) fileName() string {
	if l.FileName != "" {
		return l.FileName
	}
	return DefaultFileName
}

func (l LocalLogger) now() time.Time {
	if l.Now != nil {
		return l.Now()
	}
	return time.Now()
}
