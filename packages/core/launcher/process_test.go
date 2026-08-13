package launcher_test

import (
	"reflect"
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
