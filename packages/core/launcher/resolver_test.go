package launcher_test

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
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

func TestResolverResolvesExistingExplicitContext(t *testing.T) {
	fixture := newResolverFixture(t)
	requestedContext := devcontext.MustID("personal")
	personal := testContext("personal", "Personal")
	fixture.writeContexts(t, personal, testContext("company", "Company"))

	result, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath:      project.Path(fixture.projectDir),
		RequestedContext: &requestedContext,
		Interactive:      false,
		Source:           launcher.InvocationSourceCLI,
	})
	if err != nil {
		t.Fatalf("resolve context: %v", err)
	}

	assertResolvedContext(t, result, personal, launcher.ResolutionSourceExplicit)
	if result.SelectionRequired {
		t.Fatal("selection required = true, want false")
	}
}

func TestResolverReportsMissingExplicitContext(t *testing.T) {
	fixture := newResolverFixture(t)
	requestedContext := devcontext.MustID("missing")
	fixture.writeContexts(t, testContext("personal", "Personal"))

	_, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath:      project.Path(fixture.projectDir),
		RequestedContext: &requestedContext,
		Interactive:      false,
		Source:           launcher.InvocationSourceCLI,
	})
	if !errors.Is(err, devcontext.ErrContextNotFound) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextNotFound)
	}
}

func TestResolverReportsInvalidExplicitContextID(t *testing.T) {
	fixture := newResolverFixture(t)
	requestedContext := devcontext.ID{}
	fixture.writeContexts(t, testContext("personal", "Personal"))

	_, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath:      project.Path(fixture.projectDir),
		RequestedContext: &requestedContext,
		Interactive:      false,
		Source:           launcher.InvocationSourceCLI,
	})
	if !errors.Is(err, devcontext.ErrInvalidID) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidID)
	}
}

func TestResolverResolvesProjectBinding(t *testing.T) {
	fixture := newResolverFixture(t)
	company := testContext("company", "Company")
	fixture.writeContexts(t, testContext("personal", "Personal"), company)
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   company.ID,
		CreatedAt:   testResolverTime(10, 0),
	})

	result, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath: project.Path(fixture.projectDir),
		Interactive: true,
		Source:      launcher.InvocationSourceCLI,
	})
	if err != nil {
		t.Fatalf("resolve context: %v", err)
	}

	assertResolvedContext(t, result, company, launcher.ResolutionSourceProjectBinding)
	if result.SelectionRequired {
		t.Fatal("selection required = true, want false")
	}
}

func TestResolverRequiresSelectionForUnboundProject(t *testing.T) {
	fixture := newResolverFixture(t)
	client := testContext("client-a", "Client A")
	company := testContext("company", "Company")
	personal := testContext("personal", "Personal")
	fixture.writeContexts(t, personal, company, client)

	result, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath: project.Path(fixture.projectDir),
		Interactive: true,
		Source:      launcher.InvocationSourceCLI,
	})
	if err != nil {
		t.Fatalf("resolve context: %v", err)
	}

	assertSelectionRequired(t, result, []devcontext.Context{client, company, personal})
	if len(result.Warnings) != 0 {
		t.Fatalf("warning count = %d, want 0", len(result.Warnings))
	}
}

func TestResolverRequiresSelectionForDanglingProjectBinding(t *testing.T) {
	fixture := newResolverFixture(t)
	company := testContext("company", "Company")
	personal := testContext("personal", "Personal")
	fixture.writeContexts(t, personal, company)
	if err := os.RemoveAll(filepath.Join(fixture.contextsDir, "company")); err != nil {
		t.Fatalf("remove company context: %v", err)
	}
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   company.ID,
		CreatedAt:   testResolverTime(10, 0),
	})

	result, err := fixture.resolver.Resolve(launcher.LaunchRequest{
		ProjectPath: project.Path(fixture.projectDir),
		Interactive: true,
		Source:      launcher.InvocationSourceCLI,
	})
	if err != nil {
		t.Fatalf("resolve context: %v", err)
	}

	assertSelectionRequired(t, result, []devcontext.Context{personal})
	if len(result.Warnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(result.Warnings))
	}
	if result.Warnings[0].Code != launcher.WarningDanglingProjectBinding {
		t.Fatalf("warning code = %q, want %q", result.Warnings[0].Code, launcher.WarningDanglingProjectBinding)
	}
	if !strings.Contains(result.Warnings[0].Message, "company") {
		t.Fatalf("warning message = %q, want missing context ID", result.Warnings[0].Message)
	}
}

type resolverFixture struct {
	resolver     launcher.Resolver
	contexts     devcontext.Repository
	contextsDir  string
	bindingsPath string
	projectDir   string
}

func newResolverFixture(t *testing.T) resolverFixture {
	t.Helper()

	homeDir := t.TempDir()
	projectDir := filepath.Join(homeDir, "projects", "app")
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		t.Fatalf("create project directory: %v", err)
	}

	contextsDir := filepath.Join(homeDir, ".devctx", "contexts")
	if err := os.MkdirAll(contextsDir, 0o700); err != nil {
		t.Fatalf("create contexts directory: %v", err)
	}

	bindingsPath := filepath.Join(homeDir, ".devctx", "projects.toml")
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	contextRepository := devcontext.NewRepository(contextsDir)
	projectRepository := project.NewRepository(bindingsPath, platformPaths)

	return resolverFixture{
		resolver:     launcher.NewResolver(contextRepository, projectRepository),
		contexts:     contextRepository,
		contextsDir:  contextsDir,
		bindingsPath: bindingsPath,
		projectDir:   projectDir,
	}
}

func (f resolverFixture) writeContexts(t *testing.T, contexts ...devcontext.Context) {
	t.Helper()

	for _, ctx := range contexts {
		contextDir := filepath.Join(f.contextsDir, ctx.ID.String())
		if err := os.MkdirAll(contextDir, 0o700); err != nil {
			t.Fatalf("create context directory %q: %v", contextDir, err)
		}
		if err := f.contexts.Write(ctx); err != nil {
			t.Fatalf("write context %q: %v", ctx.ID, err)
		}
	}
}

func (f resolverFixture) writeBindings(t *testing.T, bindings ...project.Binding) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(f.bindingsPath), 0o700); err != nil {
		t.Fatalf("create binding directory: %v", err)
	}
	if err := project.WriteProjectBindingsFile(f.bindingsPath, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
}

func assertResolvedContext(t *testing.T, result launcher.ResolutionResult, want devcontext.Context, wantSource launcher.ResolutionSource) {
	t.Helper()

	if result.Context == nil {
		t.Fatal("context is nil")
	}
	if !reflect.DeepEqual(*result.Context, want) {
		t.Fatalf("context = %#v, want %#v", *result.Context, want)
	}
	if result.Source != wantSource {
		t.Fatalf("source = %q, want %q", result.Source, wantSource)
	}
	if len(result.AvailableContexts) != 0 {
		t.Fatalf("available context count = %d, want 0", len(result.AvailableContexts))
	}
}

func assertSelectionRequired(t *testing.T, result launcher.ResolutionResult, wantContexts []devcontext.Context) {
	t.Helper()

	if result.Context != nil {
		t.Fatalf("context = %#v, want nil", *result.Context)
	}
	if result.Source != launcher.ResolutionSourceUserSelection {
		t.Fatalf("source = %q, want %q", result.Source, launcher.ResolutionSourceUserSelection)
	}
	if !result.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
	if !reflect.DeepEqual(result.AvailableContexts, wantContexts) {
		t.Fatalf("available contexts = %#v, want %#v", result.AvailableContexts, wantContexts)
	}
}

func testResolverTime(hour int, minute int) time.Time {
	return time.Date(2026, 8, 13, hour, minute, 0, 0, time.UTC)
}
