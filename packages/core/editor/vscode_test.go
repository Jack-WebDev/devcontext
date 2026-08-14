package editor_test

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"devctx/packages/core/editor"
)

func TestVSCodeEditorDetectsUnixExecutableFromSearchPath(t *testing.T) {
	probe := fakeExecutableProbe{
		paths: map[string]string{
			"code": "/usr/local/bin/code",
		},
	}

	executable, err := (editor.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(editor.DefaultConfig())
	if err != nil {
		t.Fatalf("detect executable: %v", err)
	}

	if executable != "/usr/local/bin/code" {
		t.Fatalf("executable = %q, want %q", executable, "/usr/local/bin/code")
	}
	if !reflect.DeepEqual(probe.calls, []string{"code"}) {
		t.Fatalf("lookup calls = %#v, want %#v", probe.calls, []string{"code"})
	}
}

func TestVSCodeEditorDetectsWindowsExecutableFormsFromSearchPath(t *testing.T) {
	tests := []struct {
		name      string
		paths     map[string]string
		want      editor.Executable
		wantCalls []string
	}{
		{
			name: "code",
			paths: map[string]string{
				"code": `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
			},
			want:      `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
			wantCalls: []string{"code"},
		},
		{
			name: "code cmd",
			paths: map[string]string{
				"code.cmd": `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
			},
			want:      `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
			wantCalls: []string{"code", "code.cmd"},
		},
		{
			name: "code exe",
			paths: map[string]string{
				"Code.exe": `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\Code.exe`,
			},
			want:      `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\Code.exe`,
			wantCalls: []string{"code", "code.cmd", "Code.exe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := fakeExecutableProbe{paths: tt.paths}

			executable, err := (editor.VSCodeEditor{
				Probe:           &probe,
				OperatingSystem: "windows",
			}).DetectExecutable(editor.DefaultConfig())
			if err != nil {
				t.Fatalf("detect executable: %v", err)
			}

			if executable != tt.want {
				t.Fatalf("executable = %q, want %q", executable, tt.want)
			}
			if !reflect.DeepEqual(probe.calls, tt.wantCalls) {
				t.Fatalf("lookup calls = %#v, want %#v", probe.calls, tt.wantCalls)
			}
		})
	}
}

func TestVSCodeEditorReportsTypedExecutableNotFoundError(t *testing.T) {
	probe := fakeExecutableProbe{}

	_, err := (editor.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "windows",
	}).DetectExecutable(editor.DefaultConfig())
	if err == nil {
		t.Fatal("detect executable returned nil error, want not found")
	}
	if !errors.Is(err, editor.ErrExecutableNotFound) {
		t.Fatalf("error = %v, want %v", err, editor.ErrExecutableNotFound)
	}

	var notFound *editor.ExecutableNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("error = %T, want *editor.ExecutableNotFoundError", err)
	}
	if notFound.EditorID != editor.VSCodeID {
		t.Fatalf("editor id = %q, want %q", notFound.EditorID, editor.VSCodeID)
	}
	wantCandidates := []string{"code", "code.cmd", "Code.exe"}
	if !reflect.DeepEqual(notFound.Candidates, wantCandidates) {
		t.Fatalf("candidates = %#v, want %#v", notFound.Candidates, wantCandidates)
	}
	if !reflect.DeepEqual(probe.calls, wantCandidates) {
		t.Fatalf("lookup calls = %#v, want %#v", probe.calls, wantCandidates)
	}
}

func TestVSCodeEditorUsesConfiguredExecutableBeforeSearchPath(t *testing.T) {
	probe := fakeExecutableProbe{
		paths: map[string]string{
			"code": "/usr/local/bin/code",
		},
		files: map[string]os.FileInfo{
			"/opt/visual-studio-code/bin/code": fakeFileInfo{mode: 0o755},
		},
	}

	executable, err := (editor.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(editor.Config{
		Type:               editor.TypeVSCode,
		ExecutableOverride: "/opt/visual-studio-code/bin/code",
	})
	if err != nil {
		t.Fatalf("detect executable: %v", err)
	}

	if executable != "/opt/visual-studio-code/bin/code" {
		t.Fatalf("executable = %q, want %q", executable, "/opt/visual-studio-code/bin/code")
	}
	if len(probe.calls) != 0 {
		t.Fatalf("lookup calls = %#v, want none", probe.calls)
	}
	if !reflect.DeepEqual(probe.statCalls, []string{"/opt/visual-studio-code/bin/code"}) {
		t.Fatalf("stat calls = %#v, want %#v", probe.statCalls, []string{"/opt/visual-studio-code/bin/code"})
	}
}

func TestVSCodeEditorReportsMissingConfiguredExecutable(t *testing.T) {
	probe := fakeExecutableProbe{}

	_, err := (editor.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(editor.Config{
		Type:               editor.TypeVSCode,
		ExecutableOverride: "/missing/code",
	})
	if err == nil {
		t.Fatal("detect executable returned nil error, want missing executable")
	}
	if !errors.Is(err, editor.ErrExecutableNotFound) {
		t.Fatalf("error = %v, want %v", err, editor.ErrExecutableNotFound)
	}
	if len(probe.calls) != 0 {
		t.Fatalf("lookup calls = %#v, want none", probe.calls)
	}
	if !reflect.DeepEqual(probe.statCalls, []string{"/missing/code"}) {
		t.Fatalf("stat calls = %#v, want %#v", probe.statCalls, []string{"/missing/code"})
	}
}

func TestVSCodeEditorReportsNonExecutableConfiguredExecutable(t *testing.T) {
	probe := fakeExecutableProbe{
		files: map[string]os.FileInfo{
			"/opt/code": fakeFileInfo{mode: 0o644},
		},
	}

	_, err := (editor.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(editor.Config{
		Type:               editor.TypeVSCode,
		ExecutableOverride: "/opt/code",
	})
	if err == nil {
		t.Fatal("detect executable returned nil error, want non-executable error")
	}
	if !errors.Is(err, editor.ErrExecutableNotExecutable) {
		t.Fatalf("error = %v, want %v", err, editor.ErrExecutableNotExecutable)
	}
	if !reflect.DeepEqual(probe.statCalls, []string{"/opt/code"}) {
		t.Fatalf("stat calls = %#v, want %#v", probe.statCalls, []string{"/opt/code"})
	}
}

func TestVSCodeEditorBuildsStructuredLaunchCommand(t *testing.T) {
	tests := []struct {
		name    string
		request editor.CommandRequest
		want    editor.Command
	}{
		{
			name: "paths with spaces",
			request: editor.CommandRequest{
				Config:      editor.DefaultConfig(),
				Executable:  "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
				ProjectPath: "/Users/Alex/Work/Client A/API",
			},
			want: editor.Command{
				Executable: "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
				Arguments: editor.Arguments{
					"/Users/Alex/Work/Client A/API",
				},
			},
		},
		{
			name: "windows separators",
			request: editor.CommandRequest{
				Config:      editor.DefaultConfig(),
				Executable:  `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
				ProjectPath: `C:\Users\Alex\Projects\Client A\API`,
			},
			want: editor.Command{
				Executable: `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
				Arguments: editor.Arguments{
					`C:\Users\Alex\Projects\Client A\API`,
				},
			},
		},
		{
			name: "non ASCII paths",
			request: editor.CommandRequest{
				Config:      editor.DefaultConfig(),
				Executable:  "/usr/local/bin/code",
				ProjectPath: "/Users/Alex/équipe/Café Portal",
			},
			want: editor.Command{
				Executable: "/usr/local/bin/code",
				Arguments: editor.Arguments{
					"/Users/Alex/équipe/Café Portal",
				},
			},
		},
		{
			name: "custom executable",
			request: editor.CommandRequest{
				Config: editor.Config{
					Type:               editor.TypeVSCode,
					ExecutableOverride: "/opt/vscode-insiders/bin/code-insiders",
				},
				Executable:  "/opt/vscode-insiders/bin/code-insiders",
				ProjectPath: "/work/app",
			},
			want: editor.Command{
				Executable: "/opt/vscode-insiders/bin/code-insiders",
				Arguments: editor.Arguments{
					"/work/app",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			command, err := editor.VSCodeEditor{}.BuildLaunchCommand(tt.request)
			if err != nil {
				t.Fatalf("build launch command: %v", err)
			}

			if !reflect.DeepEqual(command, tt.want) {
				t.Fatalf("command = %#v, want %#v", command, tt.want)
			}
		})
	}
}

func TestVSCodeEditorBuildLaunchCommandRejectsMissingInputs(t *testing.T) {
	validRequest := editor.CommandRequest{
		Config:      editor.DefaultConfig(),
		Executable:  "/usr/local/bin/code",
		ProjectPath: "/work/client-a/api",
	}

	tests := []struct {
		name    string
		request editor.CommandRequest
		wantErr error
	}{
		{
			name: "missing executable",
			request: editor.CommandRequest{
				Config:      validRequest.Config,
				ProjectPath: validRequest.ProjectPath,
				Paths:       validRequest.Paths,
			},
			wantErr: editor.ErrMissingExecutable,
		},
		{
			name: "missing project path",
			request: editor.CommandRequest{
				Config:     validRequest.Config,
				Executable: validRequest.Executable,
				Paths:      validRequest.Paths,
			},
			wantErr: editor.ErrMissingProjectPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := editor.VSCodeEditor{}.BuildLaunchCommand(tt.request)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

type fakeExecutableProbe struct {
	paths     map[string]string
	files     map[string]os.FileInfo
	calls     []string
	statCalls []string
}

func (p *fakeExecutableProbe) LookPath(file string) (string, error) {
	p.calls = append(p.calls, file)
	if path, ok := p.paths[file]; ok {
		return path, nil
	}
	return "", errors.New("executable not found")
}

func (p *fakeExecutableProbe) Stat(path string) (os.FileInfo, error) {
	p.statCalls = append(p.statCalls, path)
	if info, ok := p.files[path]; ok {
		return info, nil
	}
	return nil, os.ErrNotExist
}

type fakeFileInfo struct {
	mode  os.FileMode
	isDir bool
}

func (f fakeFileInfo) Name() string {
	return "code"
}

func (f fakeFileInfo) Size() int64 {
	return 0
}

func (f fakeFileInfo) Mode() os.FileMode {
	return f.mode
}

func (f fakeFileInfo) ModTime() time.Time {
	return time.Time{}
}

func (f fakeFileInfo) IsDir() bool {
	return f.isDir
}

func (f fakeFileInfo) Sys() any {
	return nil
}
