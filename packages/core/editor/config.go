package editor

// Type identifies an editor implementation.
type Type string

const (
	// TypeVSCode identifies Visual Studio Code editor intent.
	TypeVSCode Type = "vscode"
)

// Config describes editor intent for a context.
type Config struct {
	Type Type

	// ExecutableOverride is optional. When empty, later launch planning can use
	// normal editor detection for the configured type.
	ExecutableOverride string
}

// DefaultConfig returns the default editor intent for new contexts.
func DefaultConfig() Config {
	return Config{Type: TypeVSCode}
}
