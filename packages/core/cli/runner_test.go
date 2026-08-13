package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devctx/packages/core/cli"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
)

func TestRunnerContextListRendersEmptyAndPopulatedContexts(t *testing.T) {
	fixture := newRunnerFixture(t)

	empty := fixture.runner().Run([]string{"context", "list"})
	assertResult(t, empty, cli.ExitSuccess, "No contexts configured.\n", "")

	client := testCLIContext("client-a", "Client A")
	company := testCLIContext("company", "Company")
	personal := testCLIContext("personal", "Personal")
	fixture.writeContext(t, personal)
	fixture.writeContext(t, company)
	fixture.writeContext(t, client)

	populated := fixture.runner().Run([]string{"context", "list"})
	assertResult(t, populated, cli.ExitSuccess, ""+
		"NAME      ID\n"+
		"Client A  client-a\n"+
		"Company   company\n"+
		"Personal  personal\n", "")
}

func TestRunnerProjectShowRendersBoundUnboundAndDanglingStates(t *testing.T) {
	t.Run("unbound", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "show"})
		assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n", "")
	})

	t.Run("bound", func(t *testing.T) {
		fixture := newRunnerFixture(t)
		fixture.writeContext(t, testCLIContext("personal", "Personal"))
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		})

		result := fixture.runner().Run([]string{"project", "show"})
		assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n", "")
	})

	t.Run("dangling", func(t *testing.T) {
		fixture := newRunnerFixture(t)
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("company"),
			CreatedAt:   fixture.now,
		})

		result := fixture.runner().Run([]string{"project", "show"})
		if result.Code != cli.ExitSuccess {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
		}
		for _, want := range []string{
			"Project:\n" + fixture.workingDir,
			"Context:\nmissing: company",
			"Binding:\ndangling",
			project.DanglingProjectBindingRecovery,
		} {
			if !strings.Contains(result.Stdout, want) {
				t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
			}
		}
	})
}

func TestRunnerProjectBindPersistsBindingAndShowReportsIt(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	result := fixture.runner().Run([]string{"project", "bind", "personal"})
	if result.Code != cli.ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
	}
	for _, want := range []string{
		"Project:\n" + fixture.workingDir,
		"Context:\npersonal",
		"Status:\nbound",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
		}
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("binding count = %d, want 1", len(stored))
	}
	if stored[0].ProjectPath != project.Path(fixture.workingDir) {
		t.Fatalf("project path = %q, want %q", stored[0].ProjectPath, fixture.workingDir)
	}
	if stored[0].ContextID != devcontext.MustID("personal") {
		t.Fatalf("context ID = %q, want personal", stored[0].ContextID.String())
	}
	if !stored[0].CreatedAt.Equal(fixture.now) {
		t.Fatalf("created at = %s, want %s", stored[0].CreatedAt, fixture.now)
	}

	restartedRunner := fixture.runner()
	show := restartedRunner.Run([]string{"project", "show"})
	assertResult(t, show, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n", "")
}

func TestRunnerProjectBindReturnsValidationErrors(t *testing.T) {
	t.Run("invalid context ID", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "bind", "Personal"})
		if result.Code != cli.ExitValidationError {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitValidationError, result.Stderr)
		}
		if !strings.Contains(result.Stderr, "Unable to use context ID") {
			t.Fatalf("stderr = %q, want context ID guidance", result.Stderr)
		}
	})

	t.Run("missing context", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "bind", "personal"})
		if result.Code != cli.ExitValidationError {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitValidationError, result.Stderr)
		}
		if !strings.Contains(result.Stderr, "Unable to find context") {
			t.Fatalf("stderr = %q, want missing context guidance", result.Stderr)
		}
	})
}

func TestRunnerProjectUnbindRemovesOnlyCurrentProjectBinding(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))
	otherProject := fixture.mkdir(t, "projects", "other")
	fixture.writeBindings(t,
		project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
		project.Binding{
			ProjectPath: project.Path(otherProject),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
	)

	result := fixture.runner().Run([]string{"project", "unbind"})
	if result.Code != cli.ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
	}
	for _, want := range []string{
		"Project:\n" + fixture.workingDir,
		"Removed context:\npersonal",
		"Status:\nunbound",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
		}
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("binding count = %d, want 1", len(stored))
	}
	if stored[0].ProjectPath != project.Path(otherProject) {
		t.Fatalf("remaining project path = %q, want %q", stored[0].ProjectPath, otherProject)
	}

	show := fixture.runner().Run([]string{"project", "show"})
	assertResult(t, show, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n", "")
}

func TestRunnerProjectUnbindIsIdempotentForUnboundProject(t *testing.T) {
	fixture := newRunnerFixture(t)

	result := fixture.runner().Run([]string{"project", "unbind"})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n\nStatus:\nunchanged\n", "")
}

func TestRunnerReturnsUsageExitCodeForInvalidCommandShapes(t *testing.T) {
	fixture := newRunnerFixture(t)

	result := fixture.runner().Run([]string{"project", "bind"})
	if result.Code != cli.ExitUsageError {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitUsageError, result.Stderr)
	}
	if !strings.Contains(result.Stderr, "Unable to parse command") {
		t.Fatalf("stderr = %q, want usage guidance", result.Stderr)
	}
}

type runnerFixture struct {
	root         string
	homeDir      string
	contextsDir  string
	workingDir   string
	bindingsPath string
	paths        filesystem.PlatformPaths
	now          time.Time
}

func newRunnerFixture(t *testing.T) runnerFixture {
	t.Helper()

	root := t.TempDir()
	fixture := runnerFixture{
		root:         root,
		homeDir:      filepath.Join(root, "home"),
		contextsDir:  filepath.Join(root, "contexts"),
		workingDir:   filepath.Join(root, "projects", "current"),
		bindingsPath: filepath.Join(root, "projects.toml"),
		now:          time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
	fixture.paths = filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return fixture.homeDir, nil
	})

	fixture.mkdir(t, "home")
	fixture.mkdir(t, "contexts")
	fixture.mkdir(t, "projects", "current")

	return fixture
}

func (f runnerFixture) runner() cli.Runner {
	return cli.Runner{
		Contexts:         devcontext.NewRepository(f.contextsDir),
		Projects:         project.NewRepository(f.bindingsPath, f.paths),
		WorkingDirectory: f.workingDir,
		Now: func() time.Time {
			return f.now
		},
	}
}

func (f runnerFixture) mkdir(t *testing.T, elements ...string) string {
	t.Helper()

	path := filepath.Join(append([]string{f.root}, elements...)...)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory fixture %q: %v", path, err)
	}
	return path
}

func (f runnerFixture) writeContext(t *testing.T, ctx devcontext.Context) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(f.contextsDir, ctx.ID.String()), 0o700); err != nil {
		t.Fatalf("create context directory %q: %v", ctx.ID.String(), err)
	}
	if err := devcontext.NewRepository(f.contextsDir).Write(ctx); err != nil {
		t.Fatalf("write context %q: %v", ctx.ID.String(), err)
	}
}

func (f runnerFixture) writeBindings(t *testing.T, bindings ...project.Binding) {
	t.Helper()

	if err := project.WriteProjectBindingsFile(f.bindingsPath, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
}

func testCLIContext(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:        devcontext.MustID(id),
		Name:      name,
		Editor:    editor.DefaultConfig(),
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}

func assertResult(t *testing.T, got cli.Result, wantCode cli.ExitCode, wantStdout string, wantStderr string) {
	t.Helper()

	if got.Code != wantCode {
		t.Fatalf("exit code = %d, want %d", got.Code, wantCode)
	}
	if got.Stdout != wantStdout {
		t.Fatalf("stdout = %q, want %q", got.Stdout, wantStdout)
	}
	if got.Stderr != wantStderr {
		t.Fatalf("stderr = %q, want %q", got.Stderr, wantStderr)
	}
}
