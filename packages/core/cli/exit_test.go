package cli_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"devctx/packages/core/cli"
	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

func TestExitCodeForErrorMapsStableCLIOutcomes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want cli.ExitCode
	}{
		{
			name: "success",
			err:  nil,
			want: cli.ExitSuccess,
		},
		{
			name: "usage",
			err:  fmt.Errorf("parse command: %w", cli.ErrInvalidCommand),
			want: cli.ExitUsageError,
		},
		{
			name: "unknown command",
			err:  fmt.Errorf("parse command: %w", cli.ErrUnknownCommand),
			want: cli.ExitUsageError,
		},
		{
			name: "validation",
			err:  fmt.Errorf("validate project: %w", project.ErrProjectDirectoryNotFound),
			want: cli.ExitValidationError,
		},
		{
			name: "context mismatch requires confirmation",
			err:  fmt.Errorf("resolve context: %w", launcher.ErrContextMismatchRequiresConfirmation),
			want: cli.ExitValidationError,
		},
		{
			name: "context selection required",
			err:  fmt.Errorf("build launch plan: %w", launcher.ErrLaunchSelectionRequired),
			want: cli.ExitValidationError,
		},
		{
			name: "config validation",
			err:  fmt.Errorf("load config: %w", config.ErrInvalidGlobalConfig),
			want: cli.ExitValidationError,
		},
		{
			name: "permission",
			err:  fmt.Errorf("open storage: %w", os.ErrPermission),
			want: cli.ExitValidationError,
		},
		{
			name: "canceled",
			err:  fmt.Errorf("confirm launch: %w", cli.ErrCanceled),
			want: cli.ExitCanceled,
		},
		{
			name: "context mismatch rejected",
			err:  fmt.Errorf("resolve context: %w", launcher.ErrContextMismatchRejected),
			want: cli.ExitCanceled,
		},
		{
			name: "launch failure",
			err:  fmt.Errorf("start editor: %w", cli.ErrLaunchFailed),
			want: cli.ExitLaunchFailure,
		},
		{
			name: "editor detection failure",
			err:  fmt.Errorf("detect editor: %w", codingtool.ErrExecutableNotFound),
			want: cli.ExitLaunchFailure,
		},
		{
			name: "mapped process launch failure",
			err:  fmt.Errorf("start editor: %w", launcher.ErrProcessExecutableNotFound),
			want: cli.ExitLaunchFailure,
		},
		{
			name: "internal",
			err:  errors.New("unexpected failure"),
			want: cli.ExitInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cli.ExitCodeForError(tt.err); got != tt.want {
				t.Fatalf("exit code = %d, want %d", got, tt.want)
			}
		})
	}
}
