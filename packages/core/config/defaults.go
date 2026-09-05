package config

import codingtool "devctx/packages/core/codingtool"

const (
	// CurrentSchemaVersion is the supported global configuration schema.
	CurrentSchemaVersion SchemaVersion = 1
)

// DefaultGlobalConfig returns the safe configuration for a new installation.
func DefaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Version:     CurrentSchemaVersion,
		DefaultTool: codingtool.TypeVSCode,
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
