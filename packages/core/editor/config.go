package editor

// ID identifies an editor implementation.
type ID string

// Type identifies an editor implementation.
//
// Type is kept as the configuration field name because existing config models
// describe editor intent. It is the same value as the editor ID used by runtime
// implementations.
type Type = ID

const (
	// VSCodeID identifies the Visual Studio Code editor implementation.
	VSCodeID ID = "vscode"

	// TypeVSCode identifies Visual Studio Code editor intent.
	TypeVSCode Type = VSCodeID
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
