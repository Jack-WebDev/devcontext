package project_test

import (
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
)

func TestBindingAssociatesProjectWithArbitraryContext(t *testing.T) {
	createdAt := time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)
	contextID := devcontext.MustID("client-a")

	binding := project.Binding{
		ProjectPath: "/work/client-a/api",
		ContextID:   contextID,
		CreatedAt:   createdAt,
	}

	if binding.ProjectPath != "/work/client-a/api" {
		t.Fatalf("project path = %q, want %q", binding.ProjectPath, "/work/client-a/api")
	}
	if binding.ContextID != contextID {
		t.Fatalf("context ID = %q, want %q", binding.ContextID, contextID)
	}
	if !binding.CreatedAt.Equal(createdAt) {
		t.Fatalf("created at = %s, want %s", binding.CreatedAt, createdAt)
	}
}
