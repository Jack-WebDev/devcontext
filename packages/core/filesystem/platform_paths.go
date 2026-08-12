package filesystem

// PlatformPaths resolves filesystem paths whose concrete values depend on the
// operating system or current user.
//
// Core packages should depend on this interface instead of reading environment
// variables, calling os.UserHomeDir directly, or branching on runtime.GOOS.
type PlatformPaths interface {
	// UserHomeDir returns the current user's home directory.
	UserHomeDir() (string, error)

	// DevContextHomeDir returns the root directory used by Dev Context for local
	// configuration, state, logs, and isolated tool data.
	DevContextHomeDir() (string, error)

	// NormalizePath converts a user-supplied path into the platform's canonical
	// path representation.
	NormalizePath(path string) (string, error)
}
