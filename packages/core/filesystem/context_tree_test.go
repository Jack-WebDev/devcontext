package filesystem_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/provider"
)

func TestCreateContextDirectoryTreeCreatesCompleteRestrictedTree(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")

	if err := filesystem.CreateContextDirectoryTree(contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	for _, dir := range []string{
		contextPaths.RootDir,
		contextPaths.ToolStorageRootDir,
		contextPaths.ToolStorageDir(codingtool.VSCodeID),
		contextPaths.ProviderStorageDir(provider.ClaudeID),
		contextPaths.ProviderStorageDir(provider.CodexID),
	} {
		assertDirectoryExists(t, dir)
		assertRestrictedMode(t, dir, filesystem.RestrictedDirectoryMode)
	}

	data, err := os.ReadFile(contextPaths.ConfigPath)
	if err != nil {
		t.Fatalf("read context config: %v", err)
	}
	decoded, err := devcontext.DecodeContextTOML(data, contextID)
	if err != nil {
		t.Fatalf("decode context config: %v", err)
	}
	if !reflect.DeepEqual(decoded, ctx) {
		t.Fatalf("decoded context = %#v, want %#v", decoded, ctx)
	}
	assertRestrictedMode(t, contextPaths.ConfigPath, filesystem.RestrictedFileMode)
	assertDirectoryEmpty(t, contextPaths.ProviderStorageDir(provider.ClaudeID))
	assertDirectoryEmpty(t, contextPaths.ProviderStorageDir(provider.CodexID))
	assertDirectoryExists(t, contextPaths.ToolStorageRootDir)
	assertDirectoryEmpty(t, contextPaths.ToolStorageDir(codingtool.VSCodeID))
}

func TestCreateContextDirectoryTreeUsesRegisteredEnabledProviders(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")
	ctx.Providers = provider.Configs{
		"registered": {Enabled: true},
		"disabled":   {Enabled: false},
		"unknown":    {Enabled: true},
	}
	registry := provider.MustNewRegistry([]provider.Provider{
		contextTreeFakeProvider{id: "registered", displayName: "Registered Provider"},
		contextTreeFakeProvider{id: "disabled", displayName: "Disabled Provider"},
	})

	if err := filesystem.CreateContextDirectoryTreeWithProviderRegistry(contextPaths, ctx, registry); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertDirectoryExists(t, contextPaths.RootDir)
	assertDirectoryExists(t, contextPaths.ToolStorageRootDir)
	assertDirectoryExists(t, contextPaths.ToolStorageDir(codingtool.VSCodeID))
	assertDirectoryExists(t, contextPaths.ProviderStorageDir("registered"))
	assertPathMissing(t, contextPaths.ProviderStorageDir("disabled"))
	assertPathMissing(t, contextPaths.ProviderStorageDir("unknown"))
}

func TestContextDirectoryTreeUsesRegisteredSelectedTool(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")
	ctx.Tool.DefaultTool = "future-tool"
	ctx.Tool.Tools = map[codingtool.ID]codingtool.Config{"future-tool": {}}
	toolRegistry := codingtool.MustNewRegistry([]codingtool.RegisteredTool{
		{Integration: contextTreeFakeTool{id: "other-tool"}, DisplayName: "Other Tool"},
		{Integration: contextTreeFakeTool{id: "future-tool"}, DisplayName: "Future Tool"},
	}, "future-tool")

	if err := filesystem.CreateContextDirectoryTreeWithRegistries(contextPaths, ctx, provider.BuiltInRegistry(), toolRegistry); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}

	assertDirectoryExists(t, contextPaths.ToolStorageDir("future-tool"))
	assertPathMissing(t, contextPaths.ToolStorageDir("other-tool"))
	assertPathMissing(t, contextPaths.ToolStorageDir(codingtool.VSCodeID))

	if err := os.RemoveAll(contextPaths.ToolStorageDir("future-tool")); err != nil {
		t.Fatalf("remove future tool storage: %v", err)
	}
	err = filesystem.ValidateContextDirectoryTreeWithRegistries(contextPaths, ctx, provider.BuiltInRegistry(), toolRegistry)
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	var storageErr *filesystem.ContextStorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("error = %T, want *filesystem.ContextStorageError", err)
	}
	wantMissing := []filesystem.MissingContextDirectory{
		{
			Kind:   filesystem.ContextDirectoryTool,
			Path:   contextPaths.ToolStorageDir("future-tool"),
			Reason: "missing",
		},
	}
	if !reflect.DeepEqual(storageErr.Missing, wantMissing) {
		t.Fatalf("missing directories = %#v, want %#v", storageErr.Missing, wantMissing)
	}
}

func TestCreateContextDirectoryTreeRejectsMismatchedContextID(t *testing.T) {
	rootDir := "/devctx/contexts/company"
	toolStorageRootDir := filepath.Join(rootDir, "tools")
	contextPaths := filesystem.ContextPaths{
		ContextID:              devcontext.MustID("company"),
		RootDir:                rootDir,
		ConfigPath:             filepath.Join(rootDir, "context.toml"),
		ProviderStorageRootDir: filepath.Join(rootDir, "providers"),
		ToolStorageRootDir:     toolStorageRootDir,
		ToolStorageDirs:        map[codingtool.ID]string{codingtool.VSCodeID: filepath.Join(toolStorageRootDir, string(codingtool.VSCodeID))},
	}
	ctx := contextTreeContext(devcontext.MustID("personal"), "Personal")

	err := filesystem.CreateContextDirectoryTreeWithPermissions(contextPaths, ctx, filesystem.NewStoragePermissions(false, nil))
	if !errors.Is(err, devcontext.ErrContextIDMismatch) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextIDMismatch)
	}
}

func TestCreateContextDirectoryTreeRejectsDuplicateWithoutModifyingTree(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	original := contextTreeContext(contextID, "Client A")
	replacement := contextTreeContext(contextID, "Replacement")

	if err := filesystem.CreateContextDirectoryTree(contextPaths, original); err != nil {
		t.Fatalf("create original context directory tree: %v", err)
	}
	sentinelPath := filepath.Join(contextPaths.ProviderStorageDir(provider.ClaudeID), "auth-state.json")
	if err := os.WriteFile(sentinelPath, []byte("keep me"), 0o600); err != nil {
		t.Fatalf("write sentinel file: %v", err)
	}
	before := snapshotTree(t, contextPaths.RootDir)

	err = filesystem.CreateContextDirectoryTree(contextPaths, replacement)
	if !errors.Is(err, devcontext.ErrContextAlreadyExists) {
		t.Fatalf("error = %v, want %v", err, devcontext.ErrContextAlreadyExists)
	}

	after := snapshotTree(t, contextPaths.RootDir)
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("tree after duplicate create = %#v, want %#v", after, before)
	}

	data, err := os.ReadFile(contextPaths.ConfigPath)
	if err != nil {
		t.Fatalf("read context config: %v", err)
	}
	decoded, err := devcontext.DecodeContextTOML(data, contextID)
	if err != nil {
		t.Fatalf("decode context config: %v", err)
	}
	if !reflect.DeepEqual(decoded, original) {
		t.Fatalf("decoded context = %#v, want %#v", decoded, original)
	}
}

func TestValidateContextDirectoryTreeReportsIncompleteStorage(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")
	if err := filesystem.CreateContextDirectoryTree(contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}
	contextPaths = contextPaths.WithProviderStorageDirs([]provider.ID{provider.ClaudeID, provider.CodexID})
	if err := os.RemoveAll(contextPaths.ProviderStorageDir(provider.CodexID)); err != nil {
		t.Fatalf("remove codex dir: %v", err)
	}

	err = filesystem.ValidateContextDirectoryTreeWithProviderRegistry(contextPaths, ctx, provider.BuiltInRegistry())
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	var storageErr *filesystem.ContextStorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("error = %T, want *filesystem.ContextStorageError", err)
	}
	wantMissing := []filesystem.MissingContextDirectory{
		{
			Kind:                filesystem.ContextDirectoryProvider,
			ProviderID:          string(provider.CodexID),
			ProviderDisplayName: "Codex",
			Path:                contextPaths.ProviderStorageDir(provider.CodexID),
			Reason:              "missing",
		},
	}
	if !reflect.DeepEqual(storageErr.Missing, wantMissing) {
		t.Fatalf("missing directories = %#v, want %#v", storageErr.Missing, wantMissing)
	}
}

func TestValidateContextDirectoryTreeChecksToolStorage(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")
	if err := filesystem.CreateContextDirectoryTree(contextPaths, ctx); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}
	if err := os.RemoveAll(contextPaths.ToolStorageDir(codingtool.VSCodeID)); err != nil {
		t.Fatalf("remove selected tool storage: %v", err)
	}

	err = filesystem.ValidateContextDirectoryTreeWithProviderRegistry(contextPaths, ctx, provider.BuiltInRegistry())
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	var storageErr *filesystem.ContextStorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("error = %T, want *filesystem.ContextStorageError", err)
	}
	wantMissing := []filesystem.MissingContextDirectory{
		{
			Kind:   filesystem.ContextDirectoryTool,
			Path:   contextPaths.ToolStorageDir(codingtool.VSCodeID),
			Reason: "missing",
		},
	}
	if !reflect.DeepEqual(storageErr.Missing, wantMissing) {
		t.Fatalf("missing directories = %#v, want %#v", storageErr.Missing, wantMissing)
	}
}

func TestValidateContextDirectoryTreeUsesRegisteredEnabledProviders(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return t.TempDir(), nil
	})
	contextID := devcontext.MustID("client-a")
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, contextID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	ctx := contextTreeContext(contextID, "Client A")
	ctx.Providers = provider.Configs{
		"registered": {Enabled: true},
		"disabled":   {Enabled: false},
		"unknown":    {Enabled: true},
	}
	registry := provider.MustNewRegistry([]provider.Provider{
		contextTreeFakeProvider{id: "registered", displayName: "Registered Provider"},
		contextTreeFakeProvider{id: "disabled", displayName: "Disabled Provider"},
	})
	if err := filesystem.CreateContextDirectoryTreeWithProviderRegistry(contextPaths, ctx, registry); err != nil {
		t.Fatalf("create context directory tree: %v", err)
	}
	if err := os.RemoveAll(contextPaths.ProviderStorageDir("registered")); err != nil {
		t.Fatalf("remove registered provider dir: %v", err)
	}

	err = filesystem.ValidateContextDirectoryTreeWithProviderRegistry(contextPaths, ctx, registry)
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	var storageErr *filesystem.ContextStorageError
	if !errors.As(err, &storageErr) {
		t.Fatalf("error = %T, want *filesystem.ContextStorageError", err)
	}
	wantMissing := []filesystem.MissingContextDirectory{
		{
			Kind:                filesystem.ContextDirectoryProvider,
			ProviderID:          "registered",
			ProviderDisplayName: "Registered Provider",
			Path:                contextPaths.ProviderStorageDir("registered"),
			Reason:              "missing",
		},
	}
	if !reflect.DeepEqual(storageErr.Missing, wantMissing) {
		t.Fatalf("missing directories = %#v, want %#v", storageErr.Missing, wantMissing)
	}
}

func TestBootstrapDefaultContextsCreateLoadableSeeds(t *testing.T) {
	homeDir := t.TempDir()
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	devContextHomeDir, err := platformPaths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	createdAt := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name      string
		bootstrap func(filesystem.PlatformPaths, time.Time) (devcontext.Context, error)
		want      devcontext.Context
	}{
		{
			name:      "personal",
			bootstrap: filesystem.BootstrapPersonalContext,
			want:      devcontext.DefaultPersonalContext(createdAt),
		},
		{
			name:      "company",
			bootstrap: filesystem.BootstrapCompanyContext,
			want:      devcontext.DefaultCompanyContext(createdAt),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seeded, err := tt.bootstrap(platformPaths, createdAt)
			if err != nil {
				t.Fatalf("bootstrap context: %v", err)
			}
			if !reflect.DeepEqual(seeded, tt.want) {
				t.Fatalf("seeded context = %#v, want %#v", seeded, tt.want)
			}

			repository := devcontext.NewRepository(filepath.Join(devContextHomeDir, "contexts"))
			stored, err := repository.Get(tt.want.ID)
			if err != nil {
				t.Fatalf("get bootstrapped context: %v", err)
			}
			if !reflect.DeepEqual(stored, tt.want) {
				t.Fatalf("stored context = %#v, want %#v", stored, tt.want)
			}
		})
	}
}

func assertDirectoryExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat directory %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", path)
	}
}

func assertRestrictedMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()

	if runtime.GOOS == "windows" {
		return
	}
	assertMode(t, path, want)
}

func assertDirectoryEmpty(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("read directory %s: %v", path, err)
	}
	if len(entries) != 0 {
		t.Fatalf("directory %s entry count = %d, want 0", path, len(entries))
	}
}

func assertPathMissing(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Stat(path); err == nil {
		t.Fatalf("path %s exists, want missing", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat path %s: %v", path, err)
	}
}

func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()

	snapshot := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			snapshot[rel] = "dir"
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		snapshot[rel] = "file:" + string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot tree %s: %v", root, err)
	}
	return snapshot
}

func contextTreeContext(id devcontext.ID, name string) devcontext.Context {
	return devcontext.Context{
		ID:   id,
		Name: name,
		Tool: codingtool.DefaultLaunchTarget(),
		Providers: provider.Configs{
			"claude": {Enabled: true},
			"codex":  {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"kind": "test",
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}

type contextTreeFakeProvider struct {
	id          provider.ID
	displayName string
}

type contextTreeFakeTool struct {
	id codingtool.ID
}

func (t contextTreeFakeTool) ID() codingtool.ID { return t.id }

func (contextTreeFakeTool) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	return "fake-tool", nil
}

func (contextTreeFakeTool) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	return codingtool.Command{Executable: request.Executable}, nil
}

func (p contextTreeFakeProvider) ID() provider.ID {
	return p.id
}

func (p contextTreeFakeProvider) DisplayName() string {
	if p.displayName != "" {
		return p.displayName
	}
	return string(p.id)
}

func (p contextTreeFakeProvider) BuildEnvironment(provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return nil, nil
}

func (p contextTreeFakeProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}
