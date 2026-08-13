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

func TestBuildForContextIsolationMatrix(t *testing.T) {
	tests := []struct {
		name          string
		contextID     devcontext.ID
		parent        []string
		contributions []provider.EnvironmentContribution
		want          environment.Variables
		wantRedacted  []string
	}{
		{
			name:      "personal context overrides provider homes and keeps inherited variables",
			contextID: devcontext.MustID("personal"),
			parent: []string{
				"PATH=/usr/bin",
				"CODEX_HOME=/parent/codex",
				"API_TOKEN=personal-secret",
			},
			contributions: []provider.EnvironmentContribution{
				{
					provider.CodexHomeEnvVar:       "/home/alex/.devctx/contexts/personal/codex",
					provider.ClaudeConfigDirEnvVar: "/home/alex/.devctx/contexts/personal/claude",
				},
			},
			want: environment.Variables{
				"PATH":                          "/usr/bin",
				"API_TOKEN":                     "personal-secret",
				environment.ActiveContextEnvVar: "personal",
				provider.CodexHomeEnvVar:        "/home/alex/.devctx/contexts/personal/codex",
				provider.ClaudeConfigDirEnvVar:  "/home/alex/.devctx/contexts/personal/claude",
			},
			wantRedacted: []string{
				"API_TOKEN=<redacted>",
				"CLAUDE_CONFIG_DIR=/home/alex/.devctx/contexts/personal/claude",
				"CODEX_HOME=/home/alex/.devctx/contexts/personal/codex",
				"DEVCTX_CONTEXT=personal",
				"PATH=/usr/bin",
			},
		},
		{
			name:      "company context uses distinct provider homes",
			contextID: devcontext.MustID("company"),
			parent:    []string{"PATH=/usr/bin"},
			contributions: []provider.EnvironmentContribution{
				{
					provider.CodexHomeEnvVar:       "/home/alex/.devctx/contexts/company/codex",
					provider.ClaudeConfigDirEnvVar: "/home/alex/.devctx/contexts/company/claude",
				},
			},
			want: environment.Variables{
				"PATH":                          "/usr/bin",
				environment.ActiveContextEnvVar: "company",
				provider.CodexHomeEnvVar:        "/home/alex/.devctx/contexts/company/codex",
				provider.ClaudeConfigDirEnvVar:  "/home/alex/.devctx/contexts/company/claude",
			},
		},
		{
			name:      "arbitrary context can disable provider contributions by omission",
			contextID: devcontext.MustID("client-a"),
			parent: []string{
				"PATH=/usr/bin",
				"CLAUDE_CONFIG_DIR=/parent/claude",
			},
			contributions: []provider.EnvironmentContribution{
				{
					provider.CodexHomeEnvVar: "/home/alex/.devctx/contexts/client-a/codex",
				},
			},
			want: environment.Variables{
				"PATH":                          "/usr/bin",
				environment.ActiveContextEnvVar: "client-a",
				provider.CodexHomeEnvVar:        "/home/alex/.devctx/contexts/client-a/codex",
				provider.ClaudeConfigDirEnvVar:  "/parent/claude",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := environment.BuildForContext(tt.parent, tt.contextID, tt.contributions...)
			if err != nil {
				t.Fatalf("build environment: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("variables = %#v, want %#v", got, tt.want)
			}
			if tt.wantRedacted != nil {
				if redacted := got.RedactedEnviron(); !reflect.DeepEqual(redacted, tt.wantRedacted) {
					t.Fatalf("redacted environ = %#v, want %#v", redacted, tt.wantRedacted)
				}
			}
		})
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
