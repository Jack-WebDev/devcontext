package filesystem_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/provider"
)

func TestContextCreationImportsCredentialsForRegisteredFutureProvider(t *testing.T) {
	homeDir := t.TempDir()
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) { return homeDir, nil })
	if err := os.WriteFile(filepath.Join(homeDir, "future.credential"), []byte("future-credential"), 0o600); err != nil {
		t.Fatalf("write future credential fixture: %v", err)
	}

	future := filesystemFixtureProvider{}
	registry := provider.MustNewRegistry([]provider.Provider{future}, future.ID())
	ctx := devcontext.DefaultPersonalContextWithProviderRegistry(time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC), registry)
	contextPaths, err := filesystem.DeriveContextPaths(paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	if err := filesystem.CreateContextDirectoryTreeWithProviderRegistryCredentialsAndPermissions(
		paths, contextPaths, ctx, registry, []string{"future"}, filesystem.NewDefaultStoragePermissions(),
	); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	destination := filepath.Join(contextPaths.ProviderStorageDir(future.ID()), "credential.json")
	credential, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("read imported future credential: %v", err)
	}
	if string(credential) != "future-credential" {
		t.Fatalf("imported credential = %q", credential)
	}

	runtime := provider.RuntimeContext{Paths: provider.ContextPaths{StorageDir: contextPaths.ProviderStorageDir(future.ID())}}
	environment, err := future.BuildEnvironment(runtime)
	if err != nil || environment["FUTURE_HOME"] != runtime.Paths.StorageDir {
		t.Fatalf("future provider environment = %#v err=%v", environment, err)
	}
	session, found, err := future.DetectGlobalCredentialSession(provider.GlobalCredentialContext{UserHomeDir: homeDir})
	if err != nil || !found || metadataValue(session.Fields, "Workspace") != "Example" {
		t.Fatalf("future provider session = %#v found=%t err=%v", session, found, err)
	}
	identity, available, err := future.DetectContextIdentity(runtime)
	if err != nil || !available || metadataValue(identity.Fields, "Workspace") != "Example" {
		t.Fatalf("future provider identity = %#v available=%t err=%v", identity, available, err)
	}
	confidence, ok := launcher.ProviderConfidenceCheck(future.ID(), future.DisplayName(), provider.ConfiguredStatus())
	if !ok || confidence.ProviderID != "future" || confidence.Label != "Future Provider" {
		t.Fatalf("future provider confidence = %#v ok=%t", confidence, ok)
	}
}

type filesystemFixtureProvider struct{}

func (filesystemFixtureProvider) ID() provider.ID { return "future" }

func (filesystemFixtureProvider) DisplayName() string { return "Future Provider" }

func (filesystemFixtureProvider) BuildEnvironment(ctx provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return provider.EnvironmentContribution{"FUTURE_HOME": ctx.Paths.StorageDir}, nil
}

func (filesystemFixtureProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ConfiguredStatus(), nil
}

func (filesystemFixtureProvider) ImportCredentials(ctx provider.CredentialImportContext) error {
	source := filepath.Join(ctx.UserHomeDir, "future.credential")
	exists, err := ctx.Files.FileExists(source)
	if err != nil || !exists {
		return err
	}
	return ctx.Files.CopyOpaqueFile(source, filepath.Join(ctx.Runtime.Paths.StorageDir, "credential.json"))
}

func (filesystemFixtureProvider) DetectGlobalCredentialSession(provider.GlobalCredentialContext) (provider.CredentialSession, bool, error) {
	return provider.CredentialSession{MetadataAvailable: true, Fields: []provider.MetadataField{{Label: "Workspace", Value: "Example"}}}, true, nil
}

func (filesystemFixtureProvider) DetectContextIdentity(provider.RuntimeContext) (provider.Identity, bool, error) {
	return provider.Identity{Fields: []provider.MetadataField{{Label: "Workspace", Value: "Example"}}}, true, nil
}

func metadataValue(fields []provider.MetadataField, label string) string {
	for _, field := range fields {
		if field.Label == label {
			return field.Value
		}
	}
	return ""
}
