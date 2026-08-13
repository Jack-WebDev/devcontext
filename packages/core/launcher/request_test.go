package launcher_test

import (
	"testing"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

func TestLaunchRequestRepresentsDirectCLIInput(t *testing.T) {
	contextID := devcontext.MustID("personal")

	request := launcher.LaunchRequest{
		ProjectPath:      project.Path("/work/personal/app"),
		RequestedContext: &contextID,
		Interactive:      false,
		Source:           launcher.InvocationSourceCLI,
	}

	if request.ProjectPath != "/work/personal/app" {
		t.Fatalf("project path = %q, want %q", request.ProjectPath, "/work/personal/app")
	}
	if request.RequestedContext == nil {
		t.Fatal("requested context is nil")
	}
	if *request.RequestedContext != contextID {
		t.Fatalf("requested context = %q, want %q", *request.RequestedContext, contextID)
	}
	if request.Interactive {
		t.Fatal("interactive = true, want false")
	}
	if request.Source != launcher.InvocationSourceCLI {
		t.Fatalf("source = %q, want %q", request.Source, launcher.InvocationSourceCLI)
	}
}

func TestLaunchRequestRepresentsMismatchConfirmationInput(t *testing.T) {
	contextID := devcontext.MustID("company")

	request := launcher.LaunchRequest{
		ProjectPath:          project.Path("/work/internal/api"),
		RequestedContext:     &contextID,
		MismatchConfirmation: launcher.ContextMismatchAccepted,
		Interactive:          false,
		Source:               launcher.InvocationSourceCLI,
	}

	if request.MismatchConfirmation != launcher.ContextMismatchAccepted {
		t.Fatalf("mismatch confirmation = %q, want %q", request.MismatchConfirmation, launcher.ContextMismatchAccepted)
	}
}

func TestLaunchRequestRepresentsInteractiveGUIInput(t *testing.T) {
	request := launcher.LaunchRequest{
		ProjectPath: project.Path("/work/client-a/api"),
		Interactive: true,
		Source:      launcher.InvocationSourceGUI,
	}

	if request.ProjectPath != "/work/client-a/api" {
		t.Fatalf("project path = %q, want %q", request.ProjectPath, "/work/client-a/api")
	}
	if request.RequestedContext != nil {
		t.Fatalf("requested context = %q, want nil", *request.RequestedContext)
	}
	if !request.Interactive {
		t.Fatal("interactive = false, want true")
	}
	if request.Source != launcher.InvocationSourceGUI {
		t.Fatalf("source = %q, want %q", request.Source, launcher.InvocationSourceGUI)
	}
}
