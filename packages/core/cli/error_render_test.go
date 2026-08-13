package cli_test

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"devctx/packages/core/cli"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

func TestRenderErrorSnapshotsRepresentativeFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "missing project path",
			err:  fmt.Errorf("validate project: %w: /missing/project", project.ErrProjectDirectoryNotFound),
			want: "" +
				"Unable to open project\n" +
				"\n" +
				"The project path does not exist.\n" +
				"\n" +
				"Next step:\n" +
				"Check the path and run the command from an existing project directory.\n",
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
			err:  fmt.Errorf("validate target context %q: %w", "personal", devcontext.ErrContextNotFound),
			want: "" +
				"Unable to find context\n" +
				"\n" +
				"The requested context is not configured on this machine.\n" +
				"\n" +
				"Next step:\n" +
				"Run `devctx context list` to see available contexts, then retry with one of those IDs.\n",
		},
		{
			name: "permission denied",
			err:  fmt.Errorf("create Dev Context directory %q: %w", "/home/alex/.devctx", os.ErrPermission),
			want: "" +
				"Unable to access local storage\n" +
				"\n" +
				"The operating system denied permission to a required file or directory.\n" +
				"\n" +
				"Next step:\n" +
				"Check ownership and permissions for the project and ~/.devctx paths, then retry.\n",
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
