package cli_test

import (
	"errors"
	"fmt"
	"testing"

	"devctx/packages/core/cli"
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
			name: "canceled",
			err:  fmt.Errorf("confirm launch: %w", cli.ErrCanceled),
			want: cli.ExitCanceled,
		},
		{
			name: "launch failure",
			err:  fmt.Errorf("start editor: %w", cli.ErrLaunchFailed),
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
