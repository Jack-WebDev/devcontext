package provider_test

import (
	"testing"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func TestCodexProviderBuildsIsolatedEnvironmentContribution(t *testing.T) {
	integration := provider.CodexProvider{}

	personalContext := providerRuntimeContext(t, "personal", provider.CodexID)
	companyContext := providerRuntimeContext(t, "company", provider.CodexID)

	personalEnvironment, err := integration.BuildEnvironment(personalContext)
	if err != nil {
		t.Fatalf("build personal environment: %v", err)
	}
	companyEnvironment, err := integration.BuildEnvironment(companyContext)
	if err != nil {
		t.Fatalf("build company environment: %v", err)
	}

	assertOnlyEnvironmentValue(t, personalEnvironment, provider.CodexHomeEnvVar, "/home/alex/.devctx/contexts/personal/providers/codex")
	assertOnlyEnvironmentValue(t, companyEnvironment, provider.CodexHomeEnvVar, "/home/alex/.devctx/contexts/company/providers/codex")
	if personalEnvironment[provider.CodexHomeEnvVar] == companyEnvironment[provider.CodexHomeEnvVar] {
		t.Fatalf("Codex homes match, want isolated paths")
	}
}

func TestClaudeProviderBuildsIsolatedEnvironmentContribution(t *testing.T) {
	integration := provider.ClaudeProvider{}

	personalContext := providerRuntimeContext(t, "personal", provider.ClaudeID)
	companyContext := providerRuntimeContext(t, "company", provider.ClaudeID)

	personalEnvironment, err := integration.BuildEnvironment(personalContext)
	if err != nil {
		t.Fatalf("build personal environment: %v", err)
	}
	companyEnvironment, err := integration.BuildEnvironment(companyContext)
	if err != nil {
		t.Fatalf("build company environment: %v", err)
	}

	assertOnlyEnvironmentValue(t, personalEnvironment, provider.ClaudeConfigDirEnvVar, "/home/alex/.devctx/contexts/personal/providers/claude")
	assertOnlyEnvironmentValue(t, companyEnvironment, provider.ClaudeConfigDirEnvVar, "/home/alex/.devctx/contexts/company/providers/claude")
	if personalEnvironment[provider.ClaudeConfigDirEnvVar] == companyEnvironment[provider.ClaudeConfigDirEnvVar] {
		t.Fatalf("Claude config dirs match, want isolated paths")
	}
}

func providerRuntimeContext(t *testing.T, contextID string, providerID provider.ID) provider.RuntimeContext {
	t.Helper()

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/alex", nil
	})
	derivedPaths, err := filesystem.DeriveContextPaths(paths, devcontext.MustID(contextID))
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}

	return provider.RuntimeContext{
		ContextID: contextID,
		Config: provider.Config{
			Enabled: true,
		},
		Paths: provider.ContextPaths{
			RootDir:           derivedPaths.RootDir,
			StorageDir:        derivedPaths.ProviderStorageDir(providerID),
			VSCodeDir:         derivedPaths.VSCodeDir,
			VSCodeUserDataDir: derivedPaths.VSCodeUserDataDir,
		},
	}
}

func assertOnlyEnvironmentValue(t *testing.T, environment provider.EnvironmentContribution, key string, want string) {
	t.Helper()

	if len(environment) != 1 {
		t.Fatalf("environment count = %d, want 1", len(environment))
	}
	if environment[key] != want {
		t.Fatalf("environment[%q] = %q, want %q", key, environment[key], want)
	}
}
