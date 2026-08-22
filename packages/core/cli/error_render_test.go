package cli_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"devctx/packages/core/cli"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestRenderErrorSnapshotsRepresentativeFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "missing project path",
			err: &project.PathError{
				Path:      "/missing/project",
				Operation: "inspect",
				Err:       project.ErrProjectDirectoryNotFound,
				Cause:     os.ErrNotExist,
			},
			want: "" +
				"Unable to open project\n" +
				"\n" +
				"The project path \"/missing/project\" does not exist.\n" +
				"\n" +
				"Next step:\n" +
				"Create the directory or pass an existing project directory.\n",
		},
		{
			name: "corrupt config",
			err: &config.GlobalConfigFileError{
				Path:     "/home/alex/.devctx/config.toml",
				Cause:    "the file is malformed or missing required settings",
				Recovery: "fix config.toml or move it aside so Dev Context can create a new default configuration",
				Err:      config.ErrInvalidGlobalConfig,
			},
			want: "" +
				"Unable to read Dev Context configuration\n" +
				"\n" +
				"The global configuration file at \"/home/alex/.devctx/config.toml\" is invalid: the file is malformed or missing required settings.\n" +
				"\n" +
				"Next step:\n" +
				"fix config.toml or move it aside so Dev Context can create a new default configuration.\n",
		},
		{
			name: "missing context",
			err: &devcontext.MissingContextError{
				ContextID: devcontext.MustID("client-a"),
				AvailableIDs: []devcontext.ID{
					devcontext.MustID("company"),
					devcontext.MustID("personal"),
				},
			},
			want: "" +
				"Unable to find context\n" +
				"\n" +
				"Context \"client-a\" is not configured on this machine.\n" +
				"\n" +
				"Next step:\n" +
				"Retry with one of these context IDs: company, personal.\n",
		},
		{
			name: "permission denied",
			err: &filesystem.StoragePermissionError{
				Operation: "create directory",
				Path:      "/home/alex/.devctx",
				Err:       os.ErrPermission,
			},
			want: "" +
				"Unable to access local storage\n" +
				"\n" +
				"The operating system denied permission to create directory at \"/home/alex/.devctx\".\n" +
				"\n" +
				"Next step:\n" +
				"Check ownership and permissions for the affected path, then retry.\n",
		},
		{
			name: "context storage incomplete",
			err: &filesystem.ContextStorageError{
				ContextID: devcontext.MustID("personal"),
				Missing: []filesystem.MissingContextDirectory{
					{
						Kind:       filesystem.ContextDirectoryProvider,
						ProviderID: string(provider.CodexID),
						Path:       "/home/alex/.devctx/contexts/personal/providers/codex",
						Reason:     "missing",
					},
				},
			},
			want: "" +
				"Context storage is incomplete\n" +
				"\n" +
				"Context \"personal\" is missing required storage directories: provider:codex \"/home/alex/.devctx/contexts/personal/providers/codex\" (missing).\n" +
				"\n" +
				"Next step:\n" +
				"Repair or recreate the context before launching. Dev Context will not recreate incomplete context storage automatically.\n",
		},
		{
			name: "context mismatch requires confirmation",
			err: &launcher.ContextMismatchError{
				ProjectPath:        "/work/constructa",
				BoundContextID:     devcontext.MustID("personal"),
				RequestedContextID: devcontext.MustID("company"),
				Err:                launcher.ErrContextMismatchRequiresConfirmation,
			},
			want: "" +
				"Context mismatch requires confirmation\n" +
				"\n" +
				"The project at \"/work/constructa\" is bound to context \"personal\", but the request selected context \"company\".\n" +
				"\n" +
				"Next step:\n" +
				"Confirm the mismatch intentionally or rerun with the bound context.\n",
		},
		{
			name: "selection required",
			err:  launcher.ErrLaunchSelectionRequired,
			want: "" +
				"Context selection required\n" +
				"\n" +
				"No explicit context or trusted project binding selected a context.\n" +
				"\n" +
				"Next step:\n" +
				"Choose a context in the selector or rerun with `--context <id>`.\n",
		},
		{
			name: "editor executable missing",
			err: &editor.ExecutableNotFoundError{
				EditorID:   editor.VSCodeID,
				Candidates: []string{"code"},
			},
			want: "" +
				"VS Code command not found\n" +
				"\n" +
				"Dev Context expected the VS Code CLI command on PATH and checked: `code`.\n" +
				"\n" +
				"Next step:\n" +
				"Install the VS Code command line launcher so `code` is on PATH, or set editor.executable_override in the context to a valid VS Code CLI path.\n",
		},
		{
			name: "process executable missing",
			err: &launcher.ProcessLaunchError{
				Executable: "/missing/code",
				Err:        launcher.ErrProcessExecutableNotFound,
				Cause:      os.ErrNotExist,
			},
			want: "" +
				"VS Code command not found\n" +
				"\n" +
				"Dev Context could not find the configured VS Code command \"/missing/code\".\n" +
				"\n" +
				"Next step:\n" +
				"Install the VS Code command line launcher, add it to PATH, or set editor.executable_override in the context to a valid VS Code CLI path.\n",
		},
		{
			name: "process permission denied",
			err: &launcher.ProcessLaunchError{
				Executable: "/opt/code",
				Err:        launcher.ErrProcessPermissionDenied,
				Cause:      os.ErrPermission,
			},
			want: "" +
				"Unable to launch editor\n" +
				"\n" +
				"The operating system denied permission to start the editor process.\n" +
				"\n" +
				"Next step:\n" +
				"Check executable, project, and Dev Context storage permissions, then retry.\n",
		},
		{
			name: "process working directory invalid",
			err: &launcher.ProcessLaunchError{
				Executable:       "/usr/local/bin/code",
				WorkingDirectory: "/missing/project",
				Err:              launcher.ErrProcessWorkingDirectoryInvalid,
				Cause:            os.ErrNotExist,
			},
			want: "" +
				"Unable to launch editor\n" +
				"\n" +
				"The editor working directory is missing or is not a directory.\n" +
				"\n" +
				"Next step:\n" +
				"Check the project path and run Dev Context from an existing project directory.\n",
		},
		{
			name: "process start failed",
			err: &launcher.ProcessLaunchError{
				Executable: "/usr/local/bin/code",
				Err:        launcher.ErrProcessStartFailed,
				Cause:      errors.New("exit status 1"),
			},
			want: "" +
				"Unable to launch editor\n" +
				"\n" +
				"The operating system could not start the editor process.\n" +
				"\n" +
				"Next step:\n" +
				"Check the editor command and project path, then retry.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cli.RenderError(tt.err, false); got != tt.want {
				t.Fatalf("rendered error = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderErrorDebugRedactsSensitiveValues(t *testing.T) {
	err := errors.New("launch failed with CODEX_TOKEN=secret-token and API_KEY=secret-key")

	got := cli.RenderError(err, true)
	if strings.Contains(got, "secret-token") || strings.Contains(got, "secret-key") {
		t.Fatalf("rendered debug error leaked secret value: %q", got)
	}
	for _, want := range []string{"CODEX_TOKEN=<redacted>", "API_KEY=<redacted>"} {
		if !strings.Contains(got, want) {
			t.Fatalf("rendered debug error = %q, want containing %q", got, want)
		}
	}
}
