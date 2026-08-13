package editor_test

import (
	"reflect"
	"testing"

	"devctx/packages/core/editor"
)

type fakeEditor struct{}

var _ editor.Editor = fakeEditor{}

func (fakeEditor) ID() editor.ID {
	return "fake-editor"
}

func (fakeEditor) DetectExecutable(config editor.Config) (editor.Executable, error) {
	if config.ExecutableOverride != "" {
		return editor.Executable(config.ExecutableOverride), nil
	}
	return editor.Executable("/usr/local/bin/fake-editor"), nil
}

func (fakeEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	return editor.Command{
		Executable: request.Executable,
		Arguments: editor.Arguments{
			"--state-dir",
			request.Paths.UserDataDir,
			request.ProjectPath,
		},
	}, nil
}

func TestEditorInterfaceAllowsGenericEditorUse(t *testing.T) {
	var implementation editor.Editor = fakeEditor{}
	config := editor.Config{
		Type:               "fake-editor",
		ExecutableOverride: "/opt/fake-editor/bin/fake-editor",
	}

	if implementation.ID() != "fake-editor" {
		t.Fatalf("id = %q, want %q", implementation.ID(), "fake-editor")
	}

	executable, err := implementation.DetectExecutable(config)
	if err != nil {
		t.Fatalf("detect executable: %v", err)
	}
	if executable != "/opt/fake-editor/bin/fake-editor" {
		t.Fatalf("executable = %q, want %q", executable, "/opt/fake-editor/bin/fake-editor")
	}

	command, err := implementation.BuildLaunchCommand(editor.CommandRequest{
		Config:      config,
		Executable:  executable,
		ProjectPath: "/work/client-a/api",
		Paths: editor.ContextPaths{
			RootDir:     "/home/alex/.devctx/contexts/client-a",
			DataDir:     "/home/alex/.devctx/contexts/client-a/fake-editor",
			UserDataDir: "/home/alex/.devctx/contexts/client-a/fake-editor/user-data",
		},
	})
	if err != nil {
		t.Fatalf("build launch command: %v", err)
	}

	want := editor.Command{
		Executable: "/opt/fake-editor/bin/fake-editor",
		Arguments: editor.Arguments{
			"--state-dir",
			"/home/alex/.devctx/contexts/client-a/fake-editor/user-data",
			"/work/client-a/api",
		},
	}
	if !reflect.DeepEqual(command, want) {
		t.Fatalf("command = %#v, want %#v", command, want)
	}
}
