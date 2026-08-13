package application

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestNewDefaultServiceOwnsDefaultDependencies(t *testing.T) {
	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	workingDirectory := filepath.Join(root, "project")
	if err := os.MkdirAll(workingDirectory, 0o700); err != nil {
		t.Fatalf("create working directory: %v", err)
	}

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	service, err := NewDefaultService(DefaultOptions{
		Paths:             paths,
		ParentEnvironment: []string{"PATH=/usr/local/bin"},
		WorkingDirectory:  workingDirectory,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("create default service: %v", err)
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".devctx", "contexts")); err != nil {
		t.Fatalf("default service did not initialize contexts directory: %v", err)
	}
	if service.dependencies.WorkingDirectory != workingDirectory {
		t.Fatalf("working directory = %q, want %q", service.dependencies.WorkingDirectory, workingDirectory)
	}
	if service.dependencies.DetachMode != launcher.DetachModeDetached {
		t.Fatalf("detach mode = %q, want %q", service.dependencies.DetachMode, launcher.DetachModeDetached)
	}
	if !service.now().Equal(now) {
		t.Fatalf("now = %s, want %s", service.now(), now)
	}

	builder := service.launchPlanBuilder()
	if builder.PlatformPaths != paths {
		t.Fatalf("builder paths = %#v, want injected paths", builder.PlatformPaths)
	}
	if !reflect.DeepEqual(builder.ParentEnvironment, []string{"PATH=/usr/local/bin"}) {
		t.Fatalf("parent environment = %#v, want injected environment", builder.ParentEnvironment)
	}
	if service.processLauncher() == nil {
		t.Fatal("process launcher is nil")
	}
}

func TestNewServiceWithDependenciesUsesFakesWithoutWails(t *testing.T) {
	root := t.TempDir()
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return filepath.Join(root, "home"), nil
	})
	contexts := devcontext.NewRepository(filepath.Join(root, "contexts"))
	projects := project.NewRepository(filepath.Join(root, "projects.toml"), paths)
	fakeProvider := applicationFakeProvider{id: "fake"}
	fakeEditor := &applicationFakeEditor{}
	fakeLauncher := &applicationFakeProcessLauncher{}
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	service := NewServiceWithDependencies(Dependencies{
		Contexts:          contexts,
		Projects:          projects,
		Paths:             paths,
		Providers:         []provider.Provider{fakeProvider},
		Editor:            fakeEditor,
		ProcessLauncher:   fakeLauncher,
		ParentEnvironment: []string{"PATH=/fixture/bin"},
		WorkingDirectory:  filepath.Join(root, "project"),
		DetachMode:        launcher.DetachModeAttached,
		Now: func() time.Time {
			return now
		},
	})

	builder := service.launchPlanBuilder()
	if builder.Editor != fakeEditor {
		t.Fatalf("builder editor = %#v, want fake editor", builder.Editor)
	}
	if !reflect.DeepEqual(builder.Providers, []provider.Provider{fakeProvider}) {
		t.Fatalf("builder providers = %#v, want fake provider", builder.Providers)
	}
	if service.processLauncher() != fakeLauncher {
		t.Fatalf("process launcher = %#v, want fake launcher", service.processLauncher())
	}
	if service.dependencies.DetachMode != launcher.DetachModeAttached {
		t.Fatalf("detach mode = %q, want %q", service.dependencies.DetachMode, launcher.DetachModeAttached)
	}
	if !service.now().Equal(now) {
		t.Fatalf("now = %s, want %s", service.now(), now)
	}
}

func TestNewErrorReturnsPresentationSafeTypedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorCode
	}{
		{
			name: "canceled",
			err:  launcher.ErrContextMismatchRejected,
			want: ErrorCodeCanceled,
		},
		{
			name: "validation",
			err:  project.ErrProjectDirectoryNotFound,
			want: ErrorCodeValidation,
		},
		{
			name: "launch",
			err:  editor.ErrExecutableNotFound,
			want: ErrorCodeLaunch,
		},
		{
			name: "internal",
			err:  errors.New("unexpected"),
			want: ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.err)
			if got.Code != tt.want {
				t.Fatalf("code = %q, want %q", got.Code, tt.want)
			}
			if got.Message == "" || got.Recovery == "" {
				t.Fatalf("message and recovery must be presentation-safe: %#v", got)
			}
			if !errors.Is(got, tt.err) {
				t.Fatalf("wrapped error = %v, want %v", got, tt.err)
			}
		})
	}

	if got := NewError(nil); got != nil {
		t.Fatalf("nil error = %#v, want nil", got)
	}
}

type applicationFakeProvider struct {
	id provider.ID
}

func (p applicationFakeProvider) ID() provider.ID {
	return p.id
}

func (p applicationFakeProvider) DisplayName() string {
	return "Fake Provider"
}

func (p applicationFakeProvider) BuildEnvironment(provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return provider.EnvironmentContribution{"FAKE": "1"}, nil
}

func (p applicationFakeProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}

type applicationFakeEditor struct{}

func (e *applicationFakeEditor) ID() editor.ID {
	return "fake-editor"
}

func (e *applicationFakeEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return "/fixture/editor", nil
}

func (e *applicationFakeEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	return editor.Command{
		Executable: request.Executable,
		Arguments:  editor.Arguments{request.ProjectPath},
	}, nil
}

type applicationFakeProcessLauncher struct{}

func (l *applicationFakeProcessLauncher) Launch(launcher.ProcessRequest) error {
	return nil
}
