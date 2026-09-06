package config

import codingtool "devctx/packages/core/codingtool"

const (
	// CurrentSchemaVersion is the supported global configuration schema.
	CurrentSchemaVersion SchemaVersion = 1
)

// DefaultGlobalConfig returns the safe configuration for a new installation.
func DefaultGlobalConfig() GlobalConfig {
	defaultTool := codingtool.BuiltInRegistry().DefaultID()

	return GlobalConfig{
		Version: CurrentSchemaVersion,
		// The registry owns which built-in tool is selected by default. Global
		// configuration persists that ID without making a tool-specific choice.
		DefaultTool: defaultTool,
		UI: UISettings{
			RememberWindowPosition: true,
			CloseAfterLaunch:       true,
			LaunchVerification:     true,
			RememberProjects:       true,
			TrayEnabled:            false,
		},
		Safety: SafetySettings{
			WarnOnContextMismatch:  true,
			ConfirmUnboundProjects: true,
		},
	}
}
