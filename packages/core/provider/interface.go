package provider

// Provider is the small contract implemented by local AI/tool integrations.
//
// Implementations should only describe their own environment variables and
// local readiness. They must not perform remote authentication checks or load
// runtime plugins.
type Provider interface {
	ID() ID
	DisplayName() string
	BuildEnvironment(RuntimeContext) (EnvironmentContribution, error)
	Status(RuntimeContext) (Status, error)
}

// RuntimeContext is the provider-owned view of a selected Dev Context.
//
// It intentionally stores plain values instead of depending on the context or
// filesystem packages, keeping provider implementations usable without import
// cycles.
type RuntimeContext struct {
	ContextID string
	Config    Config
	Paths     ContextPaths
}

// ContextPaths contains the context-owned storage locations a provider may use.
type ContextPaths struct {
	RootDir    string
	StorageDir string
}

// EnvironmentContribution stores environment variables owned by one provider.
type EnvironmentContribution map[string]string

// MetadataField contains one safe, non-secret provider metadata value suitable
// for display.
type MetadataField struct {
	Label string
	Value string
}

// CredentialSession contains safe metadata for a globally authenticated
// provider session.
type CredentialSession struct {
	MetadataAvailable bool
	Fields            []MetadataField
}

// Identity contains safe metadata for credentials stored in a context-owned
// provider directory.
type Identity struct {
	Fields []MetadataField
}

// GlobalCredentialContext contains host paths needed to inspect global
// provider credentials.
type GlobalCredentialContext struct {
	UserHomeDir string
}

// CredentialImportContext contains source and destination paths needed to
// import opaque provider credential files into one context.
type CredentialImportContext struct {
	UserHomeDir string
	Runtime     RuntimeContext
	Files       CredentialFileOperations
}

// CredentialFileOperations provides the small set of filesystem operations a
// provider needs to import opaque credential files. The filesystem package
// supplies the implementation and owns the storage-permission policy.
type CredentialFileOperations interface {
	FileExists(path string) (bool, error)
	CopyOpaqueFile(source string, destination string) error
}

// SetupGuidance describes safe provider setup copy and optional action text.
type SetupGuidance struct {
	Message    string
	ActionHint string
}

// GlobalCredentialDetector is implemented by providers that can identify a
// global authenticated session without exposing secrets.
type GlobalCredentialDetector interface {
	DetectGlobalCredentialSession(GlobalCredentialContext) (CredentialSession, bool, error)
}

// CredentialMetadataExtractor is implemented by providers that can extract safe
// metadata from their own credential files.
type CredentialMetadataExtractor interface {
	ExtractCredentialMetadata(path string) ([]MetadataField, bool, error)
}

// CredentialImporter is implemented by providers that can copy opaque global
// credential files into their own context storage.
type CredentialImporter interface {
	ImportCredentials(CredentialImportContext) error
}

// CredentialDiagnosticFile identifies one provider-owned credential file that
// can be inspected without reading its contents. Providers opt into this
// capability so application diagnostics never infer provider file layouts.
type CredentialDiagnosticFile struct {
	Label string
	Path  string
}

// CredentialDiagnosticsProvider is implemented by providers that can safely
// identify credential files in their isolated storage for diagnostics.
type CredentialDiagnosticsProvider interface {
	CredentialDiagnosticFiles(RuntimeContext) []CredentialDiagnosticFile
}

// ContextIdentityDetector is implemented by providers that can identify the
// account represented by their context-owned credential state.
type ContextIdentityDetector interface {
	DetectContextIdentity(RuntimeContext) (Identity, bool, error)
}

// SetupGuidanceProvider is implemented by providers that can describe how a
// user should configure or verify the provider for one context.
type SetupGuidanceProvider interface {
	SetupGuidance(RuntimeContext) SetupGuidance
}
