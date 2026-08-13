package project_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
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

func TestProjectBindingsPersistenceRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.toml")
	original := []project.Binding{
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	}

	if err := project.WriteProjectBindingsFile(path, original); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}

	loaded, err := project.ReadProjectBindingsFile(path)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}

	want := []project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
	}
	if !reflect.DeepEqual(loaded, want) {
		t.Fatalf("loaded bindings = %#v, want %#v", loaded, want)
	}
}

func TestConcurrentProjectBindingWritesLeaveParseableBindings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.toml")
	values := [][]project.Binding{
		{
			projectBinding("/home/jack/projects/app", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		},
		{
			projectBinding("/home/jack/work/api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
			projectBinding("/home/jack/work/site", "personal", time.Date(2026, 8, 13, 11, 30, 0, 0, time.UTC)),
		},
	}

	var wg sync.WaitGroup
	errs := make(chan error, 40)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errs <- project.WriteProjectBindingsFile(path, values[index%len(values)])
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("write project bindings: %v", err)
		}
	}

	loaded, err := project.ReadProjectBindingsFile(path)
	if err != nil {
		t.Fatalf("read final project bindings: %v", err)
	}
	if !reflect.DeepEqual(loaded, values[0]) && !reflect.DeepEqual(loaded, values[1]) {
		t.Fatalf("bindings = %#v, want one complete written value", loaded)
	}
}
