package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"devctx/packages/core/config"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
)

type fakePlatformPaths struct {
	devContextHome string
}

func (f fakePlatformPaths) UserHomeDir() (string, error) {
	return filepath.Dir(f.devContextHome), nil
}

func (f fakePlatformPaths) DevContextHomeDir() (string, error) {
	return f.devContextHome, nil
}

func (f fakePlatformPaths) NormalizePath(path string) (string, error) {
	return path, nil
}

func TestWriteGlobalConfigFileAtomicallyReplacesConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("previous config"), 0o600); err != nil {
		t.Fatalf("write previous config: %v", err)
	}

	globalConfig := config.GlobalConfig{
		Version:       config.CurrentSchemaVersion,
		DefaultEditor: editor.TypeVSCode,
		UI: config.UISettings{
			RememberWindowPosition: false,
		},
		Safety: config.SafetySettings{
			WarnOnContextMismatch:  true,
			ConfirmUnboundProjects: false,
		},
	}

	if err := config.WriteGlobalConfigFile(path, globalConfig); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read global config: %v", err)
	}
	decoded, err := config.DecodeGlobalConfigTOML(data)
	if err != nil {
		t.Fatalf("decode global config: %v", err)
	}
	if decoded != globalConfig {
		t.Fatalf("decoded config = %#v, want %#v", decoded, globalConfig)
	}
	assertPermission(t, path, filesystem.RestrictedFileMode)
}

func TestInitializeDevContextHomeCreatesLayoutIdempotently(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), ".devctx")
	paths := fakePlatformPaths{devContextHome: homeDir}

	firstLayout, err := config.InitializeDevContextHome(paths)
	if err != nil {
		t.Fatalf("initialize home: %v", err)
	}
	secondLayout, err := config.InitializeDevContextHome(paths)
	if err != nil {
		t.Fatalf("initialize home again: %v", err)
	}

	if firstLayout != secondLayout {
		t.Fatalf("second layout = %#v, want %#v", secondLayout, firstLayout)
	}

	assertDirExists(t, firstLayout.HomeDir)
	assertDirExists(t, firstLayout.ContextsDir)
	assertDirExists(t, firstLayout.LogsDir)
	assertPermission(t, firstLayout.HomeDir, filesystem.RestrictedDirectoryMode)
	assertPermission(t, firstLayout.ContextsDir, filesystem.RestrictedDirectoryMode)
	assertPermission(t, firstLayout.LogsDir, filesystem.RestrictedDirectoryMode)

	contextEntries, err := os.ReadDir(firstLayout.ContextsDir)
	if err != nil {
		t.Fatalf("read contexts dir: %v", err)
	}
	if len(contextEntries) != 0 {
		t.Fatalf("contexts dir entry count = %d, want 0", len(contextEntries))
	}

	data, err := os.ReadFile(firstLayout.ConfigPath)
	if err != nil {
		t.Fatalf("read default config: %v", err)
	}
	decoded, err := config.DecodeGlobalConfigTOML(data)
	if err != nil {
		t.Fatalf("decode default config: %v", err)
	}
	if decoded != config.DefaultGlobalConfig() {
		t.Fatalf("decoded default config = %#v, want %#v", decoded, config.DefaultGlobalConfig())
	}
	assertPermission(t, firstLayout.ConfigPath, filesystem.RestrictedFileMode)
}

func TestInitializeDevContextHomePreservesExistingGlobalConfig(t *testing.T) {
	homeDir := filepath.Join(t.TempDir(), ".devctx")
	paths := fakePlatformPaths{devContextHome: homeDir}

	layout, err := config.InitializeDevContextHome(paths)
	if err != nil {
		t.Fatalf("initialize home: %v", err)
	}

	customConfig := config.GlobalConfig{
		Version:       config.CurrentSchemaVersion,
		DefaultEditor: editor.TypeVSCode,
		UI: config.UISettings{
			RememberWindowPosition: false,
		},
		Safety: config.SafetySettings{
			WarnOnContextMismatch:  false,
			ConfirmUnboundProjects: false,
		},
	}
	if err := config.WriteGlobalConfigFile(layout.ConfigPath, customConfig); err != nil {
		t.Fatalf("write custom config: %v", err)
	}

	if _, err := config.InitializeDevContextHome(paths); err != nil {
		t.Fatalf("initialize home again: %v", err)
	}

	data, err := os.ReadFile(layout.ConfigPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	decoded, err := config.DecodeGlobalConfigTOML(data)
	if err != nil {
		t.Fatalf("decode config: %v", err)
	}
	if decoded != customConfig {
		t.Fatalf("decoded config = %#v, want %#v", decoded, customConfig)
	}
}

func assertDirExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

func assertPermission(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %v, want %v", path, got, want)
	}
}
