package project

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
)

func TestWriteProjectBindingsFileAtomicallyFailureLeavesPreviousCollectionLoadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "projects.toml")
	previous := []Binding{
		testProjectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	}
	injectedErr := errors.New("injected write failure")

	data, err := EncodeProjectBindingsTOML(previous)
	if err != nil {
		t.Fatalf("encode previous bindings: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write previous bindings: %v", err)
	}

	err = writeProjectBindingsFileAtomically(path, func(file *os.File) error {
		if _, writeErr := file.WriteString("[[projects]]\npath = "); writeErr != nil {
			return writeErr
		}
		return injectedErr
	})
	if !errors.Is(err, injectedErr) {
		t.Fatalf("error = %v, want %v", err, injectedErr)
	}

	data, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read bindings after failure: %v", err)
	}
	decoded, err := DecodeProjectBindingsTOML(data)
	if err != nil {
		t.Fatalf("decode bindings after failure: %v", err)
	}
	if !reflect.DeepEqual(decoded, previous) {
		t.Fatalf("bindings after failure = %#v, want %#v", decoded, previous)
	}

	tempFiles, err := filepath.Glob(filepath.Join(dir, ".projects.toml.tmp-*"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(tempFiles) != 0 {
		t.Fatalf("temp file count = %d, want 0: %v", len(tempFiles), tempFiles)
	}
}

func testProjectBinding(path string, contextID string, createdAt time.Time) Binding {
	return Binding{
		ProjectPath: Path(path),
		ContextID:   devcontext.MustID(contextID),
		CreatedAt:   createdAt,
	}
}
