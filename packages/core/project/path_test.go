package project_test

import (
	"errors"
	"testing"

	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
)

func TestCanonicalizePathProducesStableUnixBindingKey(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})
	baseDir := project.Path("/home/jack/projects/app")

	tests := []string{
		"~/projects/app",
		"/home/jack/projects/app",
		"/home/jack/projects/app/",
		"/home/jack/projects/../projects/app",
		".",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := project.CanonicalizePath(paths, input, baseDir)
			if err != nil {
				t.Fatalf("canonicalize path: %v", err)
			}
			if got != "/home/jack/projects/app" {
				t.Fatalf("canonical path = %q, want %q", got, "/home/jack/projects/app")
			}
		})
	}
}

func TestCanonicalizePathHandlesRelativeInputAgainstBaseDirectory(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})

	got, err := project.CanonicalizePath(paths, "../api/", "/home/jack/work/internal/web")
	if err != nil {
		t.Fatalf("canonicalize path: %v", err)
	}

	if got != "/home/jack/work/internal/api" {
		t.Fatalf("canonical path = %q, want %q", got, "/home/jack/work/internal/api")
	}
}

func TestCanonicalizePathUsesPlatformRulesForWindowsPaths(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return `C:\Users\Jack`, nil
	})

	tests := []string{
		`~\projects\app`,
		`C:\Users\Jack\projects\app`,
		`C:\Users\Jack\projects\app\`,
		`C:\Users\Jack\projects\other\..\app`,
		`.\app`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := project.CanonicalizePath(paths, input, `C:\Users\Jack\projects`)
			if err != nil {
				t.Fatalf("canonicalize path: %v", err)
			}
			if got != `C:\Users\Jack\projects\app` {
				t.Fatalf("canonical path = %q, want %q", got, `C:\Users\Jack\projects\app`)
			}
		})
	}
}

func TestCanonicalizePathRejectsRelativeBaseDirectory(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})

	_, err := project.CanonicalizePath(paths, "app", "relative/base")
	if !errors.Is(err, project.ErrInvalidProjectPath) {
		t.Fatalf("error = %v, want %v", err, project.ErrInvalidProjectPath)
	}
}

func TestCanonicalizePathRejectsMissingPlatformPaths(t *testing.T) {
	_, err := project.CanonicalizePath(nil, "/home/jack/projects/app", "/home/jack")
	if !errors.Is(err, project.ErrInvalidProjectPath) {
		t.Fatalf("error = %v, want %v", err, project.ErrInvalidProjectPath)
	}
}

func TestCanonicalizePathPreservesUnixRoot(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})

	got, err := project.CanonicalizePath(paths, "/", "/home/jack")
	if err != nil {
		t.Fatalf("canonicalize path: %v", err)
	}

	if got != "/" {
		t.Fatalf("canonical path = %q, want %q", got, "/")
	}
}

func TestCanonicalizePathDoesNotResolveSymlinksByDefault(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})

	got, err := project.CanonicalizePath(paths, "/home/jack/link", "/home/jack")
	if err != nil {
		t.Fatalf("canonicalize path: %v", err)
	}

	if got != "/home/jack/link" {
		t.Fatalf("canonical path = %q, want %q", got, "/home/jack/link")
	}
}

func TestCanonicalizePathResolvesSymlinksOnlyWhenResolverIsProvided(t *testing.T) {
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/jack", nil
	})

	got, err := project.CanonicalizePathWithSymlinkResolver(paths, "/home/jack/link", "/home/jack", func(path string) (string, error) {
		if path != "/home/jack/link" {
			t.Fatalf("resolver path = %q, want %q", path, "/home/jack/link")
		}
		return "/srv/projects/app", nil
	})
	if err != nil {
		t.Fatalf("canonicalize path with resolver: %v", err)
	}

	if got != "/srv/projects/app" {
		t.Fatalf("canonical path = %q, want %q", got, "/srv/projects/app")
	}
}
