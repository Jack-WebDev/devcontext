package config

import codingtool "devctx/packages/core/codingtool"

// SchemaVersion identifies the supported shape of global configuration.
type SchemaVersion int

// GlobalConfig stores application-wide Dev Context settings.
type GlobalConfig struct {
	Version       SchemaVersion
	DefaultTool codingtool.Type
	UI            UISettings
	Safety        SafetySettings
}

// UISettings stores application-wide user interface settings.
type UISettings struct {
	RememberWindowPosition bool
}

// SafetySettings stores application-wide safety behavior settings.
type SafetySettings struct {
	WarnOnContextMismatch  bool
	ConfirmUnboundProjects bool
}
