package codingtool

// ID identifies a coding-tool implementation.
type ID string

// Type is retained as the name used by the global configuration package. A
// context launch target always uses ID directly.
type Type = ID

const (
	// VSCodeID identifies the Visual Studio Code editor implementation.
	VSCodeID ID = "vscode"

	// TypeVSCode identifies Visual Studio Code in global configuration.
	TypeVSCode Type = VSCodeID
)

// Config contains settings owned by one coding tool.
type Config struct {
	// ExecutableOverride is optional. When empty, later launch planning can use
	// normal executable detection for the tool.
	ExecutableOverride string

	// Options contains tool-owned, non-sensitive configuration. Dev Context
	// persists these values but does not interpret them.
	Options map[string]string
}

// LaunchTarget selects a context's default coding tool and stores settings for
// each tool configured in that context. Keeping settings by tool ID allows a
// context to retain a future tool configuration without changing launch APIs.
type LaunchTarget struct {
	DefaultTool ID
	Tools       map[ID]Config
}

// DefaultConfig returns empty settings for a coding tool. The selected tool is
// owned by LaunchTarget rather than by an individual tool configuration.
func DefaultConfig() Config {
	return Config{}
}

// DefaultLaunchTarget returns the launch target for new contexts.
func DefaultLaunchTarget() LaunchTarget {
	return DefaultLaunchTargetForRegistry(BuiltInRegistry())
}

// DefaultLaunchTargetForRegistry returns a launch target that selects the
// default tool from the supplied registry.
func DefaultLaunchTargetForRegistry(registry Registry) LaunchTarget {
	defaultTool := registry.DefaultID()
	return LaunchTarget{
		DefaultTool: defaultTool,
		Tools:       map[ID]Config{defaultTool: {}},
	}
}

// ConfigFor returns settings for id. Unconfigured tools use empty settings.
func (target LaunchTarget) ConfigFor(id ID) Config {
	return target.Tools[id]
}
