package cli_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"devctx/packages/core/cli"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func BenchmarkDirectLaunchPreparation(b *testing.B) {
	root := b.TempDir()
	homeDir := filepath.Join(root, "home")
	workingDir := filepath.Join(root, "project")
	if err := os.MkdirAll(workingDir, 0o700); err != nil {
		b.Fatalf("create working directory: %v", err)
	}

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)
	context := devcontext.DefaultPersonalContext(now)
	contextPaths, err := filesystem.DeriveContextPaths(paths, context.ID)
	if err != nil {
		b.Fatalf("derive context paths: %v", err)
	}
	if err := filesystem.CreateContextDirectoryTreeWithPermissions(contextPaths, context, filesystem.NewDefaultStoragePermissions()); err != nil {
		b.Fatalf("create context tree: %v", err)
	}
	devContextHomeDir, err := paths.DevContextHomeDir()
	if err != nil {
		b.Fatalf("derive devctx home: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		runner := cli.Runner{
			Contexts:          devcontext.NewRepository(filepath.Join(devContextHomeDir, "contexts")),
			Projects:          project.NewRepository(filepath.Join(devContextHomeDir, "projects.toml"), paths),
			WorkingDirectory:  workingDir,
			Paths:             paths,
			ProviderRegistry:  provider.DefaultRegistry(),
			Editor:            &recordingCLIEditor{},
			ProcessLauncher:   &recordingProcessLauncher{},
			ParentEnvironment: []string{"PATH=/usr/local/bin"},
			Now: func() time.Time {
				return now
			},
		}

		result := runner.Run([]string{"--context", "personal", workingDir})
		if result.Code != cli.ExitSuccess {
			b.Fatalf("exit code = %d, stderr = %q", result.Code, result.Stderr)
		}
	}
}
