package filesystem_test

import (
	"testing"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
)

func TestDeriveContextPathsFromUnixHome(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/alex", nil
	})

	contextPaths, err := filesystem.DeriveContextPaths(paths, devcontext.MustID("personal"))
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}

	assertContextPaths(t, contextPaths, filesystem.ContextPaths{
		ContextID:         devcontext.MustID("personal"),
		RootDir:           "/home/alex/.devctx/contexts/personal",
		ConfigPath:        "/home/alex/.devctx/contexts/personal/context.toml",
		ClaudeDir:         "/home/alex/.devctx/contexts/personal/claude",
		CodexDir:          "/home/alex/.devctx/contexts/personal/codex",
		VSCodeDir:         "/home/alex/.devctx/contexts/personal/vscode",
		VSCodeUserDataDir: "/home/alex/.devctx/contexts/personal/vscode/user-data",
	})
}

func TestDeriveContextPathsFromWindowsHome(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return `C:\Users\Alex`, nil
	})

	contextPaths, err := filesystem.DeriveContextPaths(paths, devcontext.MustID("client-a"))
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}

	assertContextPaths(t, contextPaths, filesystem.ContextPaths{
		ContextID:         devcontext.MustID("client-a"),
		RootDir:           `C:\Users\Alex\.devctx\contexts\client-a`,
		ConfigPath:        `C:\Users\Alex\.devctx\contexts\client-a\context.toml`,
		ClaudeDir:         `C:\Users\Alex\.devctx\contexts\client-a\claude`,
		CodexDir:          `C:\Users\Alex\.devctx\contexts\client-a\codex`,
		VSCodeDir:         `C:\Users\Alex\.devctx\contexts\client-a\vscode`,
		VSCodeUserDataDir: `C:\Users\Alex\.devctx\contexts\client-a\vscode\user-data`,
	})
}

func assertContextPaths(t *testing.T, got filesystem.ContextPaths, want filesystem.ContextPaths) {
	t.Helper()

	if got != want {
		t.Fatalf("context paths = %#v, want %#v", got, want)
	}
}
