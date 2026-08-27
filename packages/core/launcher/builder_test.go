package launcher_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestLaunchPlanBuilderBuildsCompletePlan(t *testing.T) {
	projectDir := t.TempDir()
	context := devcontext.Context{
		ID:   devcontext.MustID("client-a"),
		Name: "Client A",
		Tool: codingtool.Config{Type: "fake-editor"},
		Providers: provider.Configs{
			"fake":     {Enabled: true},
			"disabled": {Enabled: false},
			"missing":  {Enabled: true},
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
	warnings := []launcher.ResolutionWarning{
		{
			Code:    "example_warning",
			Message: "example warning",
		},
	}

	devContextHome := filepath.Join(t.TempDir(), ".devctx")
	platformPaths := fakePlanPlatformPaths{
		devContextHome: devContextHome,
	}
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, context.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	createContextDirectories(t, contextPaths)
	contextPaths = contextPaths.WithProviderStorageDirs([]provider.ID{"fake"})
	contextPaths = contextPaths.WithToolStorageDirs([]codingtool.ID{context.Tool.Type})

	fakeEditor := &builderFakeEditor{}
	builder := launcher.LaunchPlanBuilder{
		Resolver: fakePlanResolver{
			result: launcher.ResolutionResult{
				Context:  &context,
				Source:   launcher.ResolutionSourceExplicit,
				Warnings: warnings,
			},
		},
		PlatformPaths: platformPaths,
		ProviderRegistry: provider.MustNewRegistry([]provider.Provider{
			builderFakeProvider{id: "fake"},
			builderFakeProvider{id: "disabled"},
		}),
		ToolRegistry: codingtool.MustNewRegistry([]codingtool.RegisteredTool{{
			Integration: fakeEditor,
			DisplayName: "Fake Tool",
		}}, fakeEditor.ID()),
		ParentEnvironment: []string{
			"PATH=/usr/local/bin",
			"CODEX_HOME=/old/codex",
		},
	}

	request := launcher.LaunchRequest{
		ProjectPath:      project.Path(projectDir),
		RequestedContext: &context.ID,
		Interactive:      false,
		Source:           launcher.InvocationSourceCLI,
	}
	plan, err := builder.Build(request)
	if err != nil {
		t.Fatalf("build launch plan: %v", err)
	}

	want := launcher.LaunchPlan{
		ProjectPath:      project.Path(projectDir),
		Context:          context,
		Tool:             context.Tool,
		Executable:       launcher.Executable("/usr/local/bin/fake-editor"),
		Arguments:        launcher.Arguments{"--state-dir", contextPaths.ToolStorageDir(context.Tool.Type), projectDir},
		WorkingDirectory: launcher.WorkingDirectory(projectDir),
		Environment: launcher.Environment{
			"PATH":           "/usr/local/bin",
			"CODEX_HOME":     "/old/codex",
			"DEVCTX_CONTEXT": "client-a",
			"FAKE_CONTEXT":   "client-a",
			"FAKE_ROOT":      contextPaths.RootDir,
			"FAKE_STORAGE":   contextPaths.ProviderStorageDir("fake"),
		},
		ContextPaths:     contextPaths,
		Warnings:         warnings,
		ResolutionSource: launcher.ResolutionSourceExplicit,
		MissingProviderIDs: []provider.ID{
			"missing",
		},
	}
	if !reflect.DeepEqual(plan, want) {
		t.Fatalf("plan = %#v, want %#v", plan, want)
	}

	wantEditorRequest := codingtool.CommandRequest{
		Config:      context.Tool,
		Executable:  "/usr/local/bin/fake-editor",
		ProjectPath: projectDir,
		Paths: codingtool.ContextPaths{
			RootDir:     contextPaths.RootDir,
			DataDir:     contextPaths.ToolStorageDir(context.Tool.Type),
			UserDataDir: contextPaths.ToolStorageDir(context.Tool.Type),
		},
	}
	if !reflect.DeepEqual(fakeEditor.commandRequests, []codingtool.CommandRequest{wantEditorRequest}) {
		t.Fatalf("editor requests = %#v, want %#v", fakeEditor.commandRequests, []codingtool.CommandRequest{wantEditorRequest})
	}
}

func TestLaunchPlanBuilderDoesNotRequireProviderCLICommands(t *testing.T) {
	projectDir := t.TempDir()
	context := devcontext.DefaultPersonalContext(time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC))
	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	context.Tool.ExecutableOverride = executable
	platformPaths := fakePlanPlatformPaths{
		devContextHome: filepath.Join(t.TempDir(), ".devctx"),
	}
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, context.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	createContextDirectories(t, contextPaths)

	builder := launcher.LaunchPlanBuilder{
		Resolver: fakePlanResolver{
			result: launcher.ResolutionResult{
				Context: &context,
				Source:  launcher.ResolutionSourceExplicit,
			},
		},
		PlatformPaths:     platformPaths,
		ProviderRegistry:  provider.BuiltInRegistry(),
		Tool:              codingtool.VSCodeEditor{},
		ParentEnvironment: []string{"PATH=/path/without/provider-clis"},
	}

	plan, err := builder.Build(launcher.LaunchRequest{
		ProjectPath:      project.Path(projectDir),
		RequestedContext: &context.ID,
		Source:           launcher.InvocationSourceCLI,
	})
	if err != nil {
		t.Fatalf("build launch plan: %v", err)
	}

	if plan.Executable != launcher.Executable(executable) {
		t.Fatalf("executable = %q, want %q", plan.Executable, executable)
	}
	if !reflect.DeepEqual(plan.Arguments, launcher.Arguments{projectDir}) {
		t.Fatalf("arguments = %#v, want only project path", plan.Arguments)
	}
	if plan.Environment[provider.CodexHomeEnvVar] != contextPaths.ProviderStorageDir(provider.CodexID) {
		t.Fatalf("CODEX_HOME = %q, want %q", plan.Environment[provider.CodexHomeEnvVar], contextPaths.ProviderStorageDir(provider.CodexID))
	}
	if plan.Environment[provider.ClaudeConfigDirEnvVar] != contextPaths.ProviderStorageDir(provider.ClaudeID) {
		t.Fatalf("CLAUDE_CONFIG_DIR = %q, want %q", plan.Environment[provider.ClaudeConfigDirEnvVar], contextPaths.ProviderStorageDir(provider.ClaudeID))
	}
}

func createContextDirectories(t *testing.T, paths filesystem.ContextPaths) {
	t.Helper()

	paths = paths.WithProviderStorageDirs([]provider.ID{provider.ClaudeID, provider.CodexID, "fake"})
	for _, dir := range []string{
		paths.RootDir,
		paths.ToolStorageRootDir,
		paths.ToolStorageDir(codingtool.VSCodeID),
		paths.ToolStorageDir("fake-editor"),
	} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create context directory %q: %v", dir, err)
		}
	}
	for _, dir := range paths.ProviderStorageDirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create provider directory %q: %v", dir, err)
		}
	}
}

func TestLaunchPlanBuilderRequiresResolvedContext(t *testing.T) {
	projectDir := t.TempDir()
	builder := launcher.LaunchPlanBuilder{
		Resolver: fakePlanResolver{
			result: launcher.ResolutionResult{
				SelectionRequired: true,
			},
		},
		PlatformPaths: fakePlanPlatformPaths{devContextHome: "/home/alex/.devctx"},
		Tool:          &builderFakeEditor{},
	}

	_, err := builder.Build(launcher.LaunchRequest{
		ProjectPath: project.Path(projectDir),
	})
	if !errors.Is(err, launcher.ErrLaunchSelectionRequired) {
		t.Fatalf("error = %v, want %v", err, launcher.ErrLaunchSelectionRequired)
	}
}

func TestLaunchPlanBuilderRejectsIncompleteContextStorage(t *testing.T) {
	projectDir := t.TempDir()
	context := devcontext.Context{
		ID:        devcontext.MustID("personal"),
		Name:      "Personal",
		Tool:      codingtool.DefaultConfig(),
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
	platformPaths := fakePlanPlatformPaths{
		devContextHome: filepath.Join(t.TempDir(), ".devctx"),
	}
	contextPaths, err := filesystem.DeriveContextPaths(platformPaths, context.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	context.Providers = provider.Configs{
		provider.CodexID: {Enabled: true},
	}
	createContextDirectories(t, contextPaths)
	contextPaths = contextPaths.WithProviderStorageDirs([]provider.ID{provider.CodexID})
	if err := os.RemoveAll(contextPaths.ProviderStorageDir(provider.CodexID)); err != nil {
		t.Fatalf("remove codex dir: %v", err)
	}
	fakeEditor := &builderFakeEditor{}

	builder := launcher.LaunchPlanBuilder{
		Resolver: fakePlanResolver{
			result: launcher.ResolutionResult{
				Context: &context,
				Source:  launcher.ResolutionSourceExplicit,
			},
		},
		PlatformPaths: platformPaths,
		Tool:          fakeEditor,
	}

	_, err = builder.Build(launcher.LaunchRequest{
		ProjectPath:      project.Path(projectDir),
		RequestedContext: &context.ID,
	})
	if !errors.Is(err, filesystem.ErrContextStorageIncomplete) {
		t.Fatalf("error = %v, want %v", err, filesystem.ErrContextStorageIncomplete)
	}
	if len(fakeEditor.commandRequests) != 0 {
		t.Fatalf("editor requests = %#v, want none", fakeEditor.commandRequests)
	}
}

type fakePlanResolver struct {
	result launcher.ResolutionResult
	err    error
}

func (r fakePlanResolver) Resolve(launcher.LaunchRequest) (launcher.ResolutionResult, error) {
	return r.result, r.err
}

type fakePlanPlatformPaths struct {
	devContextHome string
}

func (p fakePlanPlatformPaths) UserHomeDir() (string, error) {
	return "/home/alex", nil
}

func (p fakePlanPlatformPaths) DevContextHomeDir() (string, error) {
	return p.devContextHome, nil
}

func (p fakePlanPlatformPaths) NormalizePath(path string) (string, error) {
	return path, nil
}

type builderFakeProvider struct {
	id provider.ID
}

func (p builderFakeProvider) ID() provider.ID {
	return p.id
}

func (p builderFakeProvider) DisplayName() string {
	return "Fake Provider"
}

func (p builderFakeProvider) BuildEnvironment(ctx provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	return provider.EnvironmentContribution{
		"FAKE_CONTEXT": ctx.ContextID,
		"FAKE_ROOT":    ctx.Paths.RootDir,
		"FAKE_STORAGE": ctx.Paths.StorageDir,
	}, nil
}

func (p builderFakeProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}

type builderFakeEditor struct {
	commandRequests []codingtool.CommandRequest
}

func (e *builderFakeEditor) ID() codingtool.ID {
	return "fake-editor"
}

func (e *builderFakeEditor) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	return "/usr/local/bin/fake-editor", nil
}

func (e *builderFakeEditor) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	e.commandRequests = append(e.commandRequests, request)
	return codingtool.Command{
		Executable: request.Executable,
		Arguments: codingtool.Arguments{
			"--state-dir",
			request.Paths.UserDataDir,
			request.ProjectPath,
		},
	}, nil
}
