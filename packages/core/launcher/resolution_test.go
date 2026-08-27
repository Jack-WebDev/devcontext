package launcher_test

import (
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
)

func TestResolutionResultRepresentsExplicitSource(t *testing.T) {
	context := testContext("personal", "Personal")

	result := launcher.ResolutionResult{
		Context: &context,
		Source:  launcher.ResolutionSourceExplicit,
	}

	assertResolutionResult(t, result, context, launcher.ResolutionSourceExplicit, false)
}

func TestResolutionResultRepresentsProjectBindingSource(t *testing.T) {
	context := testContext("company", "Company")

	result := launcher.ResolutionResult{
		Context: &context,
		Source:  launcher.ResolutionSourceProjectBinding,
	}

	assertResolutionResult(t, result, context, launcher.ResolutionSourceProjectBinding, false)
}

func TestResolutionResultRepresentsUserSelectionSource(t *testing.T) {
	context := testContext("client-a", "Client A")

	result := launcher.ResolutionResult{
		Context: &context,
		Source:  launcher.ResolutionSourceUserSelection,
	}

	assertResolutionResult(t, result, context, launcher.ResolutionSourceUserSelection, false)
}

func TestResolutionResultCanCarryWarnings(t *testing.T) {
	context := testContext("personal", "Personal")

	result := launcher.ResolutionResult{
		Context: &context,
		Source:  launcher.ResolutionSourceExplicit,
		Warnings: []launcher.ResolutionWarning{
			{
				Code:    "example_warning",
				Message: "example warning message",
			},
		},
	}

	if len(result.Warnings) != 1 {
		t.Fatalf("warning count = %d, want 1", len(result.Warnings))
	}
	if result.Warnings[0].Code != "example_warning" {
		t.Fatalf("warning code = %q, want %q", result.Warnings[0].Code, "example_warning")
	}
	if result.Warnings[0].Message == "" {
		t.Fatal("warning message is empty")
	}
}

func TestResolutionResultCanRequireUserSelection(t *testing.T) {
	result := launcher.ResolutionResult{
		SelectionRequired: true,
	}

	if result.Context != nil {
		t.Fatalf("context = %+v, want nil", result.Context)
	}
	if !result.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
}

func assertResolutionResult(
	t *testing.T,
	result launcher.ResolutionResult,
	wantContext devcontext.Context,
	wantSource launcher.ResolutionSource,
	wantSelectionRequired bool,
) {
	t.Helper()

	if result.Context == nil {
		t.Fatal("context is nil")
	}
	if result.Context.ID != wantContext.ID {
		t.Fatalf("context ID = %q, want %q", result.Context.ID, wantContext.ID)
	}
	if result.Context.Name != wantContext.Name {
		t.Fatalf("context name = %q, want %q", result.Context.Name, wantContext.Name)
	}
	if result.Source != wantSource {
		t.Fatalf("source = %q, want %q", result.Source, wantSource)
	}
	if result.SelectionRequired != wantSelectionRequired {
		t.Fatalf("selection required = %t, want %t", result.SelectionRequired, wantSelectionRequired)
	}
}

func testContext(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:        devcontext.MustID(id),
		Name:      name,
		Tool:      codingtool.DefaultConfig(),
		CreatedAt: time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC),
	}
}
