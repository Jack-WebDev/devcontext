package provider_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devctx/packages/core/provider"
)

func TestCodexProviderDetectsLocalStatus(t *testing.T) {
	paths := localStatusFixture(t)
	tests := []struct {
		name          string
		toolAvailable bool
		directory     string
		wantState     provider.StatusState
	}{
		{
			name:          "missing tool",
			toolAvailable: false,
			directory:     paths.configuredDir,
			wantState:     provider.StatusUnavailable,
		},
		{
			name:          "missing directory",
			toolAvailable: true,
			directory:     paths.missingDir,
			wantState:     provider.StatusDirectoryMissing,
		},
		{
			name:          "empty directory",
			toolAvailable: true,
			directory:     paths.emptyDir,
			wantState:     provider.StatusNotConfigured,
		},
		{
			name:          "configured fixture",
			toolAvailable: true,
			directory:     paths.configuredDir,
			wantState:     provider.StatusReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := provider.CodexProvider{
				Probe: testStatusProbe{toolAvailable: tt.toolAvailable},
			}
			status, err := integration.Status(provider.RuntimeContext{
				Paths: provider.ContextPaths{
					CodexDir: tt.directory,
				},
			})
			if err != nil {
				t.Fatalf("status: %v", err)
			}

			assertStatusState(t, status, tt.wantState)
			assertStatusDoesNotExposeSecret(t, status, paths.rawSecret)
		})
	}
}

func TestClaudeProviderDetectsLocalStatus(t *testing.T) {
	paths := localStatusFixture(t)
	tests := []struct {
		name          string
		toolAvailable bool
		directory     string
		wantState     provider.StatusState
	}{
		{
			name:          "missing tool",
			toolAvailable: false,
			directory:     paths.configuredDir,
			wantState:     provider.StatusUnavailable,
		},
		{
			name:          "missing directory",
			toolAvailable: true,
			directory:     paths.missingDir,
			wantState:     provider.StatusDirectoryMissing,
		},
		{
			name:          "empty directory",
			toolAvailable: true,
			directory:     paths.emptyDir,
			wantState:     provider.StatusNotConfigured,
		},
		{
			name:          "configured fixture",
			toolAvailable: true,
			directory:     paths.configuredDir,
			wantState:     provider.StatusReady,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integration := provider.ClaudeProvider{
				Probe: testStatusProbe{toolAvailable: tt.toolAvailable},
			}
			status, err := integration.Status(provider.RuntimeContext{
				Paths: provider.ContextPaths{
					ClaudeDir: tt.directory,
				},
			})
			if err != nil {
				t.Fatalf("status: %v", err)
			}

			assertStatusState(t, status, tt.wantState)
			assertStatusDoesNotExposeSecret(t, status, paths.rawSecret)
		})
	}
}

type statusFixturePaths struct {
	emptyDir      string
	configuredDir string
	missingDir    string
	rawSecret     string
}

func localStatusFixture(t *testing.T) statusFixturePaths {
	t.Helper()

	root := t.TempDir()
	emptyDir := filepath.Join(root, "empty")
	configuredDir := filepath.Join(root, "configured")
	missingDir := filepath.Join(root, "missing")
	rawSecret := "raw-provider-token"

	if err := os.Mkdir(emptyDir, 0o700); err != nil {
		t.Fatalf("create empty directory: %v", err)
	}
	if err := os.Mkdir(configuredDir, 0o700); err != nil {
		t.Fatalf("create configured directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(configuredDir, "credentials.json"), []byte(rawSecret), 0o600); err != nil {
		t.Fatalf("write configured fixture: %v", err)
	}

	return statusFixturePaths{
		emptyDir:      emptyDir,
		configuredDir: configuredDir,
		missingDir:    missingDir,
		rawSecret:     rawSecret,
	}
}

type testStatusProbe struct {
	toolAvailable bool
}

func (p testStatusProbe) LookPath(file string) (string, error) {
	if !p.toolAvailable {
		return "", errors.New("tool not found")
	}
	return filepath.Join("/usr/local/bin", file), nil
}

func (testStatusProbe) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func assertStatusState(t *testing.T, status provider.Status, want provider.StatusState) {
	t.Helper()

	if status.State != want {
		t.Fatalf("status state = %q, want %q", status.State, want)
	}
}

func assertStatusDoesNotExposeSecret(t *testing.T, status provider.Status, rawSecret string) {
	t.Helper()

	if strings.Contains(status.Explanation, rawSecret) {
		t.Fatalf("status explanation exposes secret %q", rawSecret)
	}
}
