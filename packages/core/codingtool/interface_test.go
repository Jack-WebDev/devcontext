package codingtool_test

import (
	"reflect"
	"testing"

	codingtool "devctx/packages/core/codingtool"
)

type fakeEditor struct{}

var _ codingtool.CodingTool = fakeEditor{}

func (fakeEditor) ID() codingtool.ID {
	return "fake-editor"
}

func (fakeEditor) DetectExecutable(config codingtool.Config) (codingtool.Executable, error) {
	if config.ExecutableOverride != "" {
		return codingtool.Executable(config.ExecutableOverride), nil
	}
	return codingtool.Executable("/usr/local/bin/fake-editor"), nil
}

func (fakeEditor) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	return codingtool.Command{
		Executable: request.Executable,
		Arguments: codingtool.Arguments{
			"--state-dir",
			request.Paths.UserDataDir,
			request.ProjectPath,
		},
	}, nil
}

func TestEditorInterfaceAllowsGenericEditorUse(t *testing.T) {
	var implementation codingtool.CodingTool = fakeEditor{}
	config := codingtool.Config{
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

	command, err := implementation.BuildLaunchCommand(codingtool.CommandRequest{
		Config:      config,
		Executable:  executable,
		ProjectPath: "/work/client-a/api",
		Paths: codingtool.ContextPaths{
			RootDir:     "/home/alex/.devctx/contexts/client-a",
			DataDir:     "/home/alex/.devctx/contexts/client-a/fake-editor",
			UserDataDir: "/home/alex/.devctx/contexts/client-a/fake-editor/user-data",
		},
	})
	if err != nil {
		t.Fatalf("build launch command: %v", err)
	}

	want := codingtool.Command{
		Executable: "/opt/fake-editor/bin/fake-editor",
		Arguments: codingtool.Arguments{
			"--state-dir",
			"/home/alex/.devctx/contexts/client-a/fake-editor/user-data",
			"/work/client-a/api",
		},
	}
	if !reflect.DeepEqual(command, want) {
		t.Fatalf("command = %#v, want %#v", command, want)
	}
}
