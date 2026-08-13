package cli_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"devctx/packages/core/cli"
)

func TestParseRootLaunchCommandFormsListedInPRD(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "default invocation",
			args: nil,
		},
		{
			name: "current directory path",
			args: []string{"."},
		},
		{
			name: "explicit project path",
			args: []string{"~/projects/constructa"},
		},
		{
			name: "personal alias shape",
			args: []string{"--personal", "."},
		},
		{
			name: "generic personal context shape",
			args: []string{"--context", "personal", "."},
		},
		{
			name: "generic company context shape",
			args: []string{"--context", "company", "."},
		},
		{
			name: "generic context without explicit path",
			args: []string{"--context", "personal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := cli.Parse(tt.args)
			if err != nil {
				t.Fatalf("parse command: %v", err)
			}
			if command.Kind != cli.CommandRootLaunch {
				t.Fatalf("kind = %q, want %q", command.Kind, cli.CommandRootLaunch)
			}
			if !reflect.DeepEqual(command.RootLaunch.Arguments, tt.args) {
				t.Fatalf("arguments = %#v, want %#v", command.RootLaunch.Arguments, tt.args)
			}
		})
	}
}

func TestParseRootLaunchArgumentsAreCopied(t *testing.T) {
	args := []string{"~/projects/constructa"}
	command, err := cli.Parse(args)
	if err != nil {
		t.Fatalf("parse command: %v", err)
	}

	args[0] = "~/projects/other"
	if got := command.RootLaunch.Arguments[0]; got != "~/projects/constructa" {
		t.Fatalf("arguments[0] = %q, want original value", got)
	}
}

func TestParseContextCommandFormsListedInPRD(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cli.ContextCommand
	}{
		{
			name: "list",
			args: []string{"context", "list"},
			want: cli.ContextCommand{Subcommand: cli.ContextList},
		},
		{
			name: "create",
			args: []string{"context", "create", "client-a"},
			want: cli.ContextCommand{Subcommand: cli.ContextCreate, ContextID: "client-a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := cli.Parse(tt.args)
			if err != nil {
				t.Fatalf("parse command: %v", err)
			}
			if command.Kind != cli.CommandContext {
				t.Fatalf("kind = %q, want %q", command.Kind, cli.CommandContext)
			}
			if !reflect.DeepEqual(command.Context, tt.want) {
				t.Fatalf("context command = %#v, want %#v", command.Context, tt.want)
			}
		})
	}
}

func TestParseProjectCommandFormsListedInPRD(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want cli.ProjectCommand
	}{
		{
			name: "show",
			args: []string{"project", "show"},
			want: cli.ProjectCommand{Subcommand: cli.ProjectShow},
		},
		{
			name: "bind personal",
			args: []string{"project", "bind", "personal"},
			want: cli.ProjectCommand{Subcommand: cli.ProjectBind, ContextID: "personal"},
		},
		{
			name: "bind company",
			args: []string{"project", "bind", "company"},
			want: cli.ProjectCommand{Subcommand: cli.ProjectBind, ContextID: "company"},
		},
		{
			name: "unbind",
			args: []string{"project", "unbind"},
			want: cli.ProjectCommand{Subcommand: cli.ProjectUnbind},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := cli.Parse(tt.args)
			if err != nil {
				t.Fatalf("parse command: %v", err)
			}
			if command.Kind != cli.CommandProject {
				t.Fatalf("kind = %q, want %q", command.Kind, cli.CommandProject)
			}
			if !reflect.DeepEqual(command.Project, tt.want) {
				t.Fatalf("project command = %#v, want %#v", command.Project, tt.want)
			}
		})
	}
}

func TestParseVersionCommand(t *testing.T) {
	command, err := cli.Parse([]string{"--version"})
	if err != nil {
		t.Fatalf("parse command: %v", err)
	}
	if command.Kind != cli.CommandVersion {
		t.Fatalf("kind = %q, want %q", command.Kind, cli.CommandVersion)
	}
}

func TestParseRejectsUnknownSubcommandsClearly(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{
			name:        "unknown context subcommand",
			args:        []string{"context", "show"},
			wantMessage: `unknown context subcommand "show"`,
		},
		{
			name:        "unknown project subcommand",
			args:        []string{"project", "delete"},
			wantMessage: `unknown project subcommand "delete"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cli.Parse(tt.args)
			if !errors.Is(err, cli.ErrUnknownCommand) {
				t.Fatalf("error = %v, want %v", err, cli.ErrUnknownCommand)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestParseRejectsInvalidCommandShapes(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantMessage string
	}{
		{
			name:        "missing context subcommand",
			args:        []string{"context"},
			wantMessage: "context command requires a subcommand",
		},
		{
			name:        "context list extra argument",
			args:        []string{"context", "list", "extra"},
			wantMessage: "context list accepts no arguments",
		},
		{
			name:        "context create missing ID",
			args:        []string{"context", "create"},
			wantMessage: "context create requires exactly one context ID",
		},
		{
			name:        "context create extra argument",
			args:        []string{"context", "create", "client-a", "extra"},
			wantMessage: "context create requires exactly one context ID",
		},
		{
			name:        "missing project subcommand",
			args:        []string{"project"},
			wantMessage: "project command requires a subcommand",
		},
		{
			name:        "project show extra argument",
			args:        []string{"project", "show", "extra"},
			wantMessage: "project show accepts no arguments",
		},
		{
			name:        "project bind missing ID",
			args:        []string{"project", "bind"},
			wantMessage: "project bind requires exactly one context ID",
		},
		{
			name:        "project bind extra argument",
			args:        []string{"project", "bind", "personal", "extra"},
			wantMessage: "project bind requires exactly one context ID",
		},
		{
			name:        "project unbind extra argument",
			args:        []string{"project", "unbind", "extra"},
			wantMessage: "project unbind accepts no arguments",
		},
		{
			name:        "version extra argument",
			args:        []string{"--version", "extra"},
			wantMessage: "--version accepts no arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cli.Parse(tt.args)
			if !errors.Is(err, cli.ErrInvalidCommand) {
				t.Fatalf("error = %v, want %v", err, cli.ErrInvalidCommand)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}
