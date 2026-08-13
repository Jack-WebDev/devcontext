package environment_test

import (
	"reflect"
	"testing"

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
