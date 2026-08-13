package cli_test

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"devctx/packages/core/cli"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

func TestParseLaunchRequestUsesCurrentDirectoryByDefault(t *testing.T) {
	fixture := newLaunchRequestFixture(t)

	request, err := cli.ParseLaunchRequest(nil, fixture.workingDir, fixture.paths)
	if err != nil {
		t.Fatalf("parse launch request: %v", err)
	}

	assertLaunchRequest(t, request, launcher.LaunchRequest{
		ProjectPath: project.Path(fixture.workingDir),
		Interactive: true,
		Source:      launcher.InvocationSourceCLI,
	})
}

func TestParseLaunchRequestAcceptsExplicitProjectPaths(t *testing.T) {
	fixture := newLaunchRequestFixture(t)
	absoluteProject := fixture.mkdir(t, "projects", "absolute")
	relativeProject := fixture.mkdir(t, "work", "api")
	spaceProject := fixture.mkdir(t, "projects", "app with spaces")
	homeProject := fixture.mkdir(t, "home", "projects", "constructa")

	tests := []struct {
		name string
		args []string
		want project.Path
	}{
		{
			name: "absolute",
			args: []string{absoluteProject},
			want: project.Path(absoluteProject),
		},
		{
			name: "relative",
			args: []string{"../api/"},
			want: project.Path(relativeProject),
		},
		{
			name: "space-containing",
			args: []string{spaceProject},
			want: project.Path(spaceProject),
		},
		{
			name: "home-relative",
			args: []string{"~/projects/constructa"},
			want: project.Path(homeProject),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := cli.ParseLaunchRequest(tt.args, fixture.workingDir, fixture.paths)
			if err != nil {
				t.Fatalf("parse launch request: %v", err)
			}

			assertLaunchRequest(t, request, launcher.LaunchRequest{
				ProjectPath: tt.want,
				Interactive: true,
				Source:      launcher.InvocationSourceCLI,
			})
		})
	}
}

func TestParseLaunchRequestRejectsInvalidProjectPaths(t *testing.T) {
	fixture := newLaunchRequestFixture(t)
	filePath := filepath.Join(fixture.root, "README.md")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("write file fixture: %v", err)
	}

	tests := []struct {
		name string
		args []string
		want error
	}{
		{
			name: "nonexistent",
			args: []string{filepath.Join(fixture.root, "missing")},
			want: project.ErrProjectDirectoryNotFound,
		},
		{
			name: "file",
			args: []string{filePath},
			want: project.ErrProjectPathNotDirectory,
		},
		{
			name: "too many positional paths",
			args: []string{fixture.workingDir, fixture.homeDir},
			want: cli.ErrInvalidCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cli.ParseLaunchRequest(tt.args, fixture.workingDir, fixture.paths)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestParseLaunchRequestAcceptsGenericContextFlag(t *testing.T) {
	fixture := newLaunchRequestFixture(t)
	contextID := devcontext.MustID("personal")

	tests := []struct {
		name string
		args []string
		want launcher.LaunchRequest
	}{
		{
			name: "with explicit path",
			args: []string{"--context", "personal", "."},
			want: launcher.LaunchRequest{
				ProjectPath:      project.Path(fixture.workingDir),
				RequestedContext: &contextID,
				Interactive:      false,
				Source:           launcher.InvocationSourceCLI,
			},
		},
		{
			name: "without explicit path",
			args: []string{"--context", "personal"},
			want: launcher.LaunchRequest{
				ProjectPath:      project.Path(fixture.workingDir),
				RequestedContext: &contextID,
				Interactive:      false,
				Source:           launcher.InvocationSourceCLI,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := cli.ParseLaunchRequest(tt.args, fixture.workingDir, fixture.paths)
			if err != nil {
				t.Fatalf("parse launch request: %v", err)
			}

			assertLaunchRequest(t, request, tt.want)
		})
	}
}

func TestParseLaunchRequestAcceptsPersonalAndCompanyAliases(t *testing.T) {
	fixture := newLaunchRequestFixture(t)

	tests := []struct {
		name        string
		aliasArgs   []string
		genericArgs []string
	}{
		{
			name:        "personal",
			aliasArgs:   []string{"--personal", "."},
			genericArgs: []string{"--context", "personal", "."},
		},
		{
			name:        "company",
			aliasArgs:   []string{"--company", "."},
			genericArgs: []string{"--context", "company", "."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliasRequest, err := cli.ParseLaunchRequest(tt.aliasArgs, fixture.workingDir, fixture.paths)
			if err != nil {
				t.Fatalf("parse alias launch request: %v", err)
			}
			genericRequest, err := cli.ParseLaunchRequest(tt.genericArgs, fixture.workingDir, fixture.paths)
			if err != nil {
				t.Fatalf("parse generic launch request: %v", err)
			}

			assertLaunchRequest(t, aliasRequest, genericRequest)
		})
	}
}

func TestParseLaunchRequestRejectsConflictingContextSelections(t *testing.T) {
	fixture := newLaunchRequestFixture(t)

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "personal and company aliases",
			args: []string{"--personal", "--company", "."},
		},
		{
			name: "alias and generic flag",
			args: []string{"--personal", "--context", "personal", "."},
		},
		{
			name: "repeated generic flag",
			args: []string{"--context", "personal", "--context", "company", "."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cli.ParseLaunchRequest(tt.args, fixture.workingDir, fixture.paths)
			if !errors.Is(err, cli.ErrInvalidCommand) {
				t.Fatalf("error = %v, want %v", err, cli.ErrInvalidCommand)
			}
			if !strings.Contains(err.Error(), "context can only be selected once") {
				t.Fatalf("error = %q, want conflict message", err.Error())
			}
		})
	}
}

func TestParseLaunchRequestRejectsInvalidGenericContextFlagForms(t *testing.T) {
	fixture := newLaunchRequestFixture(t)

	tests := []struct {
		name        string
		args        []string
		want        error
		wantMessage string
	}{
		{
			name:        "invalid context ID",
			args:        []string{"--context", "Personal", "."},
			want:        devcontext.ErrInvalidID,
			wantMessage: "must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:        "missing context ID",
			args:        []string{"--context"},
			want:        cli.ErrInvalidCommand,
			wantMessage: "--context requires a context ID",
		},
		{
			name:        "flag where context ID belongs",
			args:        []string{"--context", "--debug"},
			want:        cli.ErrInvalidCommand,
			wantMessage: "--context requires a context ID",
		},
		{
			name:        "repeated context flag",
			args:        []string{"--context", "personal", "--context", "company", "."},
			want:        cli.ErrInvalidCommand,
			wantMessage: "context can only be selected once",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cli.ParseLaunchRequest(tt.args, fixture.workingDir, fixture.paths)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Fatalf("error = %q, want message containing %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestParseLaunchRequestRejectsNonRootCommands(t *testing.T) {
	fixture := newLaunchRequestFixture(t)

	_, err := cli.ParseLaunchRequest([]string{"context", "list"}, fixture.workingDir, fixture.paths)
	if !errors.Is(err, cli.ErrInvalidCommand) {
		t.Fatalf("error = %v, want %v", err, cli.ErrInvalidCommand)
	}
}

type launchRequestFixture struct {
	root       string
	homeDir    string
	workingDir string
	paths      filesystem.PlatformPaths
}

func newLaunchRequestFixture(t *testing.T) launchRequestFixture {
	t.Helper()

	root := t.TempDir()
	fixture := launchRequestFixture{
		root:    root,
		homeDir: filepath.Join(root, "home"),
	}
	fixture.workingDir = fixture.mkdir(t, "work", "web")
	fixture.paths = filesystem.NewDefaultPlatformPathsWithUserHome(func() (string, error) {
		return fixture.homeDir, nil
	})

	return fixture
}

func (f launchRequestFixture) mkdir(t *testing.T, elements ...string) string {
	t.Helper()

	path := filepath.Join(append([]string{f.root}, elements...)...)
	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("create directory fixture %q: %v", path, err)
	}
	return path
}

func assertLaunchRequest(t *testing.T, got launcher.LaunchRequest, want launcher.LaunchRequest) {
	t.Helper()

	if got.ProjectPath != want.ProjectPath {
		t.Fatalf("project path = %q, want %q", got.ProjectPath, want.ProjectPath)
	}
	if got.Interactive != want.Interactive {
		t.Fatalf("interactive = %v, want %v", got.Interactive, want.Interactive)
	}
	if got.Source != want.Source {
		t.Fatalf("source = %q, want %q", got.Source, want.Source)
	}
	if !reflect.DeepEqual(got.RequestedContext, want.RequestedContext) {
		t.Fatalf("requested context = %#v, want %#v", got.RequestedContext, want.RequestedContext)
	}
}
