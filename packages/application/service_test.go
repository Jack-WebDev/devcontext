package application

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestNewDefaultServiceOwnsDefaultDependencies(t *testing.T) {
	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	workingDirectory := filepath.Join(root, "project")
	if err := os.MkdirAll(workingDirectory, 0o700); err != nil {
		t.Fatalf("create working directory: %v", err)
	}

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	service, err := NewDefaultService(DefaultOptions{
		Paths:             paths,
		ParentEnvironment: []string{"PATH=/usr/local/bin"},
		WorkingDirectory:  workingDirectory,
		Now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatalf("create default service: %v", err)
	}

	if _, err := os.Stat(filepath.Join(homeDir, ".devctx", "contexts")); err != nil {
		t.Fatalf("default service did not initialize contexts directory: %v", err)
	}
	if service.dependencies.WorkingDirectory != workingDirectory {
		t.Fatalf("working directory = %q, want %q", service.dependencies.WorkingDirectory, workingDirectory)
	}
	if service.dependencies.DetachMode != launcher.DetachModeDetached {
		t.Fatalf("detach mode = %q, want %q", service.dependencies.DetachMode, launcher.DetachModeDetached)
	}
	if !service.now().Equal(now) {
		t.Fatalf("now = %s, want %s", service.now(), now)
	}

	builder := service.launchPlanBuilder()
	if builder.PlatformPaths != paths {
		t.Fatalf("builder paths = %#v, want injected paths", builder.PlatformPaths)
	}
	if !reflect.DeepEqual(builder.ParentEnvironment, []string{"PATH=/usr/local/bin"}) {
		t.Fatalf("parent environment = %#v, want injected environment", builder.ParentEnvironment)
	}
	if service.processLauncher() == nil {
		t.Fatal("process launcher is nil")
	}
}

func TestFirstRunInitializationLifecycleIsIdempotent(t *testing.T) {
	root := t.TempDir()
	homeDir := filepath.Join(root, "home")
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(projectDir, 0o700); err != nil {
		t.Fatalf("create project directory: %v", err)
	}

	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return homeDir, nil
	})
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)
	options := DefaultOptions{
		Paths:             paths,
		ParentEnvironment: []string{"PATH=/usr/local/bin"},
		WorkingDirectory:  projectDir,
		Now: func() time.Time {
			return now
		},
	}

	firstService, err := NewDefaultService(options)
	if err != nil {
		t.Fatalf("create default service: %v", err)
	}
	firstState, appErr := firstService.GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get first launch state: %v", appErr)
	}
	if !firstState.FirstRun || len(firstState.Contexts) != 0 {
		t.Fatalf("first launch state = %#v, want first-run with no contexts", firstState)
	}

	if _, appErr := firstService.CreateContext(CreateContextRequest{ContextID: "personal"}); appErr != nil {
		t.Fatalf("create personal context: %v", appErr)
	}
	if _, appErr := firstService.CreateContext(CreateContextRequest{ContextID: "company"}); appErr != nil {
		t.Fatalf("create company context: %v", appErr)
	}

	restartedService, err := NewDefaultService(options)
	if err != nil {
		t.Fatalf("restart default service: %v", err)
	}
	restartedState, appErr := restartedService.GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get restarted launch state: %v", appErr)
	}
	if restartedState.FirstRun {
		t.Fatal("restarted first run = true, want false")
	}
	if gotIDs := contextStateIDs(restartedState.Contexts); !reflect.DeepEqual(gotIDs, []string{"company", "personal"}) {
		t.Fatalf("context IDs = %#v, want company and personal", gotIDs)
	}

	layout, err := config.DevContextHomeLayout(paths)
	if err != nil {
		t.Fatalf("derive home layout: %v", err)
	}
	data, err := os.ReadFile(layout.ConfigPath)
	if err != nil {
		t.Fatalf("read default config: %v", err)
	}
	globalConfig, err := config.DecodeGlobalConfigTOML(data)
	if err != nil {
		t.Fatalf("decode default config: %v", err)
	}
	if globalConfig != config.DefaultGlobalConfig() {
		t.Fatalf("global config = %#v, want default", globalConfig)
	}
}

func TestNewServiceWithDependenciesUsesFakesWithoutWails(t *testing.T) {
	root := t.TempDir()
	paths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return filepath.Join(root, "home"), nil
	})
	contexts := devcontext.NewRepository(filepath.Join(root, "contexts"))
	projects := project.NewRepository(filepath.Join(root, "projects.toml"), paths)
	fakeProvider := applicationFakeProvider{id: "fake"}
	fakeEditor := &applicationFakeEditor{}
	fakeLauncher := &applicationFakeProcessLauncher{}
	now := time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC)

	service := NewServiceWithDependencies(Dependencies{
		Contexts:          contexts,
		Projects:          projects,
		Paths:             paths,
		ProviderRegistry:  provider.MustNewRegistry([]provider.Provider{fakeProvider}),
		ToolRegistry:      codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: fakeEditor, DisplayName: "Fake Tool"}}, fakeEditor.ID()),
		ProcessLauncher:   fakeLauncher,
		ParentEnvironment: []string{"PATH=/fixture/bin"},
		WorkingDirectory:  filepath.Join(root, "project"),
		DetachMode:        launcher.DetachModeAttached,
		Now: func() time.Time {
			return now
		},
	})

	builder := service.launchPlanBuilder()
	registeredEditor, ok := builder.ToolRegistry.Get(fakeEditor.ID())
	if !ok || registeredEditor != fakeEditor {
		t.Fatalf("builder tool = %#v, found = %t, want fake editor", registeredEditor, ok)
	}
	if !reflect.DeepEqual(builder.ProviderRegistry.All(), []provider.Provider{fakeProvider}) {
		t.Fatalf("builder providers = %#v, want fake provider", builder.ProviderRegistry.All())
	}
	if service.processLauncher() != fakeLauncher {
		t.Fatalf("process launcher = %#v, want fake launcher", service.processLauncher())
	}
	if service.dependencies.DetachMode != launcher.DetachModeAttached {
		t.Fatalf("detach mode = %q, want %q", service.dependencies.DetachMode, launcher.DetachModeAttached)
	}
	if !service.now().Equal(now) {
		t.Fatalf("now = %s, want %s", service.now(), now)
	}
}

func TestNewErrorReturnsPresentationSafeTypedErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want ErrorCode
	}{
		{
			name: "canceled",
			err:  launcher.ErrContextMismatchRejected,
			want: ErrorCodeCanceled,
		},
		{
			name: "validation",
			err:  project.ErrProjectDirectoryNotFound,
			want: ErrorCodeValidation,
		},
		{
			name: "launch",
			err:  codingtool.ErrExecutableNotFound,
			want: ErrorCodeLaunch,
		},
		{
			name: "internal",
			err:  errors.New("unexpected"),
			want: ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.err)
			if got.Code != tt.want {
				t.Fatalf("code = %q, want %q", got.Code, tt.want)
			}
			if got.Message == "" || got.Recovery == "" {
				t.Fatalf("message and recovery must be presentation-safe: %#v", got)
			}
			if !errors.Is(got, tt.err) {
				t.Fatalf("wrapped error = %v, want %v", got, tt.err)
			}
		})
	}

	if got := NewError(nil); got != nil {
		t.Fatalf("nil error = %#v, want nil", got)
	}
}

func TestNewErrorReturnsActionableRecoveryDetails(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCode     ErrorCode
		wantMessage  string
		wantRecovery []string
	}{
		{
			name: "project path",
			err: &project.PathError{
				Path: "/missing/project",
				Err:  project.ErrProjectDirectoryNotFound,
			},
			wantCode:     ErrorCodeValidation,
			wantMessage:  "Project path does not exist.",
			wantRecovery: []string{"/missing/project", "existing project directory"},
		},
		{
			name: "missing context",
			err: &devcontext.MissingContextError{
				ContextID: devcontext.MustID("client-a"),
				AvailableIDs: []devcontext.ID{
					devcontext.MustID("company"),
					devcontext.MustID("personal"),
				},
			},
			wantCode:     ErrorCodeValidation,
			wantMessage:  `Context "client-a" does not exist.`,
			wantRecovery: []string{"company", "personal", "will not launch under a different context"},
		},
		{
			name: "missing vscode command",
			err: &codingtool.ExecutableNotFoundError{
				EditorID:   codingtool.VSCodeID,
				Candidates: []string{"code"},
			},
			wantCode:     ErrorCodeLaunch,
			wantMessage:  "VS Code command not found.",
			wantRecovery: []string{"VS Code command line launcher", "code", "codingtool.executable_override"},
		},
		{
			name: "storage permission",
			err: &filesystem.StoragePermissionError{
				Operation: "write file",
				Path:      "/home/alex/.devctx/projects.toml",
				Err:       os.ErrPermission,
			},
			wantCode:     ErrorCodeValidation,
			wantMessage:  "Unable to access local storage.",
			wantRecovery: []string{"write file", "/home/alex/.devctx/projects.toml", "permissions"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.err)
			if got.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", got.Code, tt.wantCode)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("message = %q, want %q", got.Message, tt.wantMessage)
			}
			for _, want := range tt.wantRecovery {
				if !strings.Contains(got.Recovery, want) {
					t.Fatalf("recovery = %q, want containing %q", got.Recovery, want)
				}
			}
		})
	}
}

type applicationFakeProvider struct {
	id               provider.ID
	displayName      string
	statusByContext  map[string]provider.Status
	statusErr        error
	environment      provider.EnvironmentContribution
	globalSession    provider.CredentialSession
	hasGlobalSession bool
	identity         provider.Identity
	hasIdentity      bool
}

func contextStateIDs(contexts []ContextState) []string {
	ids := make([]string, len(contexts))
	for i, ctx := range contexts {
		ids[i] = ctx.ID
	}
	return ids
}

func (p applicationFakeProvider) ID() provider.ID {
	return p.id
}

func (p applicationFakeProvider) DisplayName() string {
	if p.displayName != "" {
		return p.displayName
	}
	return "Fake Provider"
}

func (p applicationFakeProvider) BuildEnvironment(ctx provider.RuntimeContext) (provider.EnvironmentContribution, error) {
	if p.environment != nil {
		return p.environment, nil
	}
	return provider.EnvironmentContribution{"FAKE_CONTEXT": ctx.ContextID}, nil
}

func (p applicationFakeProvider) Status(ctx provider.RuntimeContext) (provider.Status, error) {
	if p.statusErr != nil {
		return provider.Status{}, p.statusErr
	}
	if status, ok := p.statusByContext[ctx.ContextID]; ok {
		return status, nil
	}
	return provider.ReadyStatus(), nil
}

func (p applicationFakeProvider) DetectContextIdentity(ctx provider.RuntimeContext) (provider.Identity, bool, error) {
	if p.hasIdentity {
		return p.identity, true, nil
	}
	switch p.id {
	case provider.CodexID:
		return provider.CodexProvider{}.DetectContextIdentity(ctx)
	case provider.ClaudeID:
		return provider.ClaudeProvider{}.DetectContextIdentity(ctx)
	default:
		return provider.Identity{}, false, nil
	}
}

func (p applicationFakeProvider) DetectGlobalCredentialSession(provider.GlobalCredentialContext) (provider.CredentialSession, bool, error) {
	return p.globalSession, p.hasGlobalSession, nil
}

type applicationFakeEditor struct {
	executable codingtool.Executable
	err        error
	requests   []codingtool.CommandRequest
}

func (e *applicationFakeEditor) ID() codingtool.ID {
	return "fake-editor"
}

func (e *applicationFakeEditor) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	if e.err != nil {
		return "", e.err
	}
	if e.executable != "" {
		return e.executable, nil
	}
	return "/fixture/editor", nil
}

func (e *applicationFakeEditor) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	e.requests = append(e.requests, request)
	return codingtool.Command{
		Executable: request.Executable,
		Arguments:  codingtool.Arguments{request.ProjectPath},
	}, nil
}

type applicationFakeProcessLauncher struct {
	requests []launcher.ProcessRequest
	err      error
}

func (l *applicationFakeProcessLauncher) Launch(request launcher.ProcessRequest) error {
	l.requests = append(l.requests, request)
	if l.err != nil {
		return l.err
	}
	return nil
}
