package application

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func TestGetDiagnosticsReportsContextProviderAndToolReadiness(t *testing.T) {
	fixture := newApplicationFixture(t)
	ctx := fixture.context("personal", "Personal")
	paths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	credentialPath := filepath.Join(paths.ProviderStorageDir(fixture.provider.id), "credentials.json")
	fixture.provider.credentialDiagnosticFiles = []provider.CredentialDiagnosticFile{{Label: "Credentials", Path: credentialPath}}
	fixture.provider.hasIdentity = true
	fixture.provider.identity = provider.Identity{Fields: []provider.MetadataField{{Label: "Email", Value: "developer@example.com"}}}
	fixture.writeContext(t, ctx)
	writeFile(t, credentialPath, []byte("credential contents are never read"))

	state, appErr := fixture.service().GetDiagnostics(GetDiagnosticsRequest{ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("get diagnostics: %v", appErr)
	}
	if got := diagnosticGroupIDs(state.Groups); !reflect.DeepEqual(got, []string{"context-filesystem", "providers", "coding-tool"}) {
		t.Fatalf("diagnostic groups = %#v", got)
	}
	if check := diagnosticCheck(state.Groups, "context-directory"); check.Severity != DiagnosticSeverityReady || !hasPathDetail(check) {
		t.Fatalf("context directory check = %#v", check)
	}
	if check := diagnosticCheck(state.Groups, "provider-fake-credential-0"); check.Severity != DiagnosticSeverityReady || !hasPathDetail(check) {
		t.Fatalf("credential check = %#v", check)
	}
	if check := diagnosticCheck(state.Groups, "provider-fake-identity"); check.Severity != DiagnosticSeverityReady || !reflect.DeepEqual(check.Details, []DiagnosticDetail{{Label: "Email", Value: "developer@example.com"}}) {
		t.Fatalf("identity check = %#v", check)
	}
	if check := diagnosticCheck(state.Groups, "tool-executable"); check.Severity != DiagnosticSeverityReady || !hasPathDetail(check) {
		t.Fatalf("tool executable check = %#v", check)
	}
}

func TestGetDiagnosticsReportsMissingContextStorageAndCredentials(t *testing.T) {
	fixture := newApplicationFixture(t)
	ctx := fixture.context("personal", "Personal")
	paths, err := filesystem.DeriveContextPaths(fixture.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	fixture.provider.credentialDiagnosticFiles = []provider.CredentialDiagnosticFile{{Label: "Credentials", Path: filepath.Join(paths.ProviderStorageDir(fixture.provider.id), "credentials.json")}}
	fixture.writeContext(t, ctx)
	removeAll(t, paths.ToolStorageDir(ctx.Tool.DefaultTool))

	state, appErr := fixture.service().GetDiagnostics(GetDiagnosticsRequest{ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("get diagnostics: %v", appErr)
	}
	if check := diagnosticCheck(state.Groups, "context-storage-completeness"); check.Severity != DiagnosticSeverityBlocked || !hasPathDetail(check) {
		t.Fatalf("storage completeness check = %#v", check)
	}
	if check := diagnosticCheck(state.Groups, "provider-fake-credential-0"); check.Severity != DiagnosticSeverityNeedsAttention || !hasPathDetail(check) {
		t.Fatalf("credential check = %#v", check)
	}
	if check := diagnosticCheck(state.Groups, "tool-storage"); check.Severity != DiagnosticSeverityBlocked || !hasPathDetail(check) {
		t.Fatalf("tool storage check = %#v", check)
	}
}

func diagnosticGroupIDs(groups []DiagnosticGroup) []string {
	ids := make([]string, len(groups))
	for i, group := range groups {
		ids[i] = group.ID
	}
	return ids
}

func diagnosticCheck(groups []DiagnosticGroup, id string) DiagnosticCheck {
	for _, group := range groups {
		for _, check := range group.Checks {
			if check.ID == id {
				return check
			}
		}
	}
	return DiagnosticCheck{}
}

func hasPathDetail(check DiagnosticCheck) bool {
	for _, detail := range check.Details {
		if detail.IsPath {
			return true
		}
	}
	return false
}

func TestDiagnosticDetailMarksPathsForFrontendDisclosure(t *testing.T) {
	state := DiagnosticsState{Groups: []DiagnosticGroup{{
		ID:    "context-storage",
		Label: "Context storage",
		Checks: []DiagnosticCheck{{
			ID:       "context-root",
			Severity: DiagnosticSeverityNeedsAttention,
			Label:    "Context directory",
			Message:  "The context directory needs repair.",
			Details:  []DiagnosticDetail{{Label: "Location", Value: "/contexts/personal", IsPath: true}},
		}},
	}}}

	encoded, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal diagnostics: %v", err)
	}
	const want = `{"groups":[{"id":"context-storage","label":"Context storage","checks":[{"id":"context-root","severity":"needs_attention","label":"Context directory","message":"The context directory needs repair.","details":[{"label":"Location","value":"/contexts/personal","isPath":true}]}]}]}`
	if string(encoded) != want {
		t.Fatalf("diagnostics JSON = %s, want %s", encoded, want)
	}
}
