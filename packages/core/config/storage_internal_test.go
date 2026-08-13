package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicallyFailureLeavesPreviousFileReadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	previous := []byte("previous config")
	injectedErr := errors.New("injected write failure")

	if err := os.WriteFile(path, previous, 0o600); err != nil {
		t.Fatalf("write previous config: %v", err)
	}

	err := writeFileAtomically(path, func(file *os.File) error {
		if _, writeErr := file.WriteString("partial config"); writeErr != nil {
			return writeErr
		}
		return injectedErr
	})
	if !errors.Is(err, injectedErr) {
		t.Fatalf("error = %v, want %v", err, injectedErr)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config after failure: %v", err)
	}
	if string(data) != string(previous) {
		t.Fatalf("config after failure = %q, want %q", data, previous)
	}

	tempFiles, err := filepath.Glob(filepath.Join(dir, ".config.toml.tmp-*"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(tempFiles) != 0 {
		t.Fatalf("temp file count = %d, want 0: %v", len(tempFiles), tempFiles)
	}
}
