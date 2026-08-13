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
		"FAKE_HOME":    ctx.Paths.RootDir,
	}, nil
}

func (fakeProvider) Status(ctx provider.RuntimeContext) (provider.Status, error) {
	if !ctx.Config.Enabled {
		return provider.Status{Message: "disabled"}, nil
	}
	return provider.Status{Message: "configured"}, nil
}

func TestProviderInterfaceAllowsGenericProviderUse(t *testing.T) {
	var integration provider.Provider = fakeProvider{}
	ctx := provider.RuntimeContext{
		ContextID: "client-a",
		Config: provider.Config{
			Enabled: true,
		},
		Paths: provider.ContextPaths{
			RootDir: "/home/alex/.devctx/contexts/client-a",
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
		"FAKE_HOME":    "/home/alex/.devctx/contexts/client-a",
	}
	if !reflect.DeepEqual(environment, wantEnvironment) {
		t.Fatalf("environment = %#v, want %#v", environment, wantEnvironment)
	}

	status, err := integration.Status(ctx)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Message != "configured" {
		t.Fatalf("status message = %q, want %q", status.Message, "configured")
	}
}
