package filesystem_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"devctx/packages/core/filesystem"
)

func TestStoragePermissionsApplyOwnerOnlyModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose Unix permission bits consistently")
	}

	dir := filepath.Join(t.TempDir(), "storage")
	file := filepath.Join(dir, "config.toml")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(file, []byte("config"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	permissions := filesystem.NewStoragePermissions(true, os.Chmod)
	if err := permissions.ApplyDirectory(dir); err != nil {
		t.Fatalf("apply directory permissions: %v", err)
	}
	if err := permissions.ApplyFile(file); err != nil {
		t.Fatalf("apply file permissions: %v", err)
	}

	assertMode(t, dir, filesystem.RestrictedDirectoryMode)
	assertMode(t, file, filesystem.RestrictedFileMode)
}

func TestStoragePermissionsNoopWhenUnsupported(t *testing.T) {
	called := false
	permissions := filesystem.NewStoragePermissions(false, func(string, os.FileMode) error {
		called = true
		return errors.New("chmod should not be called")
	})

	if permissions.DirectoryMode() != filesystem.RestrictedDirectoryMode {
		t.Fatalf("directory mode = %v, want %v", permissions.DirectoryMode(), filesystem.RestrictedDirectoryMode)
	}
	if permissions.FileMode() != filesystem.RestrictedFileMode {
		t.Fatalf("file mode = %v, want %v", permissions.FileMode(), filesystem.RestrictedFileMode)
	}
	if err := permissions.ApplyDirectory("/unsupported/storage"); err != nil {
		t.Fatalf("apply directory permissions: %v", err)
	}
	if err := permissions.ApplyFile("/unsupported/config.toml"); err != nil {
		t.Fatalf("apply file permissions: %v", err)
	}
	if called {
		t.Fatal("chmod was called for unsupported permissions")
	}
}

func TestStoragePermissionsReturnChmodErrorsWhenSupported(t *testing.T) {
	expectedErr := errors.New("chmod failed")
	permissions := filesystem.NewStoragePermissions(true, func(string, os.FileMode) error {
		return expectedErr
	})

	err := permissions.ApplyDirectory("/storage")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %v, want %v", path, got, want)
	}
}
