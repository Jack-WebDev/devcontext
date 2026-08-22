package provider_test

import (
	"reflect"
	"testing"

	"devctx/packages/core/provider"
)

type fakeProvider struct{}

var _ provider.Provider = fakeProvider{}

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
	for _, fieldName := range []string{"ClaudeDir", "CodexDir"} {
		if _, ok := contextPathsType.FieldByName(fieldName); ok {
			t.Fatalf("provider.ContextPaths exposes %s, want only provider-owned storage", fieldName)
		}
	}
}
