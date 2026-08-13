package project_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"devctx/packages/core/project"
)

func TestValidateProjectDirectoryAcceptsReadableDirectory(t *testing.T) {
	dir := t.TempDir()

	if err := project.ValidateProjectDirectory(project.Path(dir)); err != nil {
		t.Fatalf("validate project directory: %v", err)
	}
}

func TestValidateProjectDirectoryDistinguishesInvalidInputs(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "README.md")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write file fixture: %v", err)
	}

	tests := []struct {
		name string
		path project.Path
		want error
	}{
		{
			name: "empty",
			path: "",
			want: project.ErrInvalidProjectPath,
		},
		{
			name: "nonexistent",
			path: project.Path(filepath.Join(tempDir, "missing")),
			want: project.ErrProjectDirectoryNotFound,
		},
		{
			name: "file",
			path: project.Path(filePath),
			want: project.ErrProjectPathNotDirectory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := project.ValidateProjectDirectory(tt.path)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateProjectDirectoryRejectsUnreadableDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix directory permissions consistently")
	}

	dir := filepath.Join(t.TempDir(), "unreadable")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("create unreadable directory fixture: %v", err)
	}
	if err := os.Chmod(dir, 0o000); err != nil {
		t.Fatalf("chmod unreadable directory fixture: %v", err)
	}
	defer os.Chmod(dir, 0o700)

	err := project.ValidateProjectDirectory(project.Path(dir))
	if err == nil {
		t.Skip("current user can still read directories with mode 000")
	}
	if !errors.Is(err, project.ErrProjectDirectoryUnreadable) {
		t.Fatalf("error = %v, want %v", err, project.ErrProjectDirectoryUnreadable)
	}
}
