package launcher_test

import (
	"errors"
	"reflect"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestLaunchPlanBuilderBuildsCompletePlan(t *testing.T) {
	projectDir := t.TempDir()
	context := devcontext.Context{
		ID:     devcontext.MustID("client-a"),
		Name:   "Client A",
		Editor: editor.DefaultConfig(),
		Providers: provider.Configs{
			"fake":     {Enabled: true},
			"disabled": {Enabled: false},
		},
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
	warnings := []launcher.ResolutionWarning{
		{
			Code:    "example_warning",
			Message: "example warning",
		},
	}

	fakeEditor := &builderFakeEditor{}
	builder := launcher.LaunchPlanBuilder{
		Resolver: fakePlanResolver{
			result: launcher.ResolutionResult{
				Context:  &context,
				Source:   launcher.ResolutionSourceExplicit,
				Warnings: warnings,
			},
		},
		PlatformPaths: fakePlanPlatformPaths{
			devContextHome: "/home/alex/.devctx",
		},
		Providers: []provider.Provider{
			builderFakeProvider{id: "fake"},
			builderFakeProvider{id: "disabled"},
		},
		Editor: fakeEditor,
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
		Editor:           context.Editor,
		Executable:       launcher.Executable("/usr/local/bin/fake-editor"),
		Arguments:        launcher.Arguments{"--state-dir", "/home/alex/.devctx/contexts/client-a/vscode/user-data", projectDir},
		WorkingDirectory: launcher.WorkingDirectory(projectDir),
		Environment: launcher.Environment{
			"PATH":           "/usr/local/bin",
			"CODEX_HOME":     "/old/codex",
			"DEVCTX_CONTEXT": "client-a",
			"FAKE_CONTEXT":   "client-a",
			"FAKE_ROOT":      "/home/alex/.devctx/contexts/client-a",
		},
		ContextPaths: filesystem.ContextPaths{
			ContextID:         devcontext.MustID("client-a"),
			RootDir:           "/home/alex/.devctx/contexts/client-a",
			ConfigPath:        "/home/alex/.devctx/contexts/client-a/context.toml",
			ClaudeDir:         "/home/alex/.devctx/contexts/client-a/claude",
			CodexDir:          "/home/alex/.devctx/contexts/client-a/codex",
			VSCodeDir:         "/home/alex/.devctx/contexts/client-a/vscode",
			VSCodeUserDataDir: "/home/alex/.devctx/contexts/client-a/vscode/user-data",
		},
		Warnings:         warnings,
		ResolutionSource: launcher.ResolutionSourceExplicit,
	}
	if !reflect.DeepEqual(plan, want) {
		t.Fatalf("plan = %#v, want %#v", plan, want)
	}

	wantEditorRequest := editor.CommandRequest{
		Config:      context.Editor,
		Executable:  "/usr/local/bin/fake-editor",
		ProjectPath: projectDir,
		Paths: editor.ContextPaths{
			RootDir:     "/home/alex/.devctx/contexts/client-a",
			DataDir:     "/home/alex/.devctx/contexts/client-a/vscode",
			UserDataDir: "/home/alex/.devctx/contexts/client-a/vscode/user-data",
		},
	}
	if !reflect.DeepEqual(fakeEditor.commandRequests, []editor.CommandRequest{wantEditorRequest}) {
		t.Fatalf("editor requests = %#v, want %#v", fakeEditor.commandRequests, []editor.CommandRequest{wantEditorRequest})
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
		Editor:        &builderFakeEditor{},
	}

	_, err := builder.Build(launcher.LaunchRequest{
		ProjectPath: project.Path(projectDir),
	})
	if !errors.Is(err, launcher.ErrLaunchSelectionRequired) {
		t.Fatalf("error = %v, want %v", err, launcher.ErrLaunchSelectionRequired)
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
	}, nil
}

func (p builderFakeProvider) Status(provider.RuntimeContext) (provider.Status, error) {
	return provider.ReadyStatus(), nil
}

type builderFakeEditor struct {
	commandRequests []editor.CommandRequest
}

func (e *builderFakeEditor) ID() editor.ID {
	return "fake-editor"
}

func (e *builderFakeEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return "/usr/local/bin/fake-editor", nil
}

func (e *builderFakeEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	e.commandRequests = append(e.commandRequests, request)
	return editor.Command{
		Executable: request.Executable,
		Arguments: editor.Arguments{
			"--state-dir",
			request.Paths.UserDataDir,
			request.ProjectPath,
		},
	}, nil
}
