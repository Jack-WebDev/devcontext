package wailsapp

// ApplicationModeKind identifies the desktop application's top-level flow.
// The host selects the mode before Wails starts; frontend code must not infer
// it from process arguments.
type ApplicationModeKind string

const (
	// ApplicationModeManagement opens the full management application.
	ApplicationModeManagement ApplicationModeKind = "management"

	// ApplicationModeLauncher opens a focused flow for one project.
	ApplicationModeLauncher ApplicationModeKind = "launcher"
)

// ApplicationMode is the host-owned startup intent for the desktop
// application. ProjectPath is set only when Type is ApplicationModeLauncher.
type ApplicationMode struct {
	Type        ApplicationModeKind `json:"type"`
	ProjectPath string              `json:"projectPath,omitempty"`
}

// ManagementMode creates the startup mode for the full management
// application.
func ManagementMode() ApplicationMode {
	return ApplicationMode{Type: ApplicationModeManagement}
}

// LauncherMode creates the startup mode for a focused project launcher.
// Project-path parsing and validation are deliberately handled by the host in
// a later startup phase.
func LauncherMode(projectPath string) ApplicationMode {
	return ApplicationMode{
		Type:        ApplicationModeLauncher,
		ProjectPath: projectPath,
	}
}
