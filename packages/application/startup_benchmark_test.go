package application

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func BenchmarkInteractiveLaunchStateLoading(b *testing.B) {
	root := b.TempDir()
	homeDir := filepath.Join(root, "home")
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		b.Fatalf("create project directory: %v", err)
	}

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	devContextHomeDir, err := paths.DevContextHomeDir()
	if err != nil {
		b.Fatalf("derive devctx home: %v", err)
	}
	contexts := devcontext.NewRepository(filepath.Join(devContextHomeDir, "contexts"))
	projects := project.NewRepository(filepath.Join(devContextHomeDir, "projects.toml"), paths)
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	for _, ctx := range []devcontext.Context{
		devcontext.DefaultPersonalContext(now),
		devcontext.DefaultCompanyContext(now),
	} {
		contextPaths, err := filesystem.DeriveContextPaths(paths, ctx.ID)
		if err != nil {
			b.Fatalf("derive context paths: %v", err)
		}
		if err := filesystem.CreateContextDirectoryTreeWithPermissions(contextPaths, ctx, filesystem.NewDefaultStoragePermissions()); err != nil {
			b.Fatalf("create context tree: %v", err)
		}
	}

	service := NewServiceWithDependencies(Dependencies{
		Contexts:          contexts,
		Projects:          projects,
		Paths:             paths,
		ProviderRegistry:  provider.MustNewRegistry([]provider.Provider{applicationFakeProvider{id: provider.ClaudeID}, applicationFakeProvider{id: provider.CodexID}}),
		ParentEnvironment: []string{"PATH=/usr/local/bin"},
		WorkingDirectory:  projectDir,
		Now: func() time.Time {
			return now
		},
	})

	request := GetLaunchStateRequest{ProjectPath: projectDir}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, appErr := service.GetLaunchState(request); appErr != nil {
			b.Fatalf("get launch state: %v", appErr)
		}
	}
}
