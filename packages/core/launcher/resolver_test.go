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

func TestResolverEnforcesContextResolutionPrecedence(t *testing.T) {
	tests := []struct {
		name               string
		requestedContextID string
		boundContextID     string
		confirmation       launcher.ContextMismatchConfirmation
		wantContextID      string
		wantSource         launcher.ResolutionSource
		wantSelection      bool
		wantWarningCodes   []launcher.WarningCode
	}{
		{
			name:          "no explicit context and no binding requires selection",
			wantSource:    launcher.ResolutionSourceUserSelection,
			wantSelection: true,
		},
		{
			name:           "no explicit context uses project binding",
			boundContextID: "personal",
			wantContextID:  "personal",
			wantSource:     launcher.ResolutionSourceProjectBinding,
		},
		{
			name:             "no explicit context requires selection for dangling binding",
			boundContextID:   "client-a",
			wantSource:       launcher.ResolutionSourceUserSelection,
			wantSelection:    true,
			wantWarningCodes: []launcher.WarningCode{launcher.WarningDanglingProjectBinding},
		},
		{
			name:               "explicit context wins without binding",
			requestedContextID: "company",
			wantContextID:      "company",
			wantSource:         launcher.ResolutionSourceExplicit,
		},
		{
			name:               "explicit context wins over matching binding without warning",
			requestedContextID: "company",
			boundContextID:     "company",
			wantContextID:      "company",
			wantSource:         launcher.ResolutionSourceExplicit,
		},
		{
			name:               "explicit context wins over conflicting binding with mismatch warning",
			requestedContextID: "company",
			boundContextID:     "personal",
			confirmation:       launcher.ContextMismatchAccepted,
			wantContextID:      "company",
			wantSource:         launcher.ResolutionSourceExplicit,
			wantWarningCodes:   []launcher.WarningCode{launcher.WarningContextMismatch},
		},
		{
			name:               "explicit context wins over dangling binding with dangling warning",
			requestedContextID: "company",
			boundContextID:     "client-a",
			wantContextID:      "company",
			wantSource:         launcher.ResolutionSourceExplicit,
			wantWarningCodes:   []launcher.WarningCode{launcher.WarningDanglingProjectBinding},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newResolverFixture(t)
			company := testContext("company", "Company")
			personal := testContext("personal", "Personal")
			fixture.writeContexts(t, personal, company)
			if tt.boundContextID != "" {
				fixture.writeBindings(t, project.Binding{
					ProjectPath: project.Path(fixture.projectDir),
					ContextID:   devcontext.MustID(tt.boundContextID),
					CreatedAt:   testResolverTime(10, 0),
				})
			}

			request := launcher.LaunchRequest{
				ProjectPath:          project.Path(fixture.projectDir),
				MismatchConfirmation: tt.confirmation,
				Interactive:          tt.requestedContextID == "",
				Source:               launcher.InvocationSourceCLI,
			}
			if tt.requestedContextID != "" {
				contextID := devcontext.MustID(tt.requestedContextID)
				request.RequestedContext = &contextID
			}

			result, err := fixture.resolver.Resolve(request)
			if err != nil {
				t.Fatalf("resolve context: %v", err)
			}

			if result.Source != tt.wantSource {
				t.Fatalf("source = %q, want %q", result.Source, tt.wantSource)
			}
			if result.SelectionRequired != tt.wantSelection {
				t.Fatalf("selection required = %t, want %t", result.SelectionRequired, tt.wantSelection)
			}
			if tt.wantSelection {
				assertSelectionRequired(t, result, []devcontext.Context{company, personal})
			} else {
				if result.Context == nil {
					t.Fatal("context is nil")
				}
				if result.Context.ID != devcontext.MustID(tt.wantContextID) {
					t.Fatalf("context ID = %q, want %q", result.Context.ID, tt.wantContextID)
				}
			}
			assertResolutionWarnings(t, result.Warnings, tt.wantWarningCodes, request, tt.boundContextID)
		})
	}
}

func TestResolverRequiresIntentionalMismatchOverride(t *testing.T) {
	tests := []struct {
		name         string
		confirmation launcher.ContextMismatchConfirmation
		interactive  bool
		wantErr      error
	}{
		{
			name:         "accepted confirmation preserves explicit context",
			confirmation: launcher.ContextMismatchAccepted,
			interactive:  false,
		},
		{
			name:         "rejected confirmation cancels resolution",
			confirmation: launcher.ContextMismatchRejected,
			interactive:  true,
			wantErr:      launcher.ErrContextMismatchRejected,
		},
		{
			name:        "non-interactive request without confirmation is rejected",
			interactive: false,
			wantErr:     launcher.ErrContextMismatchRequiresConfirmation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fixture := newResolverFixture(t)
			company := testContext("company", "Company")
			personal := testContext("personal", "Personal")
			fixture.writeContexts(t, personal, company)
			fixture.writeBindings(t, project.Binding{
				ProjectPath: project.Path(fixture.projectDir),
				ContextID:   personal.ID,
				CreatedAt:   testResolverTime(10, 0),
			})
			requestedContext := company.ID

			result, err := fixture.resolver.Resolve(launcher.LaunchRequest{
				ProjectPath:          project.Path(fixture.projectDir),
				RequestedContext:     &requestedContext,
				MismatchConfirmation: tt.confirmation,
				Interactive:          tt.interactive,
				Source:               launcher.InvocationSourceCLI,
			})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("error = %v, want %v", err, tt.wantErr)
				}
				assertContextMismatchError(t, err, project.Path(fixture.projectDir), personal.ID, company.ID)
				if result.Context != nil {
					t.Fatalf("context = %#v, want nil", *result.Context)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve context: %v", err)
			}

			assertResolvedContext(t, result, company, launcher.ResolutionSourceExplicit)
			assertResolutionWarnings(
				t,
				result.Warnings,
				[]launcher.WarningCode{launcher.WarningContextMismatch},
				launcher.LaunchRequest{
					ProjectPath:      project.Path(fixture.projectDir),
					RequestedContext: &requestedContext,
				},
				personal.ID.String(),
			)
		})
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

func assertResolutionWarnings(
	t *testing.T,
	warnings []launcher.ResolutionWarning,
	wantCodes []launcher.WarningCode,
	request launcher.LaunchRequest,
	boundContextID string,
) {
	t.Helper()

	if len(warnings) != len(wantCodes) {
		t.Fatalf("warning count = %d, want %d: %#v", len(warnings), len(wantCodes), warnings)
	}
	for index, wantCode := range wantCodes {
		warning := warnings[index]
		if warning.Code != wantCode {
			t.Fatalf("warning[%d] code = %q, want %q", index, warning.Code, wantCode)
		}
		if warning.ProjectPath != request.ProjectPath {
			t.Fatalf("warning[%d] project path = %q, want %q", index, warning.ProjectPath, request.ProjectPath)
		}
		if warning.BoundContextID != devcontext.MustID(boundContextID) {
			t.Fatalf("warning[%d] bound context = %q, want %q", index, warning.BoundContextID, boundContextID)
		}
		if wantCode == launcher.WarningContextMismatch {
			if request.RequestedContext == nil {
				t.Fatal("mismatch warning without requested context")
			}
			if warning.RequestedContextID != *request.RequestedContext {
				t.Fatalf("warning[%d] requested context = %q, want %q", index, warning.RequestedContextID, *request.RequestedContext)
			}
			if !strings.Contains(warning.Message, boundContextID) || !strings.Contains(warning.Message, request.RequestedContext.String()) {
				t.Fatalf("warning[%d] message = %q, want bound and requested contexts", index, warning.Message)
			}
		}
	}
}

func assertContextMismatchError(
	t *testing.T,
	err error,
	wantProjectPath project.Path,
	wantBoundContext devcontext.ID,
	wantRequestedContext devcontext.ID,
) {
	t.Helper()

	var mismatchErr *launcher.ContextMismatchError
	if !errors.As(err, &mismatchErr) {
		t.Fatalf("error = %v, want ContextMismatchError", err)
	}
	if mismatchErr.ProjectPath != wantProjectPath {
		t.Fatalf("mismatch project path = %q, want %q", mismatchErr.ProjectPath, wantProjectPath)
	}
	if mismatchErr.BoundContextID != wantBoundContext {
		t.Fatalf("mismatch bound context = %q, want %q", mismatchErr.BoundContextID, wantBoundContext)
	}
	if mismatchErr.RequestedContextID != wantRequestedContext {
		t.Fatalf("mismatch requested context = %q, want %q", mismatchErr.RequestedContextID, wantRequestedContext)
	}
}

func testResolverTime(hour int, minute int) time.Time {
	return time.Date(2026, 8, 13, hour, minute, 0, 0, time.UTC)
}
