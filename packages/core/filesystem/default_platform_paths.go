package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ErrUserHomeUnavailable identifies failures to resolve a usable home
// directory for the current user.
var ErrUserHomeUnavailable = errors.New("user home directory cannot be determined")

// UserHomeResolver returns the current user's home directory.
type UserHomeResolver func() (string, error)

// DefaultPlatformPaths is the default PlatformPaths implementation.
type DefaultPlatformPaths struct {
	userHomeDir UserHomeResolver
}

var _ PlatformPaths = (*DefaultPlatformPaths)(nil)

// NewDefaultPlatformPaths creates PlatformPaths backed by the local operating
// system.
func NewDefaultPlatformPaths() *DefaultPlatformPaths {
	return NewDefaultPlatformPathsWithUserHome(os.UserHomeDir)
}

// NewDefaultPlatformPathsWithUserHome creates PlatformPaths with an injected
// user home resolver.
func NewDefaultPlatformPathsWithUserHome(userHomeDir UserHomeResolver) *DefaultPlatformPaths {
	if userHomeDir == nil {
		userHomeDir = os.UserHomeDir
	}
	return &DefaultPlatformPaths{userHomeDir: userHomeDir}
}

// UserHomeDir returns a cleaned, absolute home directory.
func (p *DefaultPlatformPaths) UserHomeDir() (string, error) {
	home, err := p.userHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUserHomeUnavailable, err)
	}

	home = strings.TrimSpace(home)
	if home == "" {
		return "", ErrUserHomeUnavailable
	}

	home = cleanPlatformPath(home)
	if !isAbsolutePlatformPath(home) {
		return "", fmt.Errorf("%w: %q is not absolute", ErrUserHomeUnavailable, home)
	}

	return home, nil
}

// DevContextHomeDir returns the default root for all local Dev Context state.
func (p *DefaultPlatformPaths) DevContextHomeDir() (string, error) {
	home, err := p.UserHomeDir()
	if err != nil {
		return "", err
	}

	return joinPlatformPath(home, ".devctx"), nil
}

// NormalizePath expands the current user's home marker and cleans the path.
func (p *DefaultPlatformPaths) NormalizePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	switch {
	case path == "~":
		return p.UserHomeDir()
	case strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`):
		home, err := p.UserHomeDir()
		if err != nil {
			return "", err
		}
		return joinPlatformPath(home, strings.TrimLeft(path[1:], `/\`)), nil
	default:
		return cleanPlatformPath(path), nil
	}
}

func cleanPlatformPath(path string) string {
	if usesWindowsSeparators(path) {
		return cleanWindowsPath(path)
	}
	return filepath.Clean(path)
}

func joinPlatformPath(base string, element string) string {
	if usesWindowsSeparators(base) {
		return cleanWindowsPath(strings.TrimRight(base, `\/`) + `\` + strings.TrimLeft(element, `\/`))
	}
	return filepath.Join(base, element)
}

func isAbsolutePlatformPath(path string) bool {
	return filepath.IsAbs(path) || isWindowsAbsolutePath(path)
}

func usesWindowsSeparators(path string) bool {
	return strings.Contains(path, `\`) || isWindowsDrivePath(path)
}

func isWindowsAbsolutePath(path string) bool {
	if strings.HasPrefix(path, `\\`) {
		return len(strings.Trim(path, `\`)) > 0
	}
	return len(path) >= 3 &&
		isWindowsDrivePath(path) &&
		(path[2] == '\\' || path[2] == '/')
}

func isWindowsDrivePath(path string) bool {
	return len(path) >= 2 && unicode.IsLetter(rune(path[0])) && path[1] == ':'
}

func cleanWindowsPath(path string) string {
	path = strings.ReplaceAll(path, "/", `\`)

	prefix := ""
	rest := path
	if isWindowsDrivePath(path) {
		prefix = strings.ToUpper(path[:1]) + ":"
		rest = path[2:]
	} else if strings.HasPrefix(path, `\\`) {
		prefix = `\\`
		rest = strings.TrimLeft(path[2:], `\`)
	}

	absolute := strings.HasPrefix(rest, `\`)
	parts := strings.FieldsFunc(rest, func(r rune) bool {
		return r == '\\'
	})

	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(cleaned) > 0 && cleaned[len(cleaned)-1] != ".." {
				cleaned = cleaned[:len(cleaned)-1]
				continue
			}
			if !absolute {
				cleaned = append(cleaned, part)
			}
		default:
			cleaned = append(cleaned, part)
		}
	}

	joined := strings.Join(cleaned, `\`)
	if prefix != "" {
		if absolute {
			if joined == "" {
				return prefix + `\`
			}
			return prefix + `\` + joined
		}
		if joined == "" {
			return prefix
		}
		return prefix + joined
	}

	if absolute {
		if joined == "" {
			return `\`
		}
		return `\` + joined
	}

	if joined == "" {
		return "."
	}
	return joined
}
