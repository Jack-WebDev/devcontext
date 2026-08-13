package launcher_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

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
		Arguments: launcher.Arguments{
			"--user-data-dir",
			"/home/alex/.devctx/contexts/client-a/vscode/user-data",
			"/work/client-a/api",
		},
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

func TestNativeProcessLauncherRejectsUnsupportedDetachedLaunch(t *testing.T) {
	err := launcher.NativeProcessLauncher{}.Launch(launcher.ProcessRequest{
		Executable: launcher.Executable(os.Args[0]),
		DetachMode: launcher.DetachModeDetached,
	})

	if !errors.Is(err, launcher.ErrDetachedProcessUnsupported) {
		t.Fatalf("error = %v, want %v", err, launcher.ErrDetachedProcessUnsupported)
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
