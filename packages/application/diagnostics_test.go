package application

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGetDiagnosticsReturnsAnEmptyStructuredContractBeforeChecksAreAdded(t *testing.T) {
	fixture := newApplicationFixture(t)

	state, appErr := fixture.service().GetDiagnostics(GetDiagnosticsRequest{ContextID: "personal"})
	if appErr != nil {
		t.Fatalf("get diagnostics: %v", appErr)
	}
	if !reflect.DeepEqual(state, DiagnosticsState{Groups: []DiagnosticGroup{}}) {
		t.Fatalf("diagnostics = %#v, want an empty structured state", state)
	}
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
