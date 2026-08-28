package codingtool

import (
	"fmt"
	"path/filepath"
	"strings"
)

// StatusDataConsumer is an optional coding-tool integration contract for
// consuming Dev Context's locally exported status document. The document is
// written by Dev Context into the selected tool's isolated storage; an
// integration never receives credentials, commands, or environment values.
//
// Implementations return a file name, not a path, so status data cannot be
// exported outside the tool-owned storage directory.
type StatusDataConsumer interface {
	StatusDataFileName() string
}

// StatusDataPath resolves a consumer's status file inside its tool-owned
// storage directory.
func StatusDataPath(paths ContextPaths, consumer StatusDataConsumer) (string, error) {
	if consumer == nil {
		return "", fmt.Errorf("coding-tool status data consumer is nil")
	}
	fileName := strings.TrimSpace(consumer.StatusDataFileName())
	if fileName == "" || fileName == "." || filepath.Base(fileName) != fileName {
		return "", fmt.Errorf("coding-tool status data file name %q is invalid", fileName)
	}
	return filepath.Join(paths.StorageDir, fileName), nil
}
