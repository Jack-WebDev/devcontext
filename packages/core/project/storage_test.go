package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"devctx/packages/core/filesystem"
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

func TestWriteProjectBindingsFileReportsPermissionDeniedPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix directory permissions consistently")
	}

	dir := filepath.Join(t.TempDir(), "locked")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("create locked dir: %v", err)
	}
	path := filepath.Join(dir, "projects.toml")
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatalf("chmod locked dir: %v", err)
	}
	defer os.Chmod(dir, 0o700)

	err := project.WriteProjectBindingsFile(path, []project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	})
	if err == nil {
		t.Skip("current user can still write files below mode 000 directories")
	}
	var permissionErr *filesystem.StoragePermissionError
	if !errors.As(err, &permissionErr) {
		t.Fatalf("error = %T %v, want *filesystem.StoragePermissionError", err, err)
	}
	if permissionErr.StorageOperation() != "create temporary file" {
		t.Fatalf("operation = %q, want create temporary file", permissionErr.StorageOperation())
	}
	if permissionErr.StoragePath() != dir {
		t.Fatalf("path = %q, want %q", permissionErr.StoragePath(), dir)
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
