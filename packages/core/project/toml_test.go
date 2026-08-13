package project_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
)

func TestDecodeProjectBindingsTOMLReturnsEmptyForEmptyInput(t *testing.T) {
	bindings, err := project.DecodeProjectBindingsTOML(nil)
	if err != nil {
		t.Fatalf("decode empty project bindings: %v", err)
	}
	if len(bindings) != 0 {
		t.Fatalf("binding count = %d, want 0", len(bindings))
	}
}

func TestDecodeProjectBindingsTOMLLoadsMultiProjectBindings(t *testing.T) {
	input := []byte(`
[[projects]]
path = "/home/jack/projects/constructa"
context = "personal"
created_at = 2026-08-13T10:30:00Z

[[projects]]
path = "/home/jack/work/internal-api"
context = "company"
created_at = 2026-08-13T11:00:00Z
`)

	bindings, err := project.DecodeProjectBindingsTOML(input)
	if err != nil {
		t.Fatalf("decode project bindings: %v", err)
	}

	want := []project.Binding{
		{
			ProjectPath: "/home/jack/projects/constructa",
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC),
		},
		{
			ProjectPath: "/home/jack/work/internal-api",
			ContextID:   devcontext.MustID("company"),
			CreatedAt:   time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC),
		},
	}
	if !reflect.DeepEqual(bindings, want) {
		t.Fatalf("bindings = %#v, want %#v", bindings, want)
	}
}

func TestDecodeProjectBindingsTOMLRejectsMalformedInputs(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMessage string
	}{
		{
			name:        "malformed toml",
			input:       `[[projects]] path = `,
			wantMessage: "invalid project bindings",
		},
		{
			name: "missing path",
			input: `
[[projects]]
context = "personal"
created_at = 2026-08-13T10:30:00Z
`,
			wantMessage: "missing projects[0].path",
		},
		{
			name: "empty path",
			input: `
[[projects]]
path = ""
context = "personal"
created_at = 2026-08-13T10:30:00Z
`,
			wantMessage: "projects[0].path cannot be empty",
		},
		{
			name: "missing context",
			input: `
[[projects]]
path = "/home/jack/projects/constructa"
created_at = 2026-08-13T10:30:00Z
`,
			wantMessage: "missing projects[0].context",
		},
		{
			name: "invalid context",
			input: `
[[projects]]
path = "/home/jack/projects/constructa"
context = "Personal"
created_at = 2026-08-13T10:30:00Z
`,
			wantMessage: "invalid context ID",
		},
		{
			name: "missing created_at",
			input: `
[[projects]]
path = "/home/jack/projects/constructa"
context = "personal"
`,
			wantMessage: "missing projects[0].created_at",
		},
		{
			name: "unknown field",
			input: `
[[projects]]
path = "/home/jack/projects/constructa"
context = "personal"
created_at = 2026-08-13T10:30:00Z
unknown = true
`,
			wantMessage: "unsupported field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := project.DecodeProjectBindingsTOML([]byte(tt.input))
			if !errors.Is(err, project.ErrInvalidProjectBindings) {
				t.Fatalf("error = %v, want %v", err, project.ErrInvalidProjectBindings)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestDecodeProjectBindingsTOMLRejectsDuplicateProjectEntries(t *testing.T) {
	input := []byte(`
[[projects]]
path = "/home/jack/projects/constructa"
context = "personal"
created_at = 2026-08-13T10:30:00Z

[[projects]]
path = "/home/jack/projects/constructa"
context = "company"
created_at = 2026-08-13T11:00:00Z
`)

	_, err := project.DecodeProjectBindingsTOML(input)
	if !errors.Is(err, project.ErrDuplicateProjectBinding) {
		t.Fatalf("error = %v, want %v", err, project.ErrDuplicateProjectBinding)
	}
}

func TestEncodeProjectBindingsTOMLIsDeterministic(t *testing.T) {
	bindings := []project.Binding{
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	}

	encoded, err := project.EncodeProjectBindingsTOML(bindings)
	if err != nil {
		t.Fatalf("encode project bindings: %v", err)
	}

	want := `[[projects]]
path = "/home/jack/projects/constructa"
context = "personal"
created_at = 2026-08-13T10:30:00Z

[[projects]]
path = "/home/jack/work/internal-api"
context = "company"
created_at = 2026-08-13T11:00:00Z
`
	if string(encoded) != want {
		t.Fatalf("encoded TOML =\n%s\nwant:\n%s", encoded, want)
	}
}

func TestEncodeProjectBindingsTOMLRoundTripsThroughDecoder(t *testing.T) {
	bindings := []project.Binding{
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	}

	encoded, err := project.EncodeProjectBindingsTOML(bindings)
	if err != nil {
		t.Fatalf("encode project bindings: %v", err)
	}

	decoded, err := project.DecodeProjectBindingsTOML(encoded)
	if err != nil {
		t.Fatalf("decode encoded project bindings: %v", err)
	}

	want := []project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		projectBinding("/home/jack/work/internal-api", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Fatalf("decoded bindings = %#v, want %#v", decoded, want)
	}
}

func TestEncodeProjectBindingsTOMLRejectsInvalidBindings(t *testing.T) {
	_, err := project.EncodeProjectBindingsTOML([]project.Binding{
		projectBinding("", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
	})
	if !errors.Is(err, project.ErrInvalidProjectBindings) {
		t.Fatalf("error = %v, want %v", err, project.ErrInvalidProjectBindings)
	}

	_, err = project.EncodeProjectBindingsTOML([]project.Binding{
		projectBinding("/home/jack/projects/constructa", "personal", time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC)),
		projectBinding("/home/jack/projects/constructa", "company", time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC)),
	})
	if !errors.Is(err, project.ErrDuplicateProjectBinding) {
		t.Fatalf("error = %v, want %v", err, project.ErrDuplicateProjectBinding)
	}
}

func projectBinding(path string, contextID string, createdAt time.Time) project.Binding {
	return project.Binding{
		ProjectPath: project.Path(path),
		ContextID:   devcontext.MustID(contextID),
		CreatedAt:   createdAt,
	}
}
