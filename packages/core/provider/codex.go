package provider

const (
	// CodexID identifies the Codex provider configuration.
	CodexID ID = "codex"

	// CodexHomeEnvVar is the environment variable Codex uses for its home.
	CodexHomeEnvVar = "CODEX_HOME"
)

// CodexProvider contributes isolated Codex process configuration.
type CodexProvider struct{}

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
		CodexHomeEnvVar: ctx.Paths.CodexDir,
	}, nil
}

// Status returns local provider readiness.
func (CodexProvider) Status(RuntimeContext) (Status, error) {
	return UnavailableStatus("Codex local status detection is not implemented"), nil
}
