package filesystem_test

import (
	"errors"
	"testing"

	"devctx/packages/core/filesystem"
)

type fakePlatformPaths struct {
	userHome       string
	devContextHome string
	normalized     map[string]string
	err            error
}

func (f fakePlatformPaths) UserHomeDir() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.userHome, nil
}

func (f fakePlatformPaths) DevContextHomeDir() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.devContextHome, nil
}

func (f fakePlatformPaths) NormalizePath(path string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	if normalized, ok := f.normalized[path]; ok {
		return normalized, nil
	}
	return path, nil
}

func TestPlatformPathsCanUseFakeImplementation(t *testing.T) {
	var paths filesystem.PlatformPaths = fakePlatformPaths{
		userHome:       "/fake/home",
		devContextHome: "/fake/home/.devctx",
		normalized: map[string]string{
			"~/project": "/fake/home/project",
		},
	}

	userHome, err := paths.UserHomeDir()
	if err != nil {
		t.Fatalf("user home: %v", err)
	}
	if userHome != "/fake/home" {
		t.Fatalf("user home = %q, want %q", userHome, "/fake/home")
	}

	devContextHome, err := paths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	if devContextHome != "/fake/home/.devctx" {
		t.Fatalf("dev context home = %q, want %q", devContextHome, "/fake/home/.devctx")
	}

	normalized, err := paths.NormalizePath("~/project")
	if err != nil {
		t.Fatalf("normalize path: %v", err)
	}
	if normalized != "/fake/home/project" {
		t.Fatalf("normalized path = %q, want %q", normalized, "/fake/home/project")
	}
}

func TestPlatformPathsFakeCanReturnErrors(t *testing.T) {
	expectedErr := errors.New("path unavailable")
	var paths filesystem.PlatformPaths = fakePlatformPaths{err: expectedErr}

	_, err := paths.UserHomeDir()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("user home error = %v, want %v", err, expectedErr)
	}

	_, err = paths.DevContextHomeDir()
	if !errors.Is(err, expectedErr) {
		t.Fatalf("dev context home error = %v, want %v", err, expectedErr)
	}

	_, err = paths.NormalizePath("~/project")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("normalize error = %v, want %v", err, expectedErr)
	}
}
