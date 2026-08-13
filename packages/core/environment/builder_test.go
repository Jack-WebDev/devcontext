package environment_test

import (
	"errors"
	"reflect"
	"testing"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/environment"
	"devctx/packages/core/provider"
)

func TestBuildPreservesUnrelatedParentVariables(t *testing.T) {
	variables := environment.Build(
		[]string{
			"PATH=/usr/bin",
			"SHELL=/bin/bash",
		},
		provider.EnvironmentContribution{
			provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/personal/codex",
		},
	)

	assertVariable(t, variables, "PATH", "/usr/bin")
	assertVariable(t, variables, "SHELL", "/bin/bash")
	assertVariable(t, variables, provider.CodexHomeEnvVar, "/home/alex/.devctx/contexts/personal/codex")
}

func TestBuildReplacesDuplicateKeys(t *testing.T) {
	variables := environment.Build(
		[]string{
			"PATH=/bin",
			"PATH=/usr/bin",
			"CODEX_HOME=/global/codex",
		},
		provider.EnvironmentContribution{
			provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/company/codex",
		},
	)

	assertVariable(t, variables, "PATH", "/usr/bin")
	assertVariable(t, variables, provider.CodexHomeEnvVar, "/home/alex/.devctx/contexts/company/codex")
}

func TestBuildForContextAddsActiveContextMarker(t *testing.T) {
	personalVariables, err := environment.BuildForContext(
		[]string{
			"DEVCTX_CONTEXT=outside",
			"PATH=/usr/bin",
		},
		devcontext.MustID("personal"),
		provider.EnvironmentContribution{
			provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/personal/codex",
		},
	)
	if err != nil {
		t.Fatalf("build personal context: %v", err)
	}
	companyVariables, err := environment.BuildForContext(
		nil,
		devcontext.MustID("company"),
		provider.EnvironmentContribution{
			provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/company/codex",
		},
	)
	if err != nil {
		t.Fatalf("build company context: %v", err)
	}

	assertVariable(t, personalVariables, environment.ActiveContextEnvVar, "personal")
	assertVariable(t, personalVariables, "PATH", "/usr/bin")
	assertVariable(t, companyVariables, environment.ActiveContextEnvVar, "company")
	if personalVariables[environment.ActiveContextEnvVar] == companyVariables[environment.ActiveContextEnvVar] {
		t.Fatalf("active context markers match, want selected context IDs")
	}
}

func TestBuildForContextRejectsMissingContextID(t *testing.T) {
	var contextID devcontext.ID
	_, err := environment.BuildForContext(nil, contextID)

	if !errors.Is(err, environment.ErrMissingContextID) {
		t.Fatalf("error = %v, want %v", err, environment.ErrMissingContextID)
	}
}

func TestEnvironReturnsDeterministicEntries(t *testing.T) {
	variables := environment.Variables{
		"SHELL":                  "/bin/bash",
		provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/personal/codex",
		"PATH":                   "/usr/bin",
	}

	got := variables.Environ()
	want := []string{
		"CODEX_HOME=/home/alex/.devctx/contexts/personal/codex",
		"PATH=/usr/bin",
		"SHELL=/bin/bash",
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("environ = %#v, want %#v", got, want)
	}
}

func assertVariable(t *testing.T, variables environment.Variables, key string, want string) {
	t.Helper()

	value, ok := variables[key]
	if !ok {
		t.Fatalf("missing environment variable %q", key)
	}
	if value != want {
		t.Fatalf("environment[%q] = %q, want %q", key, value, want)
	}
}
