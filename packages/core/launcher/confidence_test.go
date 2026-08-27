package launcher_test

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/provider"
)

func TestConfidenceStatusVariantsSerialize(t *testing.T) {
	tests := []struct {
		name   string
		status launcher.ConfidenceStatus
		want   string
	}{
		{
			name:   "ready",
			status: launcher.ConfidenceReady,
			want:   `{"status":"ready"}`,
		},
		{
			name:   "needs attention",
			status: launcher.ConfidenceNeedsAttention,
			want:   `{"status":"needs_attention"}`,
		},
		{
			name:   "blocked",
			status: launcher.ConfidenceBlocked,
			want:   `{"status":"blocked"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.status.Valid() {
				t.Fatalf("status %q is not valid", tt.status)
			}

			data, err := json.Marshal(struct {
				Status launcher.ConfidenceStatus `json:"status"`
			}{Status: tt.status})
			if err != nil {
				t.Fatalf("marshal status: %v", err)
			}
			if string(data) != tt.want {
				t.Fatalf("json = %s, want %s", data, tt.want)
			}
		})
	}
}

func TestConfidenceStatusRejectsUnknownValues(t *testing.T) {
	if launcher.ConfidenceStatus("configured").Valid() {
		t.Fatal("provider status value is valid as launch confidence")
	}
	if launcher.ConfidenceStatus("healthy").Valid() {
		t.Fatal("unapproved UI synonym is valid as launch confidence")
	}
}

func TestConfidenceCheckComponentsValidate(t *testing.T) {
	tests := []launcher.ConfidenceCheckComponent{
		launcher.ConfidenceCheckProvider,
		launcher.ConfidenceCheckVSCode,
		launcher.ConfidenceCheckIsolation,
	}

	for _, component := range tests {
		if !component.Valid() {
			t.Fatalf("component %q is not valid", component)
		}
	}

	if launcher.ConfidenceCheckComponent("editor").Valid() {
		t.Fatal("unknown confidence check component is valid")
	}
}

func TestConfidenceCheckSerializesWithOptionalActionHint(t *testing.T) {
	check := launcher.ConfidenceCheck{
		Component:  launcher.ConfidenceCheckProvider,
		ProviderID: "codex",
		Severity:   launcher.ConfidenceNeedsAttention,
		Label:      "Codex",
		Message:    "Codex is not authenticated for this context.",
		ActionHint: "Sign in to Codex for this context.",
	}

	if !check.Valid() {
		t.Fatalf("check is not valid: %#v", check)
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal check: %v", err)
	}

	want := `{"component":"provider","providerId":"codex","severity":"needs_attention","label":"Codex","message":"Codex is not authenticated for this context.","actionHint":"Sign in to Codex for this context."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
}

func TestConfidenceCheckOmitsEmptyActionHint(t *testing.T) {
	check := launcher.ConfidenceCheck{
		Component: launcher.ConfidenceCheckIsolation,
		Severity:  launcher.ConfidenceReady,
		Label:     "Isolation",
		Message:   "The context isolation directories are ready.",
	}

	if !check.Valid() {
		t.Fatalf("check is not valid: %#v", check)
	}

	data, err := json.Marshal(check)
	if err != nil {
		t.Fatalf("marshal check: %v", err)
	}

	want := `{"component":"isolation","severity":"ready","label":"Isolation","message":"The context isolation directories are ready."}`
	if string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}
}

func TestConfidenceCheckRejectsIncompleteValues(t *testing.T) {
	tests := []struct {
		name  string
		check launcher.ConfidenceCheck
	}{
		{
			name: "unknown component",
			check: launcher.ConfidenceCheck{
				Component: "editor",
				Severity:  launcher.ConfidenceReady,
				Label:     "VS Code",
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "unknown severity",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  "configured",
				Label:     "VS Code",
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "missing label",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  launcher.ConfidenceReady,
				Message:   "VS Code is ready.",
			},
		},
		{
			name: "missing message",
			check: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  launcher.ConfidenceReady,
				Label:     "VS Code",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.check.Valid() {
				t.Fatalf("check is valid: %#v", tt.check)
			}
		})
	}
}

func TestProviderConfidenceCheckMapsProviderStatus(t *testing.T) {
	tests := []struct {
		name        string
		providerID  provider.ID
		displayName string
		status      provider.Status
		want        launcher.ConfidenceCheck
	}{
		{
			name:        "claude configured",
			providerID:  provider.ClaudeID,
			displayName: "Claude",
			status:      provider.ConfiguredStatus(),
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckProvider,
				ProviderID: "claude",
				Severity:   launcher.ConfidenceReady,
				Label:      "Claude",
				Message:    "Claude is ready for this context.",
			},
		},
		{
			name:        "codex not configured",
			providerID:  provider.CodexID,
			displayName: "Codex",
			status:      provider.NotConfiguredStatus("Codex is not authenticated."),
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckProvider,
				ProviderID: "codex",
				Severity:   launcher.ConfidenceNeedsAttention,
				Label:      "Codex",
				Message:    "Codex is not authenticated.",
				ActionHint: "Open and configure Codex for this context.",
			},
		},
		{
			name:        "claude directory missing",
			providerID:  provider.ClaudeID,
			displayName: "Claude",
			status:      provider.DirectoryMissingStatus("Claude isolated provider directory is missing."),
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckProvider,
				ProviderID: "claude",
				Severity:   launcher.ConfidenceBlocked,
				Label:      "Claude",
				Message:    "Claude isolated provider directory is missing.",
				ActionHint: "Run diagnostics to repair context storage.",
			},
		},
		{
			name:        "codex unavailable",
			providerID:  provider.CodexID,
			displayName: "Codex",
			status:      provider.UnavailableStatus("Codex context directory could not be inspected."),
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckProvider,
				ProviderID: "codex",
				Severity:   launcher.ConfidenceNeedsAttention,
				Label:      "Codex",
				Message:    "Codex context directory could not be inspected.",
				ActionHint: "Run diagnostics to inspect Codex.",
			},
		},
		{
			name:        "unknown provider status",
			providerID:  provider.CodexID,
			displayName: "Codex",
			status:      provider.Status{State: "expired"},
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckProvider,
				ProviderID: "codex",
				Severity:   launcher.ConfidenceNeedsAttention,
				Label:      "Codex",
				Message:    "Codex readiness could not be determined.",
				ActionHint: "Run diagnostics to inspect Codex.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := launcher.ProviderConfidenceCheck(tt.providerID, tt.displayName, tt.status)
			if !ok {
				t.Fatal("provider confidence check was not derived")
			}
			if got != tt.want {
				t.Fatalf("check = %#v, want %#v", got, tt.want)
			}
			if !got.Valid() {
				t.Fatalf("check is not valid: %#v", got)
			}
		})
	}
}

func TestProviderConfidenceCheckIncludesRegisteredFutureProviders(t *testing.T) {
	check, ok := launcher.ProviderConfidenceCheck("fake", "Fake Provider", provider.ConfiguredStatus())
	if !ok || check.Component != launcher.ConfidenceCheckProvider || check.ProviderID != "fake" {
		t.Fatalf("future provider confidence check = %#v ok=%t", check, ok)
	}
}

func TestVSCodeConfidenceCheckMapsExecutableReadiness(t *testing.T) {
	tests := []struct {
		name       string
		executable codingtool.Executable
		err        error
		want       launcher.ConfidenceCheck
	}{
		{
			name:       "ready",
			executable: "/usr/local/bin/code",
			want: launcher.ConfidenceCheck{
				Component: launcher.ConfidenceCheckVSCode,
				Severity:  launcher.ConfidenceReady,
				Label:     "VS Code",
				Message:   "VS Code is available for launch.",
			},
		},
		{
			name: "missing executable",
			err: &codingtool.ExecutableNotFoundError{
				ToolID:   codingtool.VSCodeID,
				Candidates: []string{"code"},
			},
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckVSCode,
				Severity:   launcher.ConfidenceBlocked,
				Label:      "VS Code",
				Message:    "Dev Context could not find a VS Code command to launch.",
				ActionHint: "Install the VS Code command line launcher or configure the VS Code executable.",
			},
		},
		{
			name: "invalid executable",
			err: &codingtool.ExecutableNotExecutableError{
				ToolID: codingtool.VSCodeID,
				Path:     "/tmp/code",
			},
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckVSCode,
				Severity:   launcher.ConfidenceBlocked,
				Label:      "VS Code",
				Message:    "The configured VS Code command cannot be run.",
				ActionHint: "Install the VS Code command line launcher or configure the VS Code executable.",
			},
		},
		{
			name: "empty executable without error",
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckVSCode,
				Severity:   launcher.ConfidenceBlocked,
				Label:      "VS Code",
				Message:    "Dev Context could not find a VS Code command to launch.",
				ActionHint: "Install the VS Code command line launcher or configure the VS Code executable.",
			},
		},
		{
			name: "unexpected error",
			err:  errors.New("raw path /tmp/code failed"),
			want: launcher.ConfidenceCheck{
				Component:  launcher.ConfidenceCheckVSCode,
				Severity:   launcher.ConfidenceBlocked,
				Label:      "VS Code",
				Message:    "VS Code readiness could not be checked.",
				ActionHint: "Install the VS Code command line launcher or configure the VS Code executable.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := launcher.VSCodeConfidenceCheck(tt.executable, tt.err)
			if got != tt.want {
				t.Fatalf("check = %#v, want %#v", got, tt.want)
			}
			if !got.Valid() {
				t.Fatalf("check is not valid: %#v", got)
			}
		})
	}
}

func TestIsolationConfidenceChecksRepresentStorageReadiness(t *testing.T) {
	root := t.TempDir()
	paths := filesystem.ContextPaths{
		ContextID:              devcontext.MustID("personal"),
		RootDir:                filepath.Join(root, "personal"),
		ProviderStorageRootDir: filepath.Join(root, "personal", "providers"),
		ToolStorageRootDir:     filepath.Join(root, "personal", "tools"),
		ToolStorageDirs:        map[codingtool.ID]string{codingtool.VSCodeID: filepath.Join(root, "personal", "tools", "vscode")},
	}
	paths = paths.WithProviderStorageDirs([]provider.ID{provider.ClaudeID, provider.CodexID})
	for _, dir := range []string{paths.RootDir, paths.ProviderStorageDir(provider.ClaudeID), paths.ProviderStorageDir(provider.CodexID), paths.ToolStorageRootDir, paths.ToolStorageDir(codingtool.VSCodeID)} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create directory %q: %v", dir, err)
		}
	}

	got := launcher.IsolationConfidenceChecks(paths, []provider.Provider{provider.ClaudeProvider{}, provider.CodexProvider{}})
	want := []launcher.ConfidenceCheck{
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "Context storage",
			Message:   "Context storage is ready.",
		},
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "Claude isolation",
			Message:   "Claude isolation storage is ready.",
		},
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "Codex isolation",
			Message:   "Codex isolation storage is ready.",
		},
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "VS Code profile",
			Message:   "VS Code profile isolation is ready.",
		},
	}
	if !equalConfidenceChecks(got, want) {
		t.Fatalf("checks = %#v, want %#v", got, want)
	}
}

func TestIsolationConfidenceChecksReportBlockedStorage(t *testing.T) {
	root := t.TempDir()
	paths := filesystem.ContextPaths{
		ContextID:              devcontext.MustID("personal"),
		RootDir:                filepath.Join(root, "personal"),
		ProviderStorageRootDir: filepath.Join(root, "personal", "providers"),
		ToolStorageRootDir:     filepath.Join(root, "personal", "tools"),
		ToolStorageDirs:        map[codingtool.ID]string{codingtool.VSCodeID: filepath.Join(root, "personal", "tools", "vscode")},
	}
	paths = paths.WithProviderStorageDirs([]provider.ID{provider.ClaudeID, provider.CodexID})
	for _, dir := range []string{paths.RootDir, paths.ProviderStorageDir(provider.ClaudeID), paths.ToolStorageRootDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create directory %q: %v", dir, err)
		}
	}
	if err := os.WriteFile(paths.ToolStorageDir(codingtool.VSCodeID), []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write vscode user data file: %v", err)
	}

	got := launcher.IsolationConfidenceChecks(paths, []provider.Provider{provider.ClaudeProvider{}, provider.CodexProvider{}})
	want := []launcher.ConfidenceCheck{
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "Context storage",
			Message:   "Context storage is ready.",
		},
		{
			Component: launcher.ConfidenceCheckIsolation,
			Severity:  launcher.ConfidenceReady,
			Label:     "Claude isolation",
			Message:   "Claude isolation storage is ready.",
		},
		{
			Component:  launcher.ConfidenceCheckIsolation,
			Severity:   launcher.ConfidenceBlocked,
			Label:      "Codex isolation",
			Message:    "Codex isolation storage is not ready.",
			ActionHint: "Run diagnostics to repair context storage.",
		},
		{
			Component:  launcher.ConfidenceCheckIsolation,
			Severity:   launcher.ConfidenceBlocked,
			Label:      "VS Code profile",
			Message:    "VS Code profile isolation is not ready.",
			ActionHint: "Run diagnostics to repair context storage.",
		},
	}
	if !equalConfidenceChecks(got, want) {
		t.Fatalf("checks = %#v, want %#v", got, want)
	}
	for _, check := range got {
		if !check.Valid() {
			t.Fatalf("check is not valid: %#v", check)
		}
	}
}

func equalConfidenceChecks(a []launcher.ConfidenceCheck, b []launcher.ConfidenceCheck) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
