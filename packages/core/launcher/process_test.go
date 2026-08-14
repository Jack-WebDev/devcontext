package launcher_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"devctx/packages/core/launcher"
)

type recordingProcessLauncher struct {
	requests []launcher.ProcessRequest
}

var _ launcher.ProcessLauncher = (*recordingProcessLauncher)(nil)

func (l *recordingProcessLauncher) Launch(request launcher.ProcessRequest) error {
	l.requests = append(l.requests, request)
	return nil
}

func TestProcessLauncherInterfaceRecordsCompleteProcessRequest(t *testing.T) {
	processLauncher := &recordingProcessLauncher{}
	request := launcher.ProcessRequest{
		Executable: launcher.Executable("/usr/local/bin/code"),
		Arguments:  launcher.Arguments{"/work/client-a/api"},
		Environment: launcher.Environment{
			"CLAUDE_CONFIG_DIR": "/home/alex/.devctx/contexts/client-a/claude",
			"CODEX_HOME":        "/home/alex/.devctx/contexts/client-a/codex",
			"DEVCTX_CONTEXT":    "client-a",
		},
		WorkingDirectory: launcher.WorkingDirectory("/work/client-a/api"),
		DetachMode:       launcher.DetachModeDetached,
	}

	if err := processLauncher.Launch(request); err != nil {
		t.Fatalf("launch process: %v", err)
	}

	if !reflect.DeepEqual(processLauncher.requests, []launcher.ProcessRequest{request}) {
		t.Fatalf("requests = %#v, want %#v", processLauncher.requests, []launcher.ProcessRequest{request})
	}
}

func TestNativeProcessLauncherLaunchesFixtureWithStructuredRequest(t *testing.T) {
	workingDirectory := t.TempDir()
	recordPath := workingDirectory + string(os.PathSeparator) + "process-record.json"
	processLauncher := launcher.NativeProcessLauncher{}

	request := launcher.ProcessRequest{
		Executable: launcher.Executable(os.Args[0]),
		Arguments: launcher.Arguments{
			"-test.run=TestNativeProcessLauncherHelper",
			"--",
			"/work/client a/api",
			"--literal=value with spaces",
		},
		Environment: launcher.Environment{
			"DEVCTX_HELPER_PROCESS": "1",
			"DEVCTX_RECORD_PATH":    recordPath,
			"DEVCTX_CONTEXT":        "client-a",
			"CODEX_HOME":            "/home/alex/.devctx/contexts/client-a/codex",
		},
		WorkingDirectory: launcher.WorkingDirectory(workingDirectory),
		DetachMode:       launcher.DetachModeAttached,
	}

	if err := processLauncher.Launch(request); err != nil {
		t.Fatalf("launch process: %v", err)
	}

	record := readProcessRecord(t, recordPath)
	wantArgs := []string{"/work/client a/api", "--literal=value with spaces"}
	if !reflect.DeepEqual(record.Args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", record.Args, wantArgs)
	}
	if record.WorkingDirectory != workingDirectory {
		t.Fatalf("working directory = %q, want %q", record.WorkingDirectory, workingDirectory)
	}
	if record.Environment["DEVCTX_CONTEXT"] != "client-a" {
		t.Fatalf("DEVCTX_CONTEXT = %q, want %q", record.Environment["DEVCTX_CONTEXT"], "client-a")
	}
	if record.Environment["CODEX_HOME"] != "/home/alex/.devctx/contexts/client-a/codex" {
		t.Fatalf("CODEX_HOME = %q, want %q", record.Environment["CODEX_HOME"], "/home/alex/.devctx/contexts/client-a/codex")
	}
}

func TestNativeProcessLauncherDetachesFixtureProcess(t *testing.T) {
	workingDirectory := t.TempDir()
	startedPath := workingDirectory + string(os.PathSeparator) + "detached-started"
	donePath := workingDirectory + string(os.PathSeparator) + "detached-done"

	startedAt := time.Now()
	err := launcher.NativeProcessLauncher{}.Launch(launcher.ProcessRequest{
		Executable: launcher.Executable(os.Args[0]),
		Arguments: launcher.Arguments{
			"-test.run=TestNativeProcessLauncherDetachedHelper",
		},
		Environment: launcher.Environment{
			"DEVCTX_DETACHED_HELPER_PROCESS": "1",
			"DEVCTX_DETACHED_STARTED_PATH":   startedPath,
			"DEVCTX_DETACHED_DONE_PATH":      donePath,
		},
		WorkingDirectory: launcher.WorkingDirectory(workingDirectory),
		DetachMode:       launcher.DetachModeDetached,
	})
	if err != nil {
		t.Fatalf("launch detached process: %v", err)
	}

	if elapsed := time.Since(startedAt); elapsed > 300*time.Millisecond {
		t.Fatalf("detached launch took %s, want it to return without waiting", elapsed)
	}
	waitForFile(t, startedPath)
	if _, err := os.Stat(donePath); err == nil {
		t.Fatalf("detached child completed before launcher returned")
	}
	waitForFile(t, donePath)
}

func TestNativeProcessLauncherMapsLaunchFailures(t *testing.T) {
	workingDirectory := t.TempDir()

	tests := []struct {
		name    string
		request launcher.ProcessRequest
		wantErr error
		skip    func(t *testing.T)
	}{
		{
			name: "executable missing",
			request: launcher.ProcessRequest{
				Executable:       launcher.Executable(workingDirectory + string(os.PathSeparator) + "missing-executable"),
				WorkingDirectory: launcher.WorkingDirectory(workingDirectory),
			},
			wantErr: launcher.ErrProcessExecutableNotFound,
		},
		{
			name: "permission denied",
			request: launcher.ProcessRequest{
				Executable:       launcher.Executable(nonExecutableFixture(t, workingDirectory)),
				WorkingDirectory: launcher.WorkingDirectory(workingDirectory),
			},
			wantErr: launcher.ErrProcessPermissionDenied,
			skip: func(t *testing.T) {
				t.Helper()
				if runtime.GOOS == "windows" {
					t.Skip("Windows executable permissions are not represented by Unix mode bits")
				}
			},
		},
		{
			name: "invalid working directory",
			request: launcher.ProcessRequest{
				Executable:       launcher.Executable(os.Args[0]),
				WorkingDirectory: launcher.WorkingDirectory(workingDirectory + string(os.PathSeparator) + "missing-directory"),
			},
			wantErr: launcher.ErrProcessWorkingDirectoryInvalid,
		},
		{
			name: "generic start failure",
			request: launcher.ProcessRequest{
				Executable: launcher.Executable(os.Args[0]),
				Arguments: launcher.Arguments{
					"-test.run=TestNativeProcessLauncherFailingHelper",
				},
				Environment: launcher.Environment{
					"DEVCTX_FAILING_HELPER_PROCESS": "1",
				},
				WorkingDirectory: launcher.WorkingDirectory(workingDirectory),
			},
			wantErr: launcher.ErrProcessStartFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != nil {
				tt.skip(t)
			}

			err := launcher.NativeProcessLauncher{}.Launch(tt.request)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}

			var launchError *launcher.ProcessLaunchError
			if !errors.As(err, &launchError) {
				t.Fatalf("error = %T, want *launcher.ProcessLaunchError", err)
			}
		})
	}
}

func TestEnvironmentEnvironReturnsDeterministicEntries(t *testing.T) {
	environment := launcher.Environment{
		"Z_VAR": "last",
		"A_VAR": "first",
		"":      "ignored",
	}

	got := environment.Environ()
	want := []string{"A_VAR=first", "Z_VAR=last"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("environment = %#v, want %#v", got, want)
	}
}

func TestNativeProcessLauncherHelper(t *testing.T) {
	if os.Getenv("DEVCTX_HELPER_PROCESS") != "1" {
		return
	}

	recordPath := os.Getenv("DEVCTX_RECORD_PATH")
	if recordPath == "" {
		t.Fatal("missing DEVCTX_RECORD_PATH")
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	record := processRecord{
		Args:             helperArgs(os.Args),
		Environment:      environMap(os.Environ()),
		WorkingDirectory: workingDirectory,
	}

	data, err := json.Marshal(record)
	if err != nil {
		t.Fatalf("marshal process record: %v", err)
	}
	if err := os.WriteFile(recordPath, data, 0o600); err != nil {
		t.Fatalf("write process record: %v", err)
	}
}

func TestNativeProcessLauncherDetachedHelper(t *testing.T) {
	if os.Getenv("DEVCTX_DETACHED_HELPER_PROCESS") != "1" {
		return
	}

	startedPath := os.Getenv("DEVCTX_DETACHED_STARTED_PATH")
	donePath := os.Getenv("DEVCTX_DETACHED_DONE_PATH")
	if startedPath == "" || donePath == "" {
		t.Fatal("missing detached helper paths")
	}

	if err := os.WriteFile(startedPath, []byte("started"), 0o600); err != nil {
		t.Fatalf("write started marker: %v", err)
	}
	time.Sleep(500 * time.Millisecond)
	if err := os.WriteFile(donePath, []byte("done"), 0o600); err != nil {
		t.Fatalf("write done marker: %v", err)
	}
}

func TestNativeProcessLauncherFailingHelper(t *testing.T) {
	if os.Getenv("DEVCTX_FAILING_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(7)
}

type processRecord struct {
	Args             []string          `json:"args"`
	Environment      map[string]string `json:"environment"`
	WorkingDirectory string            `json:"working_directory"`
}

func readProcessRecord(t *testing.T, path string) processRecord {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read process record: %v", err)
	}

	var record processRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("decode process record: %v", err)
	}
	return record
}

func helperArgs(args []string) []string {
	for index, arg := range args {
		if arg == "--" {
			return append([]string(nil), args[index+1:]...)
		}
	}
	return nil
}

func environMap(entries []string) map[string]string {
	environment := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		environment[key] = value
	}
	return environment
}

func waitForFile(t *testing.T, path string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %q", path)
}

func nonExecutableFixture(t *testing.T, dir string) string {
	t.Helper()

	path := dir + string(os.PathSeparator) + "not-executable"
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatalf("write non-executable fixture: %v", err)
	}
	return path
}
