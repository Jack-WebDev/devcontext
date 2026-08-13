package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
)

func TestRepositoryLookupFindsEquivalentProjectPathBindings(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	createDirectory(t, projectDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	repository := project.NewRepository(bindingsPath, platformPaths)
	binding := projectBinding(projectDir, "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{binding}); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}

	tests := []string{
		projectDir,
		projectDir + string(os.PathSeparator),
		"~/projects/app",
		".",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			lookup, err := repository.Lookup(input, project.Path(projectDir))
			if err != nil {
				t.Fatalf("lookup project binding: %v", err)
			}
			if !lookup.Bound {
				t.Fatal("lookup is unbound, want bound")
			}
			if lookup.ProjectPath != project.Path(projectDir) {
				t.Fatalf("lookup path = %q, want %q", lookup.ProjectPath, projectDir)
			}
			if !reflect.DeepEqual(lookup.Binding, binding) {
				t.Fatalf("binding = %#v, want %#v", lookup.Binding, binding)
			}
		})
	}
}

func TestRepositoryLookupReturnsExplicitUnboundResult(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "unbound")
	createDirectory(t, projectDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	repository := project.NewRepository(filepath.Join(homeDir, ".devctx", "projects.toml"), platformPaths)

	lookup, err := repository.Lookup(projectDir, project.Path(homeDir))
	if err != nil {
		t.Fatalf("lookup project binding: %v", err)
	}
	if lookup.Bound {
		t.Fatalf("lookup bound = true, want false")
	}
	if lookup.ProjectPath != project.Path(projectDir) {
		t.Fatalf("lookup path = %q, want %q", lookup.ProjectPath, projectDir)
	}
}

func TestRepositoryBindCreatesBindingAndPreservesUnrelatedEntries(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	unrelatedDir := filepath.Join(homeDir, "projects", "other")
	createDirectory(t, projectDir)
	createDirectory(t, unrelatedDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	contextsDir := filepath.Join(homeDir, ".devctx", "contexts")
	contextRepository := createContextRepository(t, contextsDir, devcontext.DefaultPersonalContext(testTime(10, 0)), devcontext.DefaultCompanyContext(testTime(10, 5)))
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	unrelated := projectBinding(unrelatedDir, "company", testTime(9, 0))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{unrelated}); err != nil {
		t.Fatalf("write initial bindings: %v", err)
	}
	repository := project.NewRepository(bindingsPath, platformPaths)

	binding, err := repository.Bind("~/projects/app", project.Path(homeDir), devcontext.MustID("personal"), contextRepository, testTime(11, 0))
	if err != nil {
		t.Fatalf("bind project: %v", err)
	}

	wantBinding := projectBinding(projectDir, "personal", testTime(11, 0))
	if !reflect.DeepEqual(binding, wantBinding) {
		t.Fatalf("binding = %#v, want %#v", binding, wantBinding)
	}

	stored, err := project.ReadProjectBindingsFile(bindingsPath)
	if err != nil {
		t.Fatalf("read stored bindings: %v", err)
	}
	want := []project.Binding{wantBinding, unrelated}
	if !reflect.DeepEqual(stored, want) {
		t.Fatalf("stored bindings = %#v, want %#v", stored, want)
	}
}

func TestRepositoryBindReplacesExistingProjectBinding(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	unrelatedDir := filepath.Join(homeDir, "projects", "other")
	createDirectory(t, projectDir)
	createDirectory(t, unrelatedDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	contextsDir := filepath.Join(homeDir, ".devctx", "contexts")
	contextRepository := createContextRepository(t, contextsDir, devcontext.DefaultPersonalContext(testTime(10, 0)), devcontext.DefaultCompanyContext(testTime(10, 5)))
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	previous := projectBinding(projectDir, "company", testTime(9, 0))
	unrelated := projectBinding(unrelatedDir, "company", testTime(9, 5))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{previous, unrelated}); err != nil {
		t.Fatalf("write initial bindings: %v", err)
	}
	repository := project.NewRepository(bindingsPath, platformPaths)

	binding, err := repository.Bind(projectDir, project.Path(homeDir), devcontext.MustID("personal"), contextRepository, testTime(11, 0))
	if err != nil {
		t.Fatalf("bind project: %v", err)
	}

	wantBinding := projectBinding(projectDir, "personal", testTime(11, 0))
	if !reflect.DeepEqual(binding, wantBinding) {
		t.Fatalf("binding = %#v, want %#v", binding, wantBinding)
	}

	stored, err := project.ReadProjectBindingsFile(bindingsPath)
	if err != nil {
		t.Fatalf("read stored bindings: %v", err)
	}
	want := []project.Binding{wantBinding, unrelated}
	if !reflect.DeepEqual(stored, want) {
		t.Fatalf("stored bindings = %#v, want %#v", stored, want)
	}
}

func TestRepositoryBindRejectsMissingTargetContext(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	createDirectory(t, projectDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	contextRepository := devcontext.NewRepository(filepath.Join(homeDir, ".devctx", "contexts"))
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	repository := project.NewRepository(bindingsPath, platformPaths)

	_, err := repository.Bind(projectDir, project.Path(homeDir), devcontext.MustID("missing"), contextRepository, testTime(11, 0))
	if !errors.Is(err, devcontext.ErrContextNotFound) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextNotFound)
	}

	stored, readErr := project.ReadProjectBindingsFile(bindingsPath)
	if readErr != nil {
		t.Fatalf("read stored bindings: %v", readErr)
	}
	if len(stored) != 0 {
		t.Fatalf("stored binding count = %d, want 0", len(stored))
	}
}

func TestRepositoryUnbindRemovesSelectedBindingOnly(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	unrelatedDir := filepath.Join(homeDir, "projects", "other")
	createDirectory(t, projectDir)
	createDirectory(t, unrelatedDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	target := projectBinding(projectDir, "personal", testTime(9, 0))
	unrelated := projectBinding(unrelatedDir, "company", testTime(9, 5))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{target, unrelated}); err != nil {
		t.Fatalf("write initial bindings: %v", err)
	}
	repository := project.NewRepository(bindingsPath, platformPaths)

	result, err := repository.Unbind("~/projects/app", project.Path(homeDir))
	if err != nil {
		t.Fatalf("unbind project: %v", err)
	}
	if !result.Removed {
		t.Fatal("removed = false, want true")
	}
	if result.ProjectPath != project.Path(projectDir) {
		t.Fatalf("result path = %q, want %q", result.ProjectPath, projectDir)
	}
	if !reflect.DeepEqual(result.Binding, target) {
		t.Fatalf("removed binding = %#v, want %#v", result.Binding, target)
	}

	stored, err := project.ReadProjectBindingsFile(bindingsPath)
	if err != nil {
		t.Fatalf("read stored bindings: %v", err)
	}
	want := []project.Binding{unrelated}
	if !reflect.DeepEqual(stored, want) {
		t.Fatalf("stored bindings = %#v, want %#v", stored, want)
	}
}

func TestRepositoryUnbindIsIdempotentForAlreadyUnboundProject(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	unrelatedDir := filepath.Join(homeDir, "projects", "other")
	createDirectory(t, projectDir)
	createDirectory(t, unrelatedDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	unrelated := projectBinding(unrelatedDir, "company", testTime(9, 5))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{unrelated}); err != nil {
		t.Fatalf("write initial bindings: %v", err)
	}
	repository := project.NewRepository(bindingsPath, platformPaths)

	result, err := repository.Unbind(projectDir, project.Path(homeDir))
	if err != nil {
		t.Fatalf("unbind project: %v", err)
	}
	if result.Removed {
		t.Fatal("removed = true, want false")
	}
	if result.ProjectPath != project.Path(projectDir) {
		t.Fatalf("result path = %q, want %q", result.ProjectPath, projectDir)
	}

	stored, err := project.ReadProjectBindingsFile(bindingsPath)
	if err != nil {
		t.Fatalf("read stored bindings: %v", err)
	}
	want := []project.Binding{unrelated}
	if !reflect.DeepEqual(stored, want) {
		t.Fatalf("stored bindings = %#v, want %#v", stored, want)
	}
}

func TestRepositoryLookupWithContextValidationReportsDanglingBinding(t *testing.T) {
	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	createDirectory(t, projectDir)
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	contextsDir := filepath.Join(homeDir, ".devctx", "contexts")
	contextRepository := createContextRepository(t, contextsDir, devcontext.DefaultPersonalContext(testTime(10, 0)), devcontext.DefaultCompanyContext(testTime(10, 5)))
	if err := os.RemoveAll(filepath.Join(contextsDir, "company")); err != nil {
		t.Fatalf("remove company context directory: %v", err)
	}
	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	createDirectory(t, filepath.Dir(bindingsPath))
	binding := projectBinding(projectDir, "company", testTime(9, 0))
	if err := project.WriteProjectBindingsFile(bindingsPath, []project.Binding{binding}); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
	repository := project.NewRepository(bindingsPath, platformPaths)

	lookup, err := repository.LookupWithContextValidation(projectDir, project.Path(homeDir), contextRepository)
	if err != nil {
		t.Fatalf("lookup with context validation: %v", err)
	}
	if lookup.Bound {
		t.Fatal("bound = true, want false")
	}
	if !lookup.Dangling {
		t.Fatal("dangling = false, want true")
	}
	if lookup.ProjectPath != project.Path(projectDir) {
		t.Fatalf("lookup path = %q, want %q", lookup.ProjectPath, projectDir)
	}
	if !reflect.DeepEqual(lookup.Binding, binding) {
		t.Fatalf("binding = %#v, want %#v", lookup.Binding, binding)
	}
	if lookup.MissingContextID != devcontext.MustID("company") {
		t.Fatalf("missing context ID = %q, want company", lookup.MissingContextID)
	}
	if !strings.Contains(lookup.Recovery, "rebind") || !strings.Contains(lookup.Recovery, "unbind") {
		t.Fatalf("recovery = %q, want rebind/unbind guidance", lookup.Recovery)
	}
}

func createContextRepository(t *testing.T, contextsDir string, contexts ...devcontext.Context) devcontext.Repository {
	t.Helper()

	repository := devcontext.NewRepository(contextsDir)
	for _, ctx := range contexts {
		contextDir := filepath.Join(contextsDir, ctx.ID.String())
		if err := os.MkdirAll(contextDir, 0o700); err != nil {
			t.Fatalf("create context directory %q: %v", contextDir, err)
		}
		if err := repository.Write(ctx); err != nil {
			t.Fatalf("write context %q: %v", ctx.ID, err)
		}
	}
	return repository
}

func createDirectory(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory %q: %v", path, err)
	}
}

func testTime(hour int, minute int) time.Time {
	return time.Date(2026, 8, 13, hour, minute, 0, 0, time.UTC)
}
