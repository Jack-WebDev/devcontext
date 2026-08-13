package editor_test

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
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

func TestVSCodeUserDataArgumentsUseDerivedContextPaths(t *testing.T) {
	platformPaths := filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return "/home/alex", nil
	})

	personalPaths, err := filesystem.DeriveContextPaths(platformPaths, devcontext.MustID("personal"))
	if err != nil {
		t.Fatalf("derive personal paths: %v", err)
	}
	companyPaths, err := filesystem.DeriveContextPaths(platformPaths, devcontext.MustID("company"))
	if err != nil {
		t.Fatalf("derive company paths: %v", err)
	}

	personalArgs, err := editor.VSCodeUserDataArguments(editor.ContextPaths{
		RootDir:     personalPaths.RootDir,
		DataDir:     personalPaths.VSCodeDir,
		UserDataDir: personalPaths.VSCodeUserDataDir,
	})
	if err != nil {
		t.Fatalf("personal user-data arguments: %v", err)
	}
	companyArgs, err := editor.VSCodeUserDataArguments(editor.ContextPaths{
		RootDir:     companyPaths.RootDir,
		DataDir:     companyPaths.VSCodeDir,
		UserDataDir: companyPaths.VSCodeUserDataDir,
	})
	if err != nil {
		t.Fatalf("company user-data arguments: %v", err)
	}

	wantPersonalArgs := editor.Arguments{
		editor.VSCodeUserDataDirFlag,
		"/home/alex/.devctx/contexts/personal/vscode/user-data",
	}
	if !reflect.DeepEqual(personalArgs, wantPersonalArgs) {
		t.Fatalf("personal args = %#v, want %#v", personalArgs, wantPersonalArgs)
	}
	wantCompanyArgs := editor.Arguments{
		editor.VSCodeUserDataDirFlag,
		"/home/alex/.devctx/contexts/company/vscode/user-data",
	}
	if !reflect.DeepEqual(companyArgs, wantCompanyArgs) {
		t.Fatalf("company args = %#v, want %#v", companyArgs, wantCompanyArgs)
	}
	if reflect.DeepEqual(personalArgs, companyArgs) {
		t.Fatalf("personal and company args should differ: %#v", personalArgs)
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
