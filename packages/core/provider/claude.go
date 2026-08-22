package provider

const (
	// ClaudeID identifies the Claude provider configuration.
	ClaudeID ID = "claude"

	// ClaudeCommand is the executable name used to detect local Claude Code
	// presence.
	ClaudeCommand = "claude"

	// ClaudeConfigDirEnvVar is the environment variable Claude Code uses for
	// its configuration directory.
	ClaudeConfigDirEnvVar = "CLAUDE_CONFIG_DIR"
)

// ClaudeProvider contributes isolated Claude Code process configuration.
type ClaudeProvider struct {
	Probe StatusProbe
}

var _ Provider = ClaudeProvider{}

// ID returns the persisted provider identifier.
func (ClaudeProvider) ID() ID {
	return ClaudeID
}

// DisplayName returns the user-facing provider name.
func (ClaudeProvider) DisplayName() string {
	return "Claude"
}

// BuildEnvironment points Claude Code at the selected context's isolated config
// directory.
func (ClaudeProvider) BuildEnvironment(ctx RuntimeContext) (EnvironmentContribution, error) {
	return EnvironmentContribution{
		ClaudeConfigDirEnvVar: ctx.Paths.StorageDir,
	}, nil
}

// Status returns local provider readiness.
func (p ClaudeProvider) Status(ctx RuntimeContext) (Status, error) {
	return detectLocalStatus(p.Probe, p.DisplayName(), ctx.Paths.StorageDir)
}
