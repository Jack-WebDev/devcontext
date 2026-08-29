package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"devctx/packages/core/cli"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/wailsapp"
)

func TestShouldRunCLIRoutesManagementAndDirectLaunchCommands(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "no arguments opens app",
			args: nil,
			want: false,
		},
		{
			name: "plain project path opens app",
			args: []string{"."},
			want: false,
		},
		{
			name: "context command",
			args: []string{"context", "list"},
			want: true,
		},
		{
			name: "project command",
			args: []string{"project", "show"},
			want: true,
		},
		{
			name: "version command",
			args: []string{"--version"},
			want: true,
		},
		{
			name: "generic direct launch",
			args: []string{"--context", "personal", "."},
			want: true,
		},
		{
			name: "personal alias direct launch",
			args: []string{"--personal", "."},
			want: true,
		},
		{
			name: "company alias direct launch",
			args: []string{"--company", "."},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldRunCLI(tt.args); got != tt.want {
				t.Fatalf("shouldRunCLI(%#v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestDesktopApplicationModeResolvesProjectInvocations(t *testing.T) {
	root := t.TempDir()
	workingDirectory := mkdir(t, root, "work", "web")
	explicitProject := mkdir(t, root, "work", "api")
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return mkdir(t, root, "home"), nil
	})

	tests := []struct {
		name string
		args []string
		want wailsapp.ApplicationMode
	}{
		{
			name: "no arguments uses management mode",
			want: wailsapp.ManagementMode(),
		},
		{
			name: "current directory",
			args: []string{"."},
			want: wailsapp.LauncherMode(workingDirectory),
		},
		{
			name: "relative project path",
			args: []string{"../api/"},
			want: wailsapp.LauncherMode(explicitProject),
		},
		{
			name: "absolute project path",
			args: []string{explicitProject},
			want: wailsapp.LauncherMode(explicitProject),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := desktopApplicationMode(tt.args, workingDirectory, paths)
			if err != nil {
				t.Fatalf("resolve desktop mode: %v", err)
			}
			if got != tt.want {
				t.Fatalf("mode = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDesktopApplicationModeReturnsTypedInvalidProjectPathErrors(t *testing.T) {
	root := t.TempDir()
	workingDirectory := mkdir(t, root, "work", "web")
	paths := filesystem.NewDefaultPlatformPaths()
	filePath := filepath.Join(root, "README.md")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write file fixture: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "missing directory",
			args: []string{filepath.Join(root, "missing")},
			want: project.ErrProjectDirectoryNotFound,
		},
		{
			name: "file instead of directory",
			args: []string{filePath},
			want: project.ErrProjectPathNotDirectory,
		},
		{
			name: "multiple project paths",
			args: []string{workingDirectory, root},
			want: cli.ErrInvalidCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := desktopApplicationMode(tt.args, workingDirectory, paths)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func mkdir(t *testing.T, root string, elements ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{root}, elements...)...)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory fixture %q: %v", path, err)
	}
	return path
}

func TestRunCLIVersionDoesNotInitializeStorage(t *testing.T) {
	root := t.TempDir()
	homeFile := filepath.Join(root, "home-is-a-file")
	if err := os.WriteFile(homeFile, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write home fixture: %v", err)
	}
	t.Setenv("HOME", homeFile)
	t.Setenv("USERPROFILE", homeFile)

	if got := runCLI([]string{"--version"}); got != 0 {
		t.Fatalf("exit code = %d, want 0", got)
	}
	if _, err := os.Stat(filepath.Join(homeFile, ".devctx")); err == nil {
		t.Fatalf("version command initialized storage, stat error = %v", err)
	}
}
