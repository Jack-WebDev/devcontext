package application

import (
	"os"
	"path/filepath"
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

func TestGetLaunchStateReturnsBoundProjectState(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if state.Project != (ProjectState{Name: "current", Path: fixture.projectDir}) {
		t.Fatalf("project = %#v, want current project", state.Project)
	}
	if !state.Binding.Bound || state.Binding.ContextID != "personal" {
		t.Fatalf("binding = %#v, want personal binding", state.Binding)
	}
	if state.SelectedContextID != "personal" {
		t.Fatalf("selected context = %q, want personal", state.SelectedContextID)
	}
	if state.SelectionRequired {
		t.Fatal("selection required = true, want false")
	}
	if state.FirstRun {
		t.Fatal("first run = true, want false")
	}
	if state.ResolutionSource != string(launcher.ResolutionSourceProjectBinding) {
		t.Fatalf("resolution source = %q, want project binding", state.ResolutionSource)
	}
	assertContextStates(t, state.Contexts, []ContextState{
		{
			ID:     "personal",
			Name:   "Personal",
			Editor: EditorState{Type: string(editor.TypeVSCode)},
			Providers: []ProviderState{
				{ID: "fake", Name: "Fake Provider", Enabled: true, State: string(provider.StatusReady)},
			},
			Metadata: map[string]string{"accent": "blue"},
		},
	})
}

func TestGetLaunchStateReturnsUnboundSelectorState(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if state.Binding.Bound {
		t.Fatalf("binding = %#v, want unbound", state.Binding)
	}
	if !state.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
	if state.SelectedContextID != "" {
		t.Fatalf("selected context = %q, want empty", state.SelectedContextID)
	}
	if len(state.Contexts) != 2 {
		t.Fatalf("context count = %d, want 2", len(state.Contexts))
	}
	if state.FirstRun {
		t.Fatal("first run = true, want false")
	}
}

func TestGetLaunchStateReportsMissingProviderStatus(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.provider.statusByContext = map[string]provider.Status{
		"personal": provider.UnavailableStatus("Fake Provider command was not found"),
	}
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	got := state.Contexts[0].Providers[0]
	want := ProviderState{
		ID:          "fake",
		Name:        "Fake Provider",
		Enabled:     true,
		State:       string(provider.StatusUnavailable),
		Explanation: "Fake Provider command was not found",
	}
	if got != want {
		t.Fatalf("provider status = %#v, want %#v", got, want)
	}
}

func TestGetLaunchStateReportsDanglingBindingWarning(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if !state.Binding.Dangling || state.Binding.MissingContextID != "company" {
		t.Fatalf("binding = %#v, want dangling company binding", state.Binding)
	}
	if len(state.Warnings) != 1 || state.Warnings[0].Code != string(launcher.WarningDanglingProjectBinding) {
		t.Fatalf("warnings = %#v, want dangling binding warning", state.Warnings)
	}
}

func TestGetLaunchStateReportsFirstRunWhenNoContextsExist(t *testing.T) {
	fixture := newApplicationFixture(t)

	state, appErr := fixture.service().GetLaunchState(GetLaunchStateRequest{ProjectPath: "."})
	if appErr != nil {
		t.Fatalf("get launch state: %v", appErr)
	}

	if !state.FirstRun {
		t.Fatal("first run = false, want true")
	}
	if len(state.Contexts) != 0 {
		t.Fatalf("contexts = %#v, want none", state.Contexts)
	}
	if !state.SelectionRequired {
		t.Fatal("selection required = false, want true")
	}
}

func TestLaunchProjectBuildsPlanAndStartsProcess(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	result, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}

	if result.Project.Path != fixture.projectDir || result.Context.ID != "personal" {
		t.Fatalf("launch result = %#v, want personal current project", result)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}
	request := fixture.process.requests[0]
	if request.DetachMode != launcher.DetachModeAttached {
		t.Fatalf("detach mode = %q, want %q", request.DetachMode, launcher.DetachModeAttached)
	}
	if request.Environment["DEVCTX_CONTEXT"] != "personal" {
		t.Fatalf("DEVCTX_CONTEXT = %q, want personal", request.Environment["DEVCTX_CONTEXT"])
	}
	if request.Environment["FAKE_CONTEXT"] != "personal" {
		t.Fatalf("FAKE_CONTEXT = %q, want personal", request.Environment["FAKE_CONTEXT"])
	}
	if !reflect.DeepEqual(request.Arguments[len(request.Arguments)-1:], launcher.Arguments{fixture.projectDir}) {
		t.Fatalf("arguments = %#v, want project path as final argument", request.Arguments)
	}
}

func TestLaunchProjectRequiresMismatchConfirmation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr == nil {
		t.Fatal("launch error = nil, want mismatch confirmation error")
	}
	if appErr.Code != ErrorCodeContextMismatch {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeContextMismatch)
	}
	if appErr.ContextMismatch == nil ||
		appErr.ContextMismatch.BoundContextID != "company" ||
		appErr.ContextMismatch.RequestedContextID != "personal" {
		t.Fatalf("context mismatch = %#v, want company/personal details", appErr.ContextMismatch)
	}
	if len(fixture.process.requests) != 0 {
		t.Fatalf("process requests = %#v, want none", fixture.process.requests)
	}
}

func TestLaunchProjectAcceptsConfirmedMismatch(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.writeContext(t, fixture.context("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.projectDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	result, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath:            fixture.projectDir,
		ContextID:              "personal",
		ConfirmContextMismatch: true,
	})
	if appErr != nil {
		t.Fatalf("launch project: %v", appErr)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Code != string(launcher.WarningContextMismatch) {
		t.Fatalf("warnings = %#v, want mismatch warning", result.Warnings)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}
}

func TestLaunchProjectReturnsPresentationSafeLaunchFailure(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))
	fixture.process.err = launcher.ErrProcessStartFailed

	_, appErr := fixture.service().LaunchProject(LaunchProjectRequest{
		ProjectPath: fixture.projectDir,
		ContextID:   "personal",
	})
	if appErr == nil {
		t.Fatal("launch error = nil, want launch failure")
	}
	if appErr.Code != ErrorCodeLaunch {
		t.Fatalf("error code = %q, want %q", appErr.Code, ErrorCodeLaunch)
	}
	if len(fixture.process.requests) != 1 {
		t.Fatalf("process request count = %d, want 1", len(fixture.process.requests))
	}
}

func TestBindProjectPersistsCanonicalAssociation(t *testing.T) {
	fixture := newApplicationFixture(t)
	fixture.writeContext(t, fixture.context("personal", "Personal"))

	state, appErr := fixture.service().BindProject(BindProjectRequest{
		ProjectPath: ".",
		ContextID:   "personal",
	})
	if appErr != nil {
		t.Fatalf("bind project: %v", appErr)
	}
	if !state.Bound || state.ContextID != "personal" || state.ProjectPath != fixture.projectDir {
		t.Fatalf("binding state = %#v, want personal binding", state)
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if !reflect.DeepEqual(stored, []project.Binding{
		{
			ProjectPath: project.Path(fixture.projectDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
	}) {
		t.Fatalf("stored bindings = %#v, want personal binding", stored)
	}
}

func TestUnbindProjectReturnsRefreshedBindingState(t *testing.T) {
	t.Run("bound", func(t *testing.T) {
		fixture := newApplicationFixture(t)
		fixture.writeContext(t, fixture.context("personal", "Personal"))
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.projectDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		})

		state, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("unbind project: %v", appErr)
		}
		if state.Bound || state.ProjectPath != fixture.projectDir {
			t.Fatalf("binding state = %#v, want unbound project", state)
		}

		stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
		if err != nil {
			t.Fatalf("read project bindings: %v", err)
		}
		if len(stored) != 0 {
			t.Fatalf("stored bindings = %#v, want empty", stored)
		}
	})

	t.Run("already unbound", func(t *testing.T) {
		fixture := newApplicationFixture(t)

		state, appErr := fixture.service().UnbindProject(UnbindProjectRequest{ProjectPath: "."})
		if appErr != nil {
			t.Fatalf("unbind project: %v", appErr)
		}
		if state.Bound || state.ProjectPath != fixture.projectDir {
			t.Fatalf("binding state = %#v, want unbound project", state)
		}
	})
}

func assertContextStates(t *testing.T, got []ContextState, want []ContextState) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("contexts = %#v, want %#v", got, want)
	}
}

type applicationFixture struct {
	root         string
	homeDir      string
	contextsDir  string
	projectDir   string
	bindingsPath string
	paths        filesystem.PlatformPaths
	now          time.Time
	provider     *applicationFakeProvider
	editor       *applicationFakeEditor
	process      *applicationFakeProcessLauncher
}

func newApplicationFixture(t *testing.T) applicationFixture {
	t.Helper()

	root := t.TempDir()
	fixture := applicationFixture{
		root:         root,
		homeDir:      filepath.Join(root, "home"),
		contextsDir:  filepath.Join(root, "contexts"),
		projectDir:   filepath.Join(root, "projects", "current"),
		bindingsPath: filepath.Join(root, "projects.toml"),
		now:          time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
		provider:     &applicationFakeProvider{id: "fake"},
		editor:       &applicationFakeEditor{},
		process:      &applicationFakeProcessLauncher{},
	}
	fixture.paths = filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return fixture.homeDir, nil
	})
	mkdir(t, fixture.homeDir)
	mkdir(t, fixture.contextsDir)
	mkdir(t, fixture.projectDir)
	return fixture
}

func (f applicationFixture) service() *Service {
	return NewServiceWithDependencies(Dependencies{
		Contexts:          devcontext.NewRepository(f.contextsDir),
		Projects:          project.NewRepository(f.bindingsPath, f.paths),
		Paths:             f.paths,
		Providers:         []provider.Provider{f.provider},
		Editor:            f.editor,
		ProcessLauncher:   f.process,
		ParentEnvironment: []string{"PATH=/fixture/bin"},
		WorkingDirectory:  f.projectDir,
		DetachMode:        launcher.DetachModeAttached,
		Now: func() time.Time {
			return f.now
		},
	})
}

func (f applicationFixture) context(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:     devcontext.MustID(id),
		Name:   name,
		Editor: editor.DefaultConfig(),
		Providers: provider.Configs{
			"fake": {Enabled: true},
		},
		Metadata: devcontext.Metadata{
			"accent": "blue",
		},
		CreatedAt: f.now,
	}
}

func (f applicationFixture) writeContext(t *testing.T, ctx devcontext.Context) {
	t.Helper()

	mkdir(t, filepath.Join(f.contextsDir, ctx.ID.String()))
	if err := devcontext.NewRepository(f.contextsDir).Write(ctx); err != nil {
		t.Fatalf("write context %q: %v", ctx.ID.String(), err)
	}
}

func (f applicationFixture) writeBindings(t *testing.T, bindings ...project.Binding) {
	t.Helper()

	if err := project.WriteProjectBindingsFile(f.bindingsPath, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory %q: %v", path, err)
	}
}
