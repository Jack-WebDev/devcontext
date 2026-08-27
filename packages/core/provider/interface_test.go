package provider_test

import (
	"reflect"
	"testing"

	"devctx/packages/core/provider"
)

type fakeProvider struct{}

var _ provider.Provider = fakeProvider{}
var _ provider.GlobalCredentialDetector = fakeProvider{}
var _ provider.CredentialMetadataExtractor = fakeProvider{}
var _ provider.CredentialImporter = fakeProvider{}
var _ provider.ContextIdentityDetector = fakeProvider{}
var _ provider.SetupGuidanceProvider = fakeProvider{}

func (fakeProvider) ID() provider.ID {
	return "fake"
}

func (fakeProvider) DisplayName() string {
	return "Fake Provider"
}

func (fakeProvider) BuildEnvironment(ctx provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return provider.EnvironmentContribution{
		"FAKE_CONTEXT": ctx.ContextID,
		"FAKE_HOME":    ctx.Paths.StorageDir,
	}, nil
}

func (fakeProvider) Status(ctx provider.RuntimeContext) (provider.Status, error) {
	if !ctx.Config.Enabled {
		return provider.NotConfiguredStatus("disabled"), nil
	}
	return provider.ReadyStatus(), nil
}

func (fakeProvider) DetectGlobalCredentialSession(provider.GlobalCredentialContext) (provider.CredentialSession, bool, error) {
	return provider.CredentialSession{
		MetadataAvailable: true,
		Fields: []provider.MetadataField{
			{Label: "Account", Value: "fake@example.com"},
		},
	}, true, nil
}

func (fakeProvider) ExtractCredentialMetadata(string) ([]provider.MetadataField, bool, error) {
	return []provider.MetadataField{
		{Label: "Account", Value: "fake@example.com"},
	}, true, nil
}

func (fakeProvider) ImportCredentials(provider.CredentialImportContext) error {
	return nil
}

func (fakeProvider) DetectContextIdentity(provider.RuntimeContext) (provider.Identity, bool, error) {
	return provider.Identity{
		Fields: []provider.MetadataField{
			{Label: "Account", Value: "fake@example.com"},
		},
	}, true, nil
}

func (fakeProvider) SetupGuidance(provider.RuntimeContext) provider.SetupGuidance {
	return provider.SetupGuidance{
		Message:    "Configure Fake Provider.",
		ActionHint: "Run fake auth login.",
	}
}

func TestProviderInterfaceAllowsGenericProviderUse(t *testing.T) {
	var integration provider.Provider = fakeProvider{}
	ctx := provider.RuntimeContext{
		ContextID: "client-a",
		Config: provider.Config{
			Enabled: true,
		},
		Paths: provider.ContextPaths{
			RootDir:    "/home/alex/.devctx/contexts/client-a",
			StorageDir: "/home/alex/.devctx/contexts/client-a/providers/fake",
		},
	}

	if integration.ID() != "fake" {
		t.Fatalf("id = %q, want %q", integration.ID(), "fake")
	}
	if integration.DisplayName() != "Fake Provider" {
		t.Fatalf("display name = %q, want %q", integration.DisplayName(), "Fake Provider")
	}

	environment, err := integration.BuildEnvironment(ctx)
	if err != nil {
		t.Fatalf("build environment: %v", err)
	}
	wantEnvironment := provider.EnvironmentContribution{
		"FAKE_CONTEXT": "client-a",
		"FAKE_HOME":    "/home/alex/.devctx/contexts/client-a/providers/fake",
	}
	if !reflect.DeepEqual(environment, wantEnvironment) {
		t.Fatalf("environment = %#v, want %#v", environment, wantEnvironment)
	}

	status, err := integration.Status(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.State != provider.StatusReady {
		t.Fatalf("status state = %q, want %q", status.State, provider.StatusReady)
	}
}

func TestRuntimeContextPathsExposeOnlyProviderStorageDir(t *testing.T) {
	contextPathsType := reflect.TypeOf(provider.ContextPaths{})

	if _, ok := contextPathsType.FieldByName("StorageDir"); !ok {
		t.Fatal("provider.ContextPaths is missing StorageDir")
	}
	for _, fieldName := range []string{"ClaudeDir", "CodexDir", "VSCodeDir", "VSCodeUserDataDir"} {
		if _, ok := contextPathsType.FieldByName(fieldName); ok {
			t.Fatalf("provider.ContextPaths exposes %s, want only provider-owned storage", fieldName)
		}
	}
}

func TestProviderOptionalCapabilityInterfacesUsePlainRuntimeValues(t *testing.T) {
	var integration provider.Provider = fakeProvider{}
	runtimeContext := provider.RuntimeContext{
		ContextID: "client-a",
		Config:    provider.Config{Enabled: true},
		Paths: provider.ContextPaths{
			RootDir:    "/home/alex/.devctx/contexts/client-a",
			StorageDir: "/home/alex/.devctx/contexts/client-a/providers/fake",
		},
	}

	globalDetector := integration.(provider.GlobalCredentialDetector)
	session, ok, err := globalDetector.DetectGlobalCredentialSession(provider.GlobalCredentialContext{
		UserHomeDir: "/home/alex",
	})
	if err != nil || !ok || !session.MetadataAvailable {
		t.Fatalf("global session = %#v ok=%t err=%v", session, ok, err)
	}

	importer := integration.(provider.CredentialImporter)
	if err := importer.ImportCredentials(provider.CredentialImportContext{
		UserHomeDir: "/home/alex",
		Runtime:     runtimeContext,
	}); err != nil {
		t.Fatalf("import credentials: %v", err)
	}

	identityDetector := integration.(provider.ContextIdentityDetector)
	identity, ok, err := identityDetector.DetectContextIdentity(runtimeContext)
	if err != nil || !ok || !reflect.DeepEqual(identity.Fields, session.Fields) {
		t.Fatalf("identity = %#v ok=%t err=%v, want session fields %#v", identity, ok, err, session.Fields)
	}

	metadataExtractor := integration.(provider.CredentialMetadataExtractor)
	fields, ok, err := metadataExtractor.ExtractCredentialMetadata("/home/alex/.fake/auth.json")
	if err != nil || !ok || !reflect.DeepEqual(fields, session.Fields) {
		t.Fatalf("metadata fields = %#v ok=%t err=%v, want %#v", fields, ok, err, session.Fields)
	}

	guidance := integration.(provider.SetupGuidanceProvider).SetupGuidance(runtimeContext)
	if guidance.ActionHint == "" || guidance.Message == "" {
		t.Fatalf("setup guidance = %#v, want message and action hint", guidance)
	}
}
