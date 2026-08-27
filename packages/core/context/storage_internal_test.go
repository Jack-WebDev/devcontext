package context

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/provider"
)

func TestWriteContextFileAtomicallyFailureLeavesPreviousContextLoadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, contextConfigFileName)
	previous := testStoredContext("personal", "Personal")
	injectedErr := errors.New("injected write failure")

	data, err := EncodeContextTOML(previous)
	if err != nil {
		t.Fatalf("encode previous context: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write previous context: %v", err)
	}

	err = writeContextFileAtomically(path, func(file *os.File) error {
		if _, writeErr := file.WriteString("id = "); writeErr != nil {
			return writeErr
		}
		return injectedErr
	})
	if !errors.Is(err, injectedErr) {
		t.Fatalf("error = %v, want %v", err, injectedErr)
	}

	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read context after failure: %v", err)
	}
	decoded, err := DecodeContextTOML(data, previous.ID)
	if err != nil {
		t.Fatalf("decode context after failure: %v", err)
	}
	if !reflect.DeepEqual(decoded, previous) {
		t.Fatalf("context after failure = %#v, want %#v", decoded, previous)
	}

	tempFiles, err := filepath.Glob(filepath.Join(dir, "."+contextConfigFileName+".tmp-*"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(tempFiles) != 0 {
		t.Fatalf("temp file count = %d, want 0: %v", len(tempFiles), tempFiles)
	}
}

func testStoredContext(id string, name string) Context {
	return Context{
		ID:   MustID(id),
		Name: name,
		Tool: codingtool.DefaultLaunchTarget(),
		Providers: provider.Configs{
			"claude": {Enabled: true},
			"codex":  {Enabled: true},
		},
		Metadata: Metadata{
			"kind": "test",
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}
