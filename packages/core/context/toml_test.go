package context_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

func TestDecodeContextTOMLLoadsPersonalCompanyAndArbitraryContexts(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectedID devcontext.ID
		want       devcontext.Context
	}{
		{
			name: "personal",
			input: `
id = "personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"

[providers.claude]
enabled = true

[providers.codex]
enabled = true
`,
			expectedID: devcontext.MustID("personal"),
			want: devcontext.Context{
				ID:   devcontext.MustID("personal"),
				Name: "Personal",
				Tool: codingtool.DefaultConfig(),
				Providers: provider.Configs{
					"claude": {Enabled: true},
					"codex":  {Enabled: true},
				},
				CreatedAt: time.Date(2026, 8, 13, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			name: "company",
			input: `
id = "company"
name = "Company"
created_at = 2026-08-13T10:45:00Z

[editor]
type = "vscode"

[providers.claude]
enabled = true

[providers.codex]
enabled = true

[metadata]
kind = "default"
`,
			expectedID: devcontext.MustID("company"),
			want: devcontext.Context{
				ID:   devcontext.MustID("company"),
				Name: "Company",
				Tool: codingtool.DefaultConfig(),
				Providers: provider.Configs{
					"claude": {Enabled: true},
					"codex":  {Enabled: true},
				},
				Metadata: devcontext.Metadata{
					"kind": "default",
				},
				CreatedAt: time.Date(2026, 8, 13, 10, 45, 0, 0, time.UTC),
			},
		},
		{
			name: "arbitrary",
			input: `
id = "client-a"
name = "Client A"
created_at = 2026-08-13T11:00:00Z

[editor]
type = "vscode"
executable_override = "/opt/code"

[providers.claude]
enabled = true

[providers.claude.options]
profile = "client-a"

[providers.codex]
enabled = false

[providers.future-provider]
enabled = true

[providers.future-provider.options]
mode = "sandbox"

[metadata]
owner = "client-a"
`,
			expectedID: devcontext.MustID("client-a"),
			want: devcontext.Context{
				ID:   devcontext.MustID("client-a"),
				Name: "Client A",
				Tool: codingtool.Config{
					Type:               codingtool.TypeVSCode,
					ExecutableOverride: "/opt/code",
				},
				Providers: provider.Configs{
					"claude": {
						Enabled: true,
						Options: provider.Options{
							"profile": "client-a",
						},
					},
					"codex": {Enabled: false},
					"future-provider": {
						Enabled: true,
						Options: provider.Options{
							"mode": "sandbox",
						},
					},
				},
				Metadata: devcontext.Metadata{
					"owner": "client-a",
				},
				CreatedAt: time.Date(2026, 8, 13, 11, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := devcontext.DecodeContextTOML([]byte(tt.input), tt.expectedID)
			if err != nil {
				t.Fatalf("decode context TOML: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("decoded context = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeContextTOMLRejectsIDMismatch(t *testing.T) {
	input := []byte(`
id = "personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"
`)

	_, err := devcontext.DecodeContextTOML(input, devcontext.MustID("company"))
	if !errors.Is(err, devcontext.ErrContextIDMismatch) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextIDMismatch)
	}
}

func TestDecodeContextTOMLRejectsStructurallyInvalidValues(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMessage string
	}{
		{
			name:        "malformed toml",
			input:       `id = `,
			wantMessage: "invalid context configuration",
		},
		{
			name: "missing id",
			input: `
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"
`,
			wantMessage: "missing id",
		},
		{
			name: "invalid id",
			input: `
id = "Personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"
`,
			wantMessage: "invalid context ID",
		},
		{
			name: "missing name",
			input: `
id = "personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"
`,
			wantMessage: "missing name",
		},
		{
			name: "missing editor type",
			input: `
id = "personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
`,
			wantMessage: "missing codingtool.type",
		},
		{
			name: "provider missing enabled",
			input: `
id = "personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z

[editor]
type = "vscode"

[providers.claude]
`,
			wantMessage: "missing providers.claude.enabled",
		},
		{
			name: "unknown field",
			input: `
id = "personal"
name = "Personal"
created_at = 2026-08-13T10:30:00Z
unknown = true

[editor]
type = "vscode"
`,
			wantMessage: `unsupported field "unknown"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := devcontext.DecodeContextTOML([]byte(tt.input), devcontext.MustID("personal"))
			if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
				t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidContextConfig)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestEncodeContextTOMLIsDeterministic(t *testing.T) {
	ctx := contextWithClaudeAndCodex()
	ctx.Providers["future-provider"] = provider.Config{
		Enabled: true,
		Options: provider.Options{
			"mode": "sandbox",
		},
	}
	ctx.Metadata["team"] = "platform"

	encoded, err := devcontext.EncodeContextTOML(ctx)
	if err != nil {
		t.Fatalf("encode context TOML: %v", err)
	}

	want := `id = "client-a"
name = "Client A"
created_at = 2026-08-13T12:30:00Z

[editor]
type = "vscode"
executable_override = "/opt/code"

[providers.claude]
enabled = true

[providers.claude.options]
profile = "client-a"

[providers.codex]
enabled = true

[providers.future-provider]
enabled = true

[providers.future-provider.options]
mode = "sandbox"

[metadata]
owner = "client-a"
team = "platform"
`
	if string(encoded) != want {
		t.Fatalf("encoded TOML =\n%s\nwant:\n%s", encoded, want)
	}
}

func TestEncodeContextTOMLRoundTripsThroughDecoder(t *testing.T) {
	ctx := contextWithClaudeAndCodex()

	encoded, err := devcontext.EncodeContextTOML(ctx)
	if err != nil {
		t.Fatalf("encode context TOML: %v", err)
	}

	decoded, err := devcontext.DecodeContextTOML(encoded, ctx.ID)
	if err != nil {
		t.Fatalf("decode encoded context TOML: %v", err)
	}

	if !reflect.DeepEqual(decoded, ctx) {
		t.Fatalf("decoded context = %#v, want %#v", decoded, ctx)
	}
}

func TestEncodeContextTOMLRejectsInvalidContext(t *testing.T) {
	_, err := devcontext.EncodeContextTOML(devcontext.Context{
		Name:      "Missing ID",
		Tool:      codingtool.DefaultConfig(),
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	})
	if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidContextConfig)
	}

	_, err = devcontext.EncodeContextTOML(devcontext.Context{
		ID:        devcontext.MustID("personal"),
		Name:      "Personal",
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	})
	if !errors.Is(err, devcontext.ErrInvalidContextConfig) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrInvalidContextConfig)
	}
}

func contextWithClaudeAndCodex() devcontext.Context {
	return devcontext.Context{
		ID:   devcontext.MustID("client-a"),
		Name: "Client A",
		Tool: codingtool.Config{
			Type:               codingtool.TypeVSCode,
			ExecutableOverride: "/opt/code",
		},
		Providers: provider.Configs{
			"claude": {
				Enabled: true,
				Options: provider.Options{
					"profile": "client-a",
				},
			},
			"codex": {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"owner": "client-a",
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}
