package project

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"devctx/packages/core/filesystem"
)

// ErrInvalidProjectPath identifies project paths that cannot be canonicalized.
var ErrInvalidProjectPath = errors.New("invalid project path")

// SymlinkResolver resolves a path through the caller's explicit symlink policy.
type SymlinkResolver func(string) (string, error)

// CanonicalizePath converts a project path spelling into a stable absolute path.
//
// Relative input is resolved against baseDir. Symlinks are not resolved by this
// function; use CanonicalizePathWithSymlinkResolver when that policy is needed.
func CanonicalizePath(paths filesystem.PlatformPaths, input string, baseDir Path) (Path, error) {
	return CanonicalizePathWithSymlinkResolver(paths, input, baseDir, nil)
}

// CanonicalizePathWithSymlinkResolver converts a project path spelling into a
// stable absolute path and applies the supplied symlink resolver when non-nil.
func CanonicalizePathWithSymlinkResolver(paths filesystem.PlatformPaths, input string, baseDir Path, resolveSymlink SymlinkResolver) (Path, error) {
	normalizedInput, err := normalizeProjectPath(paths, input)
	if err != nil {
		return "", err
	}

	canonical := normalizedInput
	if !isAbsoluteProjectPath(canonical) {
		normalizedBase, err := normalizeProjectPath(paths, string(baseDir))
		if err != nil {
			return "", fmt.Errorf("%w: base directory: %w", ErrInvalidProjectPath, err)
		}
		if !isAbsoluteProjectPath(normalizedBase) {
			return "", fmt.Errorf("%w: base directory %q is not absolute", ErrInvalidProjectPath, normalizedBase)
		}
		canonical, err = normalizeProjectPath(paths, joinProjectPath(normalizedBase, canonical))
		if err != nil {
			return "", err
		}
	}

	if resolveSymlink != nil {
		resolved, err := resolveSymlink(canonical)
		if err != nil {
			return "", fmt.Errorf("%w: resolve symlink %q: %w", ErrInvalidProjectPath, canonical, err)
		}
		canonical, err = normalizeProjectPath(paths, resolved)
		if err != nil {
			return "", err
		}
		if !isAbsoluteProjectPath(canonical) {
			return "", fmt.Errorf("%w: resolved path %q is not absolute", ErrInvalidProjectPath, canonical)
		}
	}

	if !isAbsoluteProjectPath(canonical) {
		return "", fmt.Errorf("%w: %q is not absolute", ErrInvalidProjectPath, canonical)
	}
	return Path(canonical), nil
}

func normalizeProjectPath(paths filesystem.PlatformPaths, value string) (string, error) {
	if paths == nil {
		return "", fmt.Errorf("%w: platform paths cannot be nil", ErrInvalidProjectPath)
	}

	normalized, err := paths.NormalizePath(value)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrInvalidProjectPath, err)
	}
	normalized = trimTrailingProjectSeparators(normalized)
	if normalized == "" {
		return "", fmt.Errorf("%w: path cannot be empty", ErrInvalidProjectPath)
	}
	return normalized, nil
}

func trimTrailingProjectSeparators(path string) string {
	if path == "/" {
		return path
	}
	trimmed := strings.TrimRight(path, `/\`)
	if isProjectVolumeRoot(trimmed) {
		return trimmed + `\`
	}
	if trimmed == "" {
		return path
	}
	return trimmed
}

func joinProjectPath(base string, element string) string {
	if usesWindowsProjectPath(base) {
		return strings.TrimRight(base, `\/`) + `\` + strings.TrimLeft(element, `\/`)
	}
	return filepath.Join(base, element)
}

func isAbsoluteProjectPath(path string) bool {
	return filepath.IsAbs(path) || isWindowsAbsoluteProjectPath(path)
}

func isWindowsAbsoluteProjectPath(path string) bool {
	if strings.HasPrefix(path, `\\`) {
		return len(strings.Trim(path, `\`)) > 0
	}
	return len(path) >= 3 &&
		isWindowsDriveProjectPath(path) &&
		(path[2] == '\\' || path[2] == '/')
}

func usesWindowsProjectPath(path string) bool {
	return strings.Contains(path, `\`) || isWindowsDriveProjectPath(path)
}

func isWindowsDriveProjectPath(path string) bool {
	return len(path) >= 2 && unicode.IsLetter(rune(path[0])) && path[1] == ':'
}

func isProjectVolumeRoot(path string) bool {
	return len(path) == 2 && isWindowsDriveProjectPath(path)
}
