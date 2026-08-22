package provider

const (
	// CodexID identifies the Codex provider configuration.
	CodexID ID = "codex"

	// CodexCommand is the executable name used to detect local Codex presence.
	CodexCommand = "codex"

	// CodexHomeEnvVar is the environment variable Codex uses for its home.
	CodexHomeEnvVar = "CODEX_HOME"
)

// CodexProvider contributes isolated Codex process configuration.
type CodexProvider struct {
	Probe StatusProbe
}

var _ Provider = CodexProvider{}

// ID returns the persisted provider identifier.
func (CodexProvider) ID() ID {
	return CodexID
}

// DisplayName returns the user-facing provider name.
func (CodexProvider) DisplayName() string {
	return "Codex"
}

// BuildEnvironment points Codex at the selected context's isolated home.
func (CodexProvider) BuildEnvironment(ctx RuntimeContext) (EnvironmentContribution, error) {
	return EnvironmentContribution{
		CodexHomeEnvVar: ctx.Paths.StorageDir,
	}, nil
}

// Status returns local provider readiness.
func (p CodexProvider) Status(ctx RuntimeContext) (Status, error) {
	return detectLocalStatus(p.Probe, p.DisplayName(), ctx.Paths.StorageDir)
}
