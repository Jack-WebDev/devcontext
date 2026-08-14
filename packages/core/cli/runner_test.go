package cli_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"devctx/packages/core/cli"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

func TestRunnerContextListRendersEmptyAndPopulatedContexts(t *testing.T) {
	fixture := newRunnerFixture(t)

	empty := fixture.runner().Run([]string{"context", "list"})
	assertResult(t, empty, cli.ExitSuccess, "No contexts configured.\n", "")

	client := testCLIContext("client-a", "Client A")
	company := testCLIContext("company", "Company")
	personal := testCLIContext("personal", "Personal")
	fixture.writeContext(t, personal)
	fixture.writeContext(t, company)
	fixture.writeContext(t, client)

	populated := fixture.runner().Run([]string{"context", "list"})
	assertResult(t, populated, cli.ExitSuccess, ""+
		"NAME      ID\n"+
		"Client A  client-a\n"+
		"Company   company\n"+
		"Personal  personal\n", "")
}

func TestRunnerContextCreateDoesNotImportUnclassifiedProviderCredentials(t *testing.T) {
	fixture := newRunnerFixture(t)
	writeCLICredentialFixture(t, filepath.Join(fixture.homeDir, ".codex", "auth.json"), []byte("codex-auth-fixture"))
	writeCLICredentialFixture(t, filepath.Join(fixture.homeDir, ".claude", ".credentials.json"), []byte("claude-credentials-fixture"))
	writeCLICredentialFixture(t, filepath.Join(fixture.homeDir, ".claude", "settings.json"), []byte("claude-settings-fixture"))

	result := fixture.runner().Run([]string{"context", "create", "personal"})
	assertResult(t, result, cli.ExitSuccess, "Context:\npersonal\n\nStatus:\ncreated\n", "")

	stored, err := devcontext.NewRepository(fixture.contextsDir).Get(devcontext.MustID("personal"))
	if err != nil {
		t.Fatalf("get created context: %v", err)
	}
	if !reflect.DeepEqual(stored, devcontext.DefaultPersonalContext(fixture.now)) {
		t.Fatalf("stored context = %#v, want default personal", stored)
	}

	contextRoot := filepath.Join(fixture.homeDir, ".devctx", "contexts", "personal")
	assertCLIDirectoryEmpty(t, filepath.Join(contextRoot, "codex"))
	assertCLIDirectoryEmpty(t, filepath.Join(contextRoot, "claude"))
}

func TestRunnerProjectShowRendersBoundUnboundAndDanglingStates(t *testing.T) {
	t.Run("unbound", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "show"})
		assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n", "")
	})

	t.Run("bound", func(t *testing.T) {
		fixture := newRunnerFixture(t)
		fixture.writeContext(t, testCLIContext("personal", "Personal"))
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		})

		result := fixture.runner().Run([]string{"project", "show"})
		assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n", "")
	})

	t.Run("dangling", func(t *testing.T) {
		fixture := newRunnerFixture(t)
		fixture.writeBindings(t, project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("company"),
			CreatedAt:   fixture.now,
		})

		result := fixture.runner().Run([]string{"project", "show"})
		if result.Code != cli.ExitSuccess {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
		}
		for _, want := range []string{
			"Project:\n" + fixture.workingDir,
			"Context:\nmissing: company",
			"Binding:\ndangling",
			project.DanglingProjectBindingRecovery,
		} {
			if !strings.Contains(result.Stdout, want) {
				t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
			}
		}
	})
}

func TestRunnerProjectBindPersistsBindingAndShowReportsIt(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	result := fixture.runner().Run([]string{"project", "bind", "personal"})
	if result.Code != cli.ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
	}
	for _, want := range []string{
		"Project:\n" + fixture.workingDir,
		"Context:\npersonal",
		"Status:\nbound",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
		}
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("binding count = %d, want 1", len(stored))
	}
	if stored[0].ProjectPath != project.Path(fixture.workingDir) {
		t.Fatalf("project path = %q, want %q", stored[0].ProjectPath, fixture.workingDir)
	}
	if stored[0].ContextID != devcontext.MustID("personal") {
		t.Fatalf("context ID = %q, want personal", stored[0].ContextID.String())
	}
	if !stored[0].CreatedAt.Equal(fixture.now) {
		t.Fatalf("created at = %s, want %s", stored[0].CreatedAt, fixture.now)
	}

	restartedRunner := fixture.runner()
	show := restartedRunner.Run([]string{"project", "show"})
	assertResult(t, show, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n", "")
}

func TestRunnerProjectBindReturnsValidationErrors(t *testing.T) {
	t.Run("invalid context ID", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "bind", "Personal"})
		if result.Code != cli.ExitValidationError {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitValidationError, result.Stderr)
		}
		if !strings.Contains(result.Stderr, "Unable to use context ID") {
			t.Fatalf("stderr = %q, want context ID guidance", result.Stderr)
		}
	})

	t.Run("missing context", func(t *testing.T) {
		fixture := newRunnerFixture(t)

		result := fixture.runner().Run([]string{"project", "bind", "personal"})
		if result.Code != cli.ExitValidationError {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitValidationError, result.Stderr)
		}
		if !strings.Contains(result.Stderr, "Unable to find context") {
			t.Fatalf("stderr = %q, want missing context guidance", result.Stderr)
		}
	})
}

func TestRunnerProjectUnbindRemovesOnlyCurrentProjectBinding(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))
	otherProject := fixture.mkdir(t, "projects", "other")
	fixture.writeBindings(t,
		project.Binding{
			ProjectPath: project.Path(fixture.workingDir),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
		project.Binding{
			ProjectPath: project.Path(otherProject),
			ContextID:   devcontext.MustID("personal"),
			CreatedAt:   fixture.now,
		},
	)

	result := fixture.runner().Run([]string{"project", "unbind"})
	if result.Code != cli.ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
	}
	for _, want := range []string{
		"Project:\n" + fixture.workingDir,
		"Removed context:\npersonal",
		"Status:\nunbound",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
		}
	}

	stored, err := project.ReadProjectBindingsFile(fixture.bindingsPath)
	if err != nil {
		t.Fatalf("read project bindings: %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("binding count = %d, want 1", len(stored))
	}
	if stored[0].ProjectPath != project.Path(otherProject) {
		t.Fatalf("remaining project path = %q, want %q", stored[0].ProjectPath, otherProject)
	}

	show := fixture.runner().Run([]string{"project", "show"})
	assertResult(t, show, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n", "")
}

func TestRunnerProjectUnbindIsIdempotentForUnboundProject(t *testing.T) {
	fixture := newRunnerFixture(t)

	result := fixture.runner().Run([]string{"project", "unbind"})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\nunbound\n\nStatus:\nunchanged\n", "")
}

func TestRunnerRootLaunchBuildsPlanAndStartsDetachedProcess(t *testing.T) {
	fixture := newRunnerFixture(t)
	context := testCLIContext("personal", "Personal")
	context.Providers = provider.Configs{
		provider.ClaudeID: {Enabled: true},
		provider.CodexID:  {Enabled: true},
	}
	fixture.writeContext(t, context)

	launchEditor := &recordingCLIEditor{}
	processLauncher := &recordingProcessLauncher{}
	runner := fixture.runner()
	runner.Providers = []provider.Provider{provider.ClaudeProvider{}, provider.CodexProvider{}}
	runner.Editor = launchEditor
	runner.ProcessLauncher = processLauncher
	runner.ParentEnvironment = []string{
		"PATH=/usr/local/bin",
		"CODEX_HOME=/parent/codex",
		"CLAUDE_CONFIG_DIR=/parent/claude",
	}

	result := runner.Run([]string{"--context", "personal", "."})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n\nStatus:\nlaunched\n", "")

	contextRoot := filepath.Join(fixture.homeDir, ".devctx", "contexts", "personal")
	wantRequest := launcher.ProcessRequest{
		Executable: launcher.Executable("/recording/code"),
		Arguments:  launcher.Arguments{fixture.workingDir},
		Environment: launcher.Environment{
			"PATH":              "/usr/local/bin",
			"CODEX_HOME":        filepath.Join(contextRoot, "codex"),
			"CLAUDE_CONFIG_DIR": filepath.Join(contextRoot, "claude"),
			"DEVCTX_CONTEXT":    "personal",
		},
		WorkingDirectory: launcher.WorkingDirectory(fixture.workingDir),
		DetachMode:       launcher.DetachModeDetached,
	}
	if !reflect.DeepEqual(processLauncher.requests, []launcher.ProcessRequest{wantRequest}) {
		t.Fatalf("process requests = %#v, want %#v", processLauncher.requests, []launcher.ProcessRequest{wantRequest})
	}

	wantEditorRequest := editor.CommandRequest{
		Config:      context.Editor,
		Executable:  "/recording/code",
		ProjectPath: fixture.workingDir,
		Paths: editor.ContextPaths{
			RootDir:     contextRoot,
			DataDir:     filepath.Join(contextRoot, "vscode"),
			UserDataDir: filepath.Join(contextRoot, "vscode", "user-data"),
		},
	}
	if !reflect.DeepEqual(launchEditor.requests, []editor.CommandRequest{wantEditorRequest}) {
		t.Fatalf("editor requests = %#v, want %#v", launchEditor.requests, []editor.CommandRequest{wantEditorRequest})
	}
}

func TestRunnerVersionRendersWithoutRepositories(t *testing.T) {
	result := cli.Runner{}.Run([]string{"--version"})
	assertResult(t, result, cli.ExitSuccess, "devctx dev\n", "")
}

func TestRunnerRootLaunchWritesLifecycleEvents(t *testing.T) {
	fixture := newRunnerFixture(t)
	context := testCLIContext("personal", "Personal")
	context.Providers = provider.Configs{
		"missing-provider": {Enabled: true},
	}
	fixture.writeContext(t, context)

	runner := fixture.runner()
	runner.Editor = &recordingCLIEditor{}
	runner.ProcessLauncher = &recordingProcessLauncher{}
	runner.ParentEnvironment = []string{"PATH=/usr/local/bin"}
	runner.Logger = devlog.NewLocalLogger(filepath.Join(fixture.root, "logs"), filesystem.NewDefaultStoragePermissions(), func() time.Time {
		return fixture.now
	})

	result := runner.Run([]string{"--context", "personal", "."})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n\nStatus:\nlaunched\n", "")

	events := readLaunchLogEvents(t, filepath.Join(fixture.root, "logs", devlog.DefaultFileName))
	wantNames := []devlog.EventName{
		devlog.EventContextResolution,
		devlog.EventLaunchProviderMissing,
		devlog.EventLaunchSucceeded,
	}
	if got := eventNames(events); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("event names = %#v, want %#v", got, wantNames)
	}
	for _, event := range events {
		if event.ProjectPath != fixture.workingDir {
			t.Fatalf("event project path = %q, want %q", event.ProjectPath, fixture.workingDir)
		}
		if event.ContextID != "personal" {
			t.Fatalf("event context ID = %q, want personal", event.ContextID)
		}
	}
	if events[1].ErrorCategory != devlog.ErrorCategoryProvider {
		t.Fatalf("provider event category = %q, want %q", events[1].ErrorCategory, devlog.ErrorCategoryProvider)
	}
}

func TestRunnerRootLaunchIgnoresLoggingFailures(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	runner := fixture.runner()
	runner.Editor = &recordingCLIEditor{}
	runner.ProcessLauncher = &recordingProcessLauncher{}
	runner.ParentEnvironment = []string{"PATH=/usr/local/bin"}
	runner.Logger = failingLogger{}

	result := runner.Run([]string{"--context", "personal", "."})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n\nStatus:\nlaunched\n", "")
}

func TestRunnerRootLaunchDebugOutputRedactsSensitiveEnvironment(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	runner := fixture.runner()
	runner.Editor = &recordingCLIEditor{}
	runner.ProcessLauncher = &recordingProcessLauncher{}
	runner.ParentEnvironment = []string{
		"PATH=/usr/local/bin",
		"API_TOKEN=top-secret-token",
	}

	result := runner.Run([]string{"--debug", "--context", "personal", "."})
	if result.Code != cli.ExitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
	}
	for _, want := range []string{
		"Debug:\n",
		"resolution_source: explicit\n",
		"editor_id: vscode\n",
		"editor_executable: /recording/code\n",
		"context_directories:\n",
		"arguments:\n",
		"environment:\n",
		"API_TOKEN=<redacted>\n",
	} {
		if !strings.Contains(result.Stdout, want) {
			t.Fatalf("stdout = %q, want containing %q", result.Stdout, want)
		}
	}
	if strings.Contains(result.Stdout, "top-secret-token") {
		t.Fatalf("debug output leaked sensitive environment value: %q", result.Stdout)
	}
}

func TestRunnerRootLaunchLogsSanitizedProcessFailure(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	processErr := &launcher.ProcessLaunchError{
		Executable:       "/recording/code",
		WorkingDirectory: launcher.WorkingDirectory(fixture.workingDir),
		Err:              launcher.ErrProcessStartFailed,
		Cause:            fmt.Errorf("API_TOKEN=top-secret-token authorization: Bearer bearer-secret"),
	}
	runner := fixture.runner()
	runner.Editor = &recordingCLIEditor{}
	runner.ProcessLauncher = &recordingProcessLauncher{err: processErr}
	runner.ParentEnvironment = []string{"API_TOKEN=top-secret-token"}
	runner.Logger = devlog.NewLocalLogger(filepath.Join(fixture.root, "logs"), filesystem.NewDefaultStoragePermissions(), func() time.Time {
		return fixture.now
	})

	result := runner.Run([]string{"--context", "personal", "."})
	if result.Code != cli.ExitLaunchFailure {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitLaunchFailure, result.Stderr)
	}

	events := readLaunchLogEvents(t, filepath.Join(fixture.root, "logs", devlog.DefaultFileName))
	if got := eventNames(events); !reflect.DeepEqual(got, []devlog.EventName{devlog.EventContextResolution, devlog.EventLaunchProcessFailure}) {
		t.Fatalf("event names = %#v", got)
	}
	failure := events[1]
	if failure.ErrorCategory != devlog.ErrorCategoryProcess {
		t.Fatalf("error category = %q, want %q", failure.ErrorCategory, devlog.ErrorCategoryProcess)
	}
	if strings.Contains(failure.Error, "top-secret-token") || strings.Contains(failure.Error, "bearer-secret") {
		t.Fatalf("log event leaked sensitive error: %#v", failure)
	}
}

func TestRunnerRootLaunchRequiresMismatchConfirmation(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))
	fixture.writeContext(t, testCLIContext("company", "Company"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.workingDir),
		ContextID:   devcontext.MustID("company"),
		CreatedAt:   fixture.now,
	})

	processLauncher := &recordingProcessLauncher{}
	runner := fixture.runner()
	runner.Editor = &recordingCLIEditor{}
	runner.ProcessLauncher = processLauncher
	runner.ParentEnvironment = []string{"PATH=/usr/local/bin"}

	result := runner.Run([]string{"--context", "personal", "."})
	if result.Code != cli.ExitValidationError {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitValidationError, result.Stderr)
	}
	if !strings.Contains(result.Stderr, "Context mismatch requires confirmation") {
		t.Fatalf("stderr = %q, want mismatch confirmation guidance", result.Stderr)
	}
	if len(processLauncher.requests) != 0 {
		t.Fatalf("process requests = %#v, want none", processLauncher.requests)
	}
}

func TestRunnerRootLaunchExecutesDirectCLIWithRecordingExecutable(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))

	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	recordPath := filepath.Join(fixture.root, "recording.txt")

	runner := fixture.runner()
	runner.Editor = recordingExecutableEditor{executable: executable}
	runner.ProcessLauncher = launcher.NativeProcessLauncher{}
	runner.ParentEnvironment = []string{
		"DEVCTX_RECORDING_EXECUTABLE=1",
		"DEVCTX_RECORDING_PATH=" + recordPath,
		"PATH=/usr/local/bin",
	}
	runner.DetachMode = launcher.DetachModeAttached

	result := runner.Run([]string{"--context", "personal", fixture.workingDir})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n\nStatus:\nlaunched\n", "")

	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("read recording: %v", err)
	}
	recording := string(data)
	for _, want := range []string{
		"cwd=" + fixture.workingDir + "\n",
		"arg=-test.run=TestDirectCLIRecordingExecutableHelper\n",
		"arg=--\n",
		"arg=" + fixture.workingDir + "\n",
		"env=personal\n",
	} {
		if !strings.Contains(recording, want) {
			t.Fatalf("recording = %q, want containing %q", recording, want)
		}
	}
}

func TestRunnerRootLaunchExecutesBindingDerivedCLIWithRecordingExecutable(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, testCLIContext("personal", "Personal"))
	fixture.writeBindings(t, project.Binding{
		ProjectPath: project.Path(fixture.workingDir),
		ContextID:   devcontext.MustID("personal"),
		CreatedAt:   fixture.now,
	})

	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}
	recordPath := filepath.Join(fixture.root, "binding-recording.txt")

	runner := fixture.runner()
	runner.Editor = recordingExecutableEditor{executable: executable}
	runner.ProcessLauncher = launcher.NativeProcessLauncher{}
	runner.ParentEnvironment = []string{
		"DEVCTX_RECORDING_EXECUTABLE=1",
		"DEVCTX_RECORDING_PATH=" + recordPath,
		"PATH=/usr/local/bin",
	}
	runner.DetachMode = launcher.DetachModeAttached

	result := runner.Run([]string{fixture.workingDir})
	assertResult(t, result, cli.ExitSuccess, "Project:\n"+fixture.workingDir+"\n\nContext:\npersonal\n\nStatus:\nlaunched\n", "")

	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("read recording: %v", err)
	}
	recording := string(data)
	for _, want := range []string{
		"cwd=" + fixture.workingDir + "\n",
		"arg=-test.run=TestDirectCLIRecordingExecutableHelper\n",
		"arg=--\n",
		"arg=" + fixture.workingDir + "\n",
		"env=personal\n",
	} {
		if !strings.Contains(recording, want) {
			t.Fatalf("recording = %q, want containing %q", recording, want)
		}
	}
}

func TestRunnerRootLaunchKeepsSimultaneousContextsIsolated(t *testing.T) {
	fixture := newRunnerFixture(t)
	fixture.writeContext(t, devcontext.DefaultPersonalContext(fixture.now))
	fixture.writeContext(t, devcontext.DefaultCompanyContext(fixture.now))

	executable, err := os.Executable()
	if err != nil {
		t.Fatalf("resolve test executable: %v", err)
	}

	isolationDir := fixture.mkdir(t, "isolation")
	personalProject := fixture.mkdir(t, "projects", "personal")
	companyProject := fixture.mkdir(t, "projects", "company")

	baseRunner := fixture.runner()
	baseRunner.Editor = simultaneousIsolationEditor{executable: executable}
	baseRunner.ProcessLauncher = launcher.NativeProcessLauncher{}
	baseRunner.ParentEnvironment = []string{
		"DEVCTX_SIMULTANEOUS_ISOLATION_HELPER=1",
		"DEVCTX_SIMULTANEOUS_ISOLATION_DIR=" + isolationDir,
		"PATH=/usr/local/bin",
	}
	baseRunner.DetachMode = launcher.DetachModeAttached

	type launchCase struct {
		contextID string
		project   string
	}
	cases := []launchCase{
		{contextID: "personal", project: personalProject},
		{contextID: "company", project: companyProject},
	}

	var wg sync.WaitGroup
	outcomes := make(chan cli.Result, len(cases))
	for _, launch := range cases {
		launch := launch
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner := baseRunner
			outcomes <- runner.Run([]string{"--context", launch.contextID, launch.project})
		}()
	}
	wg.Wait()
	close(outcomes)

	for result := range outcomes {
		if result.Code != cli.ExitSuccess {
			t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitSuccess, result.Stderr)
		}
	}

	personal := readSimultaneousIsolationRecord(t, filepath.Join(isolationDir, "personal.json"))
	company := readSimultaneousIsolationRecord(t, filepath.Join(isolationDir, "company.json"))

	assertDifferent(t, personal.CodexHome, company.CodexHome, "CODEX_HOME")
	assertDifferent(t, personal.ClaudeConfigDir, company.ClaudeConfigDir, "CLAUDE_CONFIG_DIR")
	assertDifferent(t, personal.DevContext, company.DevContext, "DEVCTX_CONTEXT")
	assertDifferent(t, personal.ProjectPath, company.ProjectPath, "project path")

	assertContextOwnedMarker(t, personal, "personal")
	assertContextOwnedMarker(t, company, "company")
	assertNoContextOwnedMarker(t, personal, "company")
	assertNoContextOwnedMarker(t, company, "personal")
}

func TestSimultaneousContextIsolationHelper(t *testing.T) {
	if os.Getenv("DEVCTX_SIMULTANEOUS_ISOLATION_HELPER") != "1" {
		return
	}

	contextID := os.Getenv("DEVCTX_CONTEXT")
	isolationDir := os.Getenv("DEVCTX_SIMULTANEOUS_ISOLATION_DIR")
	if contextID == "" || isolationDir == "" {
		t.Fatal("missing simultaneous isolation helper environment")
	}

	if err := os.WriteFile(filepath.Join(isolationDir, contextID+".started"), []byte("started"), 0o600); err != nil {
		t.Fatalf("write started marker: %v", err)
	}
	waitForSimultaneousIsolationPeer(t, isolationDir, "personal")
	waitForSimultaneousIsolationPeer(t, isolationDir, "company")

	args := argsAfterDoubleDash(os.Args)
	projectPath := lastArgument(args)
	if projectPath == "" {
		t.Fatalf("invalid helper arguments: %#v", args)
	}

	record := simultaneousIsolationRecord{
		DevContext:      contextID,
		CodexHome:       os.Getenv(provider.CodexHomeEnvVar),
		ClaudeConfigDir: os.Getenv(provider.ClaudeConfigDirEnvVar),
		ProjectPath:     projectPath,
	}

	writeContextOwnedMarker(t, record.CodexHome, contextID)
	writeContextOwnedMarker(t, record.ClaudeConfigDir, contextID)
	writeSimultaneousIsolationRecord(t, filepath.Join(isolationDir, contextID+".json"), record)
	os.Exit(0)
}

func TestDirectCLIRecordingExecutableHelper(t *testing.T) {
	if os.Getenv("DEVCTX_RECORDING_EXECUTABLE") != "1" {
		return
	}

	var builder strings.Builder
	workingDirectory, err := os.Getwd()
	if err != nil {
		os.Exit(1)
	}
	fmt.Fprintf(&builder, "cwd=%s\n", workingDirectory)
	for _, arg := range os.Args[1:] {
		fmt.Fprintf(&builder, "arg=%s\n", arg)
	}
	fmt.Fprintf(&builder, "env=%s\n", os.Getenv("DEVCTX_CONTEXT"))

	if err := os.WriteFile(os.Getenv("DEVCTX_RECORDING_PATH"), []byte(builder.String()), 0o600); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestRunnerReturnsUsageExitCodeForInvalidCommandShapes(t *testing.T) {
	fixture := newRunnerFixture(t)

	result := fixture.runner().Run([]string{"project", "bind"})
	if result.Code != cli.ExitUsageError {
		t.Fatalf("exit code = %d, want %d; stderr = %q", result.Code, cli.ExitUsageError, result.Stderr)
	}
	if !strings.Contains(result.Stderr, "Unable to parse command") {
		t.Fatalf("stderr = %q, want usage guidance", result.Stderr)
	}
}

type runnerFixture struct {
	root         string
	homeDir      string
	contextsDir  string
	workingDir   string
	bindingsPath string
	paths        filesystem.PlatformPaths
	now          time.Time
}

func newRunnerFixture(t *testing.T) runnerFixture {
	t.Helper()

	root := t.TempDir()
	fixture := runnerFixture{
		root:         root,
		homeDir:      filepath.Join(root, "home"),
		contextsDir:  filepath.Join(root, "contexts"),
		workingDir:   filepath.Join(root, "projects", "current"),
		bindingsPath: filepath.Join(root, "projects.toml"),
		now:          time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
	fixture.paths = filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return fixture.homeDir, nil
	})
	devContextHomeDir, err := fixture.paths.DevContextHomeDir()
	if err != nil {
		t.Fatalf("dev context home: %v", err)
	}
	fixture.contextsDir = filepath.Join(devContextHomeDir, "contexts")

	fixture.mkdir(t, "home")
	if err := os.MkdirAll(fixture.contextsDir, 0o700); err != nil {
		t.Fatalf("create contexts directory fixture %q: %v", fixture.contextsDir, err)
	}
	fixture.mkdir(t, "projects", "current")

	return fixture
}

func (f runnerFixture) runner() cli.Runner {
	return cli.Runner{
		Contexts:         devcontext.NewRepository(f.contextsDir),
		Projects:         project.NewRepository(f.bindingsPath, f.paths),
		WorkingDirectory: f.workingDir,
		Paths:            f.paths,
		Now: func() time.Time {
			return f.now
		},
	}
}

func (f runnerFixture) mkdir(t *testing.T, elements ...string) string {
	t.Helper()

	path := filepath.Join(append([]string{f.root}, elements...)...)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory fixture %q: %v", path, err)
	}
	return path
}

func (f runnerFixture) writeContext(t *testing.T, ctx devcontext.Context) {
	t.Helper()

	contextPaths, err := filesystem.DeriveContextPaths(f.paths, ctx.ID)
	if err != nil {
		t.Fatalf("derive context paths: %v", err)
	}
	for _, dir := range []string{
		contextPaths.RootDir,
		contextPaths.ClaudeDir,
		contextPaths.CodexDir,
	} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("create context directory %q: %v", dir, err)
		}
	}
	if err := devcontext.NewRepository(f.contextsDir).Write(ctx); err != nil {
		t.Fatalf("write context %q: %v", ctx.ID.String(), err)
	}
}

func (f runnerFixture) writeBindings(t *testing.T, bindings ...project.Binding) {
	t.Helper()

	if err := project.WriteProjectBindingsFile(f.bindingsPath, bindings); err != nil {
		t.Fatalf("write project bindings: %v", err)
	}
}

func testCLIContext(id string, name string) devcontext.Context {
	return devcontext.Context{
		ID:        devcontext.MustID(id),
		Name:      name,
		Editor:    editor.DefaultConfig(),
		CreatedAt: time.Date(2026, 8, 13, 12, 30, 0, 0, time.UTC),
	}
}

func assertResult(t *testing.T, got cli.Result, wantCode cli.ExitCode, wantStdout string, wantStderr string) {
	t.Helper()

	if got.Code != wantCode {
		t.Fatalf("exit code = %d, want %d", got.Code, wantCode)
	}
	if got.Stdout != wantStdout {
		t.Fatalf("stdout = %q, want %q", got.Stdout, wantStdout)
	}
	if got.Stderr != wantStderr {
		t.Fatalf("stderr = %q, want %q", got.Stderr, wantStderr)
	}
}

func writeCLICredentialFixture(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create credential fixture directory %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write credential fixture %q: %v", path, err)
	}
}

func assertCLIDirectoryEmpty(t *testing.T, path string) {
	t.Helper()

	entries, err := os.ReadDir(path)
	if err != nil {
		t.Fatalf("read directory %q: %v", path, err)
	}
	if len(entries) != 0 {
		t.Fatalf("directory %q entries = %#v, want empty", path, entries)
	}
}

func readLaunchLogEvents(t *testing.T, path string) []devlog.Event {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read launch log %q: %v", path, err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	events := make([]devlog.Event, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var event devlog.Event
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatalf("decode launch log record %q: %v", line, err)
		}
		events = append(events, event)
	}
	return events
}

func eventNames(events []devlog.Event) []devlog.EventName {
	names := make([]devlog.EventName, len(events))
	for i, event := range events {
		names[i] = event.Name
	}
	return names
}

type recordingCLIEditor struct {
	requests []editor.CommandRequest
}

type failingLogger struct{}

func (failingLogger) Record(devlog.Event) error {
	return fmt.Errorf("logging unavailable")
}

func (e *recordingCLIEditor) ID() editor.ID {
	return editor.VSCodeID
}

func (e *recordingCLIEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return "/recording/code", nil
}

func (e *recordingCLIEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	e.requests = append(e.requests, request)
	return editor.Command{
		Executable: request.Executable,
		Arguments:  editor.Arguments{request.ProjectPath},
	}, nil
}

type recordingProcessLauncher struct {
	requests []launcher.ProcessRequest
	err      error
}

func (l *recordingProcessLauncher) Launch(request launcher.ProcessRequest) error {
	l.requests = append(l.requests, request)
	return l.err
}

type recordingExecutableEditor struct {
	executable string
}

func (e recordingExecutableEditor) ID() editor.ID {
	return editor.VSCodeID
}

func (e recordingExecutableEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return editor.Executable(e.executable), nil
}

func (e recordingExecutableEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	return editor.Command{
		Executable: request.Executable,
		Arguments: editor.Arguments{
			"-test.run=TestDirectCLIRecordingExecutableHelper",
			"--",
			request.ProjectPath,
		},
	}, nil
}

type simultaneousIsolationEditor struct {
	executable string
}

func (e simultaneousIsolationEditor) ID() editor.ID {
	return editor.VSCodeID
}

func (e simultaneousIsolationEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return editor.Executable(e.executable), nil
}

func (e simultaneousIsolationEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	return editor.Command{
		Executable: request.Executable,
		Arguments: editor.Arguments{
			"-test.run=TestSimultaneousContextIsolationHelper",
			"--",
			request.ProjectPath,
		},
	}, nil
}

type simultaneousIsolationRecord struct {
	DevContext      string `json:"devctx_context"`
	CodexHome       string `json:"codex_home"`
	ClaudeConfigDir string `json:"claude_config_dir"`
	ProjectPath     string `json:"project_path"`
}

func writeSimultaneousIsolationRecord(t *testing.T, path string, record simultaneousIsolationRecord) {
	t.Helper()

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal simultaneous isolation record: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write simultaneous isolation record: %v", err)
	}
}

func readSimultaneousIsolationRecord(t *testing.T, path string) simultaneousIsolationRecord {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read simultaneous isolation record: %v", err)
	}

	var record simultaneousIsolationRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode simultaneous isolation record: %v", err)
	}
	return record
}

func assertDifferent(t *testing.T, first string, second string, label string) {
	t.Helper()

	if first == "" || second == "" {
		t.Fatalf("%s values must both be present: %q, %q", label, first, second)
	}
	if first == second {
		t.Fatalf("%s = %q for both launches, want isolated values", label, first)
	}
}

func assertContextOwnedMarker(t *testing.T, record simultaneousIsolationRecord, contextID string) {
	t.Helper()

	for _, dir := range []string{record.CodexHome, record.ClaudeConfigDir} {
		data, err := os.ReadFile(filepath.Join(dir, contextID+".txt"))
		if err != nil {
			t.Fatalf("read %s marker in %q: %v", contextID, dir, err)
		}
		if string(data) != contextID {
			t.Fatalf("marker in %q = %q, want %q", dir, string(data), contextID)
		}
	}
}

func assertNoContextOwnedMarker(t *testing.T, record simultaneousIsolationRecord, contextID string) {
	t.Helper()

	for _, dir := range []string{record.CodexHome, record.ClaudeConfigDir} {
		path := filepath.Join(dir, contextID+".txt")
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("unexpected %s marker in %q", contextID, dir)
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect %s marker in %q: %v", contextID, dir, err)
		}
	}
}

func writeContextOwnedMarker(t *testing.T, dir string, contextID string) {
	t.Helper()

	if dir == "" {
		t.Fatalf("missing context-owned directory for %s", contextID)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create context-owned directory %q: %v", dir, err)
	}
	if err := os.WriteFile(filepath.Join(dir, contextID+".txt"), []byte(contextID), 0o600); err != nil {
		t.Fatalf("write %s marker in %q: %v", contextID, dir, err)
	}
}

func waitForSimultaneousIsolationPeer(t *testing.T, isolationDir string, contextID string) {
	t.Helper()

	path := filepath.Join(isolationDir, contextID+".started")
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		} else if !os.IsNotExist(err) {
			t.Fatalf("inspect started marker %q: %v", path, err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s launch to start", contextID)
}

func argsAfterDoubleDash(args []string) []string {
	for index, arg := range args {
		if arg == "--" {
			return append([]string(nil), args[index+1:]...)
		}
	}
	return nil
}

func lastArgument(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[len(args)-1]
}
