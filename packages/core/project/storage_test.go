package project_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"devctx/packages/core/project"
)

func TestWriteProjectBindingsFileAtomicallyReplacesBindings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.toml")
	bindings := []project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
	}

	if err := project.WriteProjectBindingsFile(path, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	decoded, err := project.DecodeProjectBindingsTOML(data)
	if err != nil {
		t.Fatalf("decode project bindings: %v", err)
	}

	want := []project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("decoded bindings = %#v, want %#v", decoded, want)
	}
}
