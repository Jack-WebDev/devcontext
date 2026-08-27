package filesystem_test

import (
	"reflect"
	"testing"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
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
		ContextID:              devcontext.MustID("personal"),
		RootDir:                "/home/alex/.devctx/contexts/personal",
		ConfigPath:             "/home/alex/.devctx/contexts/personal/context.toml",
		ProviderStorageRootDir: "/home/alex/.devctx/contexts/personal/providers",
		ProviderStorageDirs:    map[provider.ID]string{},
		ToolStorageRootDir:     "/home/alex/.devctx/contexts/personal/tools",
		ToolStorageDirs:        map[codingtool.ID]string{},
	})
	if got := contextPaths.ProviderStorageDir(provider.CodexID); got != "/home/alex/.devctx/contexts/personal/providers/codex" {
		t.Fatalf("codex provider storage dir = %q", got)
	}
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
		ContextID:              devcontext.MustID("client-a"),
		RootDir:                `C:\Users\Alex\.devctx\contexts\client-a`,
		ConfigPath:             `C:\Users\Alex\.devctx\contexts\client-a\context.toml`,
		ProviderStorageRootDir: `C:\Users\Alex\.devctx\contexts\client-a\providers`,
		ProviderStorageDirs:    map[provider.ID]string{},
		ToolStorageRootDir:     `C:\Users\Alex\.devctx\contexts\client-a\tools`,
		ToolStorageDirs:        map[codingtool.ID]string{},
	})
	if got := contextPaths.ProviderStorageDir(provider.ClaudeID); got != `C:\Users\Alex\.devctx\contexts\client-a\providers\claude` {
		t.Fatalf("claude provider storage dir = %q", got)
	}
}

func assertContextPaths(t *testing.T, got filesystem.ContextPaths, want filesystem.ContextPaths) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("context paths = %#v, want %#v", got, want)
	}
}
