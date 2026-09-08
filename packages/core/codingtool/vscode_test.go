package codingtool_test

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
)

func TestVSCodeEditorDetectsUnixExecutableFromSearchPath(t *testing.T) {
	probe := fakeExecutableProbe{
		paths: map[string]string{
			"code": "/usr/local/bin/code",
		},
	}

	executable, err := (codingtool.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(codingtool.DefaultConfig())
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
		want      codingtool.Executable
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

			executable, err := (codingtool.VSCodeEditor{
				Probe:           &probe,
				OperatingSystem: "windows",
			}).DetectExecutable(codingtool.DefaultConfig())
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

func TestVSCodeEditorDetectsWindowsInstalledApplication(t *testing.T) {
	installed := `C:\\Users\\Alex\\AppData\\Local\\Programs\\Microsoft VS Code\\Code.exe`
	probe := fakeExecutableProbe{files: map[string]os.FileInfo{
		installed: fakeFileInfo{mode: 0o644},
	}}

	executable, err := (codingtool.VSCodeEditor{
		Probe:               &probe,
		OperatingSystem:     "windows",
		WindowsInstallPaths: []string{installed},
	}).DetectExecutable(codingtool.DefaultConfig())
	if err != nil {
		t.Fatalf("detect installed VS Code: %v", err)
	}
	if executable != codingtool.Executable(installed) {
		t.Fatalf("executable = %q, want %q", executable, installed)
	}
}

func TestVSCodeEditorReportsHowWindowsExecutableWasResolved(t *testing.T) {
	installed := `C:\\Users\\Alex\\AppData\\Local\\Programs\\Microsoft VS Code\\Code.exe`
	tests := []struct {
		name   string
		config codingtool.Config
		probe  fakeExecutableProbe
		paths  []string
		want   codingtool.ExecutableDetectionSource
	}{
		{
			name:  "PATH command",
			probe: fakeExecutableProbe{paths: map[string]string{"code": `C:\\Program Files\\Microsoft VS Code\\bin\\code.cmd`}},
			want:  codingtool.ExecutableDetectionPath,
		},
		{
			name:  "installed application without PATH command",
			probe: fakeExecutableProbe{files: map[string]os.FileInfo{installed: fakeFileInfo{mode: 0o644}}},
			paths: []string{installed},
			want:  codingtool.ExecutableDetectionInstalled,
		},
		{
			name:   "configured executable",
			config: codingtool.Config{ExecutableOverride: installed},
			probe:  fakeExecutableProbe{files: map[string]os.FileInfo{installed: fakeFileInfo{mode: 0o644}}},
			want:   codingtool.ExecutableDetectionConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection, err := (codingtool.VSCodeEditor{
				Probe:               &tt.probe,
				OperatingSystem:     "windows",
				WindowsInstallPaths: tt.paths,
			}).DetectExecutableDetailed(tt.config)
			if err != nil {
				t.Fatalf("detect executable: %v", err)
			}
			if detection.Platform != "windows" || detection.Source != tt.want || detection.Executable == "" {
				t.Fatalf("detection = %#v", detection)
			}
		})
	}
}

func TestVSCodeEditorReportsHowLinuxExecutableWasResolved(t *testing.T) {
	installed := "/opt/visual-studio-code/code"
	tests := []struct {
		name   string
		config codingtool.Config
		probe  fakeExecutableProbe
		paths  []string
		want   codingtool.ExecutableDetectionSource
	}{
		{
			name:  "PATH command",
			probe: fakeExecutableProbe{paths: map[string]string{"code": "/usr/local/bin/code"}},
			want:  codingtool.ExecutableDetectionPath,
		},
		{
			name:  "installed application without PATH command",
			probe: fakeExecutableProbe{files: map[string]os.FileInfo{installed: fakeFileInfo{mode: 0o755}}},
			paths: []string{installed},
			want:  codingtool.ExecutableDetectionInstalled,
		},
		{
			name:   "configured executable",
			config: codingtool.Config{ExecutableOverride: installed},
			probe:  fakeExecutableProbe{files: map[string]os.FileInfo{installed: fakeFileInfo{mode: 0o755}}},
			want:   codingtool.ExecutableDetectionConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection, err := (codingtool.VSCodeEditor{
				Probe:             &tt.probe,
				OperatingSystem:   "linux",
				LinuxInstallPaths: tt.paths,
			}).DetectExecutableDetailed(tt.config)
			if err != nil {
				t.Fatalf("detect executable: %v", err)
			}
			if detection.Platform != "linux" || detection.Source != tt.want || detection.Executable == "" {
				t.Fatalf("detection = %#v", detection)
			}
		})
	}
}

func TestVSCodeEditorReportsNonExecutableLinuxInstallation(t *testing.T) {
	installed := "/opt/visual-studio-code/code"
	detection, err := (codingtool.VSCodeEditor{
		Probe:             &fakeExecutableProbe{files: map[string]os.FileInfo{installed: fakeFileInfo{mode: 0o644}}},
		OperatingSystem:   "linux",
		LinuxInstallPaths: []string{installed},
	}).DetectExecutableDetailed(codingtool.DefaultConfig())
	if !errors.Is(err, codingtool.ErrExecutableNotExecutable) {
		t.Fatalf("error = %v, want %v", err, codingtool.ErrExecutableNotExecutable)
	}
	if detection.Platform != "linux" || detection.Source != codingtool.ExecutableDetectionInstalled {
		t.Fatalf("detection = %#v", detection)
	}
}

func TestVSCodeEditorReportsHowMacOSExecutableWasResolved(t *testing.T) {
	bundleCommand := "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"
	tests := []struct {
		name   string
		config codingtool.Config
		probe  fakeExecutableProbe
		paths  []string
		want   codingtool.ExecutableDetectionSource
	}{
		{
			name:  "PATH command",
			probe: fakeExecutableProbe{paths: map[string]string{"code": "/usr/local/bin/code"}},
			want:  codingtool.ExecutableDetectionPath,
		},
		{
			name:  "application bundle without shell command",
			probe: fakeExecutableProbe{files: map[string]os.FileInfo{bundleCommand: fakeFileInfo{mode: 0o755}}},
			paths: []string{bundleCommand},
			want:  codingtool.ExecutableDetectionInstalled,
		},
		{
			name:   "configured executable",
			config: codingtool.Config{ExecutableOverride: bundleCommand},
			probe:  fakeExecutableProbe{files: map[string]os.FileInfo{bundleCommand: fakeFileInfo{mode: 0o755}}},
			want:   codingtool.ExecutableDetectionConfigured,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detection, err := (codingtool.VSCodeEditor{
				Probe:             &tt.probe,
				OperatingSystem:   "darwin",
				MacOSInstallPaths: tt.paths,
			}).DetectExecutableDetailed(tt.config)
			if err != nil {
				t.Fatalf("detect executable: %v", err)
			}
			if detection.Platform != "darwin" || detection.Source != tt.want || detection.Executable == "" {
				t.Fatalf("detection = %#v", detection)
			}
		})
	}
}

func TestVSCodeEditorReportsUnusableMacOSApplicationBundle(t *testing.T) {
	bundleCommand := "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"
	detection, err := (codingtool.VSCodeEditor{
		Probe:             &fakeExecutableProbe{files: map[string]os.FileInfo{bundleCommand: fakeFileInfo{mode: 0o644}}},
		OperatingSystem:   "darwin",
		MacOSInstallPaths: []string{bundleCommand},
	}).DetectExecutableDetailed(codingtool.DefaultConfig())
	if !errors.Is(err, codingtool.ErrExecutableNotExecutable) {
		t.Fatalf("error = %v, want %v", err, codingtool.ErrExecutableNotExecutable)
	}
	if detection.Platform != "darwin" || detection.Source != codingtool.ExecutableDetectionInstalled {
		t.Fatalf("detection = %#v", detection)
	}
}

func TestVSCodeEditorReportsTypedExecutableNotFoundError(t *testing.T) {
	probe := fakeExecutableProbe{}

	_, err := (codingtool.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "windows",
	}).DetectExecutable(codingtool.DefaultConfig())
	if err == nil {
		t.Fatal("detect executable returned nil error, want not found")
	}
	if !errors.Is(err, codingtool.ErrExecutableNotFound) {
		t.Fatalf("error = %v, want %v", err, codingtool.ErrExecutableNotFound)
	}

	var notFound *codingtool.ExecutableNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("error = %T, want *codingtool.ExecutableNotFoundError", err)
	}
	if notFound.ToolID != codingtool.VSCodeID {
		t.Fatalf("editor id = %q, want %q", notFound.ToolID, codingtool.VSCodeID)
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

	executable, err := (codingtool.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(codingtool.Config{
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

	_, err := (codingtool.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(codingtool.Config{
		ExecutableOverride: "/missing/code",
	})
	if err == nil {
		t.Fatal("detect executable returned nil error, want missing executable")
	}
	if !errors.Is(err, codingtool.ErrExecutableNotFound) {
		t.Fatalf("error = %v, want %v", err, codingtool.ErrExecutableNotFound)
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

	_, err := (codingtool.VSCodeEditor{
		Probe:           &probe,
		OperatingSystem: "linux",
	}).DetectExecutable(codingtool.Config{
		ExecutableOverride: "/opt/code",
	})
	if err == nil {
		t.Fatal("detect executable returned nil error, want non-executable error")
	}
	if !errors.Is(err, codingtool.ErrExecutableNotExecutable) {
		t.Fatalf("error = %v, want %v", err, codingtool.ErrExecutableNotExecutable)
	}
	if !reflect.DeepEqual(probe.statCalls, []string{"/opt/code"}) {
		t.Fatalf("stat calls = %#v, want %#v", probe.statCalls, []string{"/opt/code"})
	}
}

func TestVSCodeEditorBuildsStructuredLaunchCommand(t *testing.T) {
	tests := []struct {
		name    string
		request codingtool.CommandRequest
		want    codingtool.Command
	}{
		{
			name: "paths with spaces",
			request: codingtool.CommandRequest{
				Config:      codingtool.DefaultConfig(),
				Executable:  "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
				ProjectPath: "/Users/Alex/Work/Client A/API",
				Paths: codingtool.ContextPaths{
					StorageDir: "/Users/Alex/.devctx/contexts/client-a/tools/vscode",
				},
			},
			want: codingtool.Command{
				Executable: "/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
				Arguments: codingtool.Arguments{
					"--user-data-dir", "/Users/Alex/.devctx/contexts/client-a/tools/vscode",
					"--extensions-dir", "/Users/Alex/.devctx/contexts/client-a/tools/vscode/extensions",
					"--new-window",
					"/Users/Alex/Work/Client A/API",
				},
			},
		},
		{
			name: "windows separators",
			request: codingtool.CommandRequest{
				Config:      codingtool.DefaultConfig(),
				Executable:  `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
				ProjectPath: `C:\Users\Alex\Projects\Client A\API`,
				Paths: codingtool.ContextPaths{
					StorageDir: `C:\Users\Alex\.devctx\contexts\client-a\tools\vscode`,
				},
			},
			want: codingtool.Command{
				Executable: `C:\Users\Alex\AppData\Local\Programs\Microsoft VS Code\bin\code.cmd`,
				Arguments: codingtool.Arguments{
					"--user-data-dir", `C:\Users\Alex\.devctx\contexts\client-a\tools\vscode`,
					"--extensions-dir", `C:\Users\Alex\.devctx\contexts\client-a\tools\vscode\extensions`,
					"--new-window",
					`C:\Users\Alex\Projects\Client A\API`,
				},
			},
		},
		{
			name: "non ASCII paths",
			request: codingtool.CommandRequest{
				Config:      codingtool.DefaultConfig(),
				Executable:  "/usr/local/bin/code",
				ProjectPath: "/Users/Alex/équipe/Café Portal",
				Paths: codingtool.ContextPaths{
					StorageDir: "/Users/Alex/.devctx/contexts/equipe/tools/vscode",
				},
			},
			want: codingtool.Command{
				Executable: "/usr/local/bin/code",
				Arguments: codingtool.Arguments{
					"--user-data-dir", "/Users/Alex/.devctx/contexts/equipe/tools/vscode",
					"--extensions-dir", "/Users/Alex/.devctx/contexts/equipe/tools/vscode/extensions",
					"--new-window",
					"/Users/Alex/équipe/Café Portal",
				},
			},
		},
		{
			name: "custom executable",
			request: codingtool.CommandRequest{
				Config: codingtool.Config{
					ExecutableOverride: "/opt/vscode-insiders/bin/code-insiders",
				},
				Executable:  "/opt/vscode-insiders/bin/code-insiders",
				ProjectPath: "/work/app",
				Paths: codingtool.ContextPaths{
					StorageDir: "/home/alex/.devctx/contexts/personal/tools/vscode",
				},
			},
			want: codingtool.Command{
				Executable: "/opt/vscode-insiders/bin/code-insiders",
				Arguments: codingtool.Arguments{
					"--user-data-dir", "/home/alex/.devctx/contexts/personal/tools/vscode",
					"--extensions-dir", "/home/alex/.devctx/contexts/personal/tools/vscode/extensions",
					"--new-window",
					"/work/app",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			editor := codingtool.VSCodeEditor{}
			if tt.name == "windows separators" {
				editor.OperatingSystem = "windows"
			}
			command, err := editor.BuildLaunchCommand(tt.request)
			if err != nil {
				t.Fatalf("build launch command: %v", err)
			}

			if !reflect.DeepEqual(command, tt.want) {
				t.Fatalf("command = %#v, want %#v", command, tt.want)
			}
		})
	}
}

func TestVSCodeEditorBuildLaunchCommandKeepsContextProfilesSeparate(t *testing.T) {
	editor := codingtool.VSCodeEditor{}
	personal, err := editor.BuildLaunchCommand(codingtool.CommandRequest{
		Executable:  "/usr/local/bin/code",
		ProjectPath: "/work/personal-project",
		Paths: codingtool.ContextPaths{
			StorageDir: "/home/alex/.devctx/contexts/personal/tools/vscode",
		},
	})
	if err != nil {
		t.Fatalf("build Personal command: %v", err)
	}
	company, err := editor.BuildLaunchCommand(codingtool.CommandRequest{
		Executable:  "/usr/local/bin/code",
		ProjectPath: "/work/company-project",
		Paths: codingtool.ContextPaths{
			StorageDir: "/home/alex/.devctx/contexts/company/tools/vscode",
		},
	})
	if err != nil {
		t.Fatalf("build Company command: %v", err)
	}

	if personal.Arguments[1] == company.Arguments[1] {
		t.Fatalf("user-data directories must differ: %q", personal.Arguments[1])
	}
	if personal.Arguments[3] == company.Arguments[3] {
		t.Fatalf("extension directories must differ: %q", personal.Arguments[3])
	}
	if personal.Arguments[4] != "--new-window" || company.Arguments[4] != "--new-window" {
		t.Fatalf("commands must force separate windows: personal=%#v company=%#v", personal.Arguments, company.Arguments)
	}
}

func TestVSCodeEditorBuildLaunchCommandRejectsMissingInputs(t *testing.T) {
	validRequest := codingtool.CommandRequest{
		Config:      codingtool.DefaultConfig(),
		Executable:  "/usr/local/bin/code",
		ProjectPath: "/work/client-a/api",
		Paths: codingtool.ContextPaths{
			StorageDir: "/home/alex/.devctx/contexts/personal/tools/vscode",
		},
	}

	tests := []struct {
		name    string
		request codingtool.CommandRequest
		wantErr error
	}{
		{
			name: "missing executable",
			request: codingtool.CommandRequest{
				Config:      validRequest.Config,
				ProjectPath: validRequest.ProjectPath,
				Paths:       validRequest.Paths,
			},
			wantErr: codingtool.ErrMissingExecutable,
		},
		{
			name: "missing project path",
			request: codingtool.CommandRequest{
				Config:     validRequest.Config,
				Executable: validRequest.Executable,
				Paths:      validRequest.Paths,
			},
			wantErr: codingtool.ErrMissingProjectPath,
		},
		{
			name: "missing tool storage directory",
			request: codingtool.CommandRequest{
				Config:      validRequest.Config,
				Executable:  validRequest.Executable,
				ProjectPath: validRequest.ProjectPath,
			},
			wantErr: codingtool.ErrMissingStorageDir,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := codingtool.VSCodeEditor{}.BuildLaunchCommand(tt.request)
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
