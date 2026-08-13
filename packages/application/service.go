package application

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

// Dependencies contains the core collaborators used by application use cases.
type Dependencies struct {
	Contexts          devcontext.Repository
	Projects          project.Repository
	Paths             filesystem.PlatformPaths
	Providers         []provider.Provider
	Editor            editor.Editor
	ProcessLauncher   launcher.ProcessLauncher
	ParentEnvironment []string
	WorkingDirectory  string
	DetachMode        launcher.DetachMode
	Now               func() time.Time
}

// DefaultOptions contains host-provided values needed to construct the default
// application service.
type DefaultOptions struct {
	Paths             filesystem.PlatformPaths
	ParentEnvironment []string
	WorkingDirectory  string
	Now               func() time.Time
}

// Service coordinates product use cases for framework adapters.
type Service struct {
	dependencies Dependencies
}

// NewService creates an application service backed by local Dev Context state.
func NewService() (*Service, error) {
	return NewDefaultService(DefaultOptions{})
}

// NewDefaultService creates an application service backed by local Dev Context
// state using injectable host values for tests.
func NewDefaultService(options DefaultOptions) (*Service, error) {
	paths := options.Paths
	if paths == nil {
		paths = filesystem.NewDefaultPlatformPaths()
	}

	layout, err := config.InitializeDevContextHome(paths)
	if err != nil {
		return nil, fmt.Errorf("initialize Dev Context home: %w", err)
	}

	workingDirectory := options.WorkingDirectory
	if workingDirectory == "" {
		workingDirectory, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
	}

	return NewServiceWithDependencies(Dependencies{
		Contexts:          devcontext.NewRepository(layout.ContextsDir),
		Projects:          project.NewRepository(filepath.Join(layout.HomeDir, "projects.toml"), paths),
		Paths:             paths,
		Providers:         []provider.Provider{provider.ClaudeProvider{}, provider.CodexProvider{}},
		Editor:            editor.VSCodeEditor{},
		ProcessLauncher:   launcher.NativeProcessLauncher{},
		ParentEnvironment: options.ParentEnvironment,
		WorkingDirectory:  workingDirectory,
		DetachMode:        launcher.DetachModeDetached,
		Now:               options.Now,
	}), nil
}

// NewServiceWithDependencies creates an application service with supplied
// collaborators. Tests and non-Wails adapters can use this to exercise
// application behavior without starting Wails.
func NewServiceWithDependencies(dependencies Dependencies) *Service {
	return &Service{
		dependencies: normalizeDependencies(dependencies),
	}
}

func (s *Service) launchPlanBuilder() launcher.LaunchPlanBuilder {
	return launcher.LaunchPlanBuilder{
		Resolver:          launcher.NewResolver(s.dependencies.Contexts, s.dependencies.Projects),
		PlatformPaths:     s.dependencies.Paths,
		Providers:         s.dependencies.Providers,
		Editor:            s.dependencies.Editor,
		ParentEnvironment: s.dependencies.ParentEnvironment,
	}
}

func (s *Service) processLauncher() launcher.ProcessLauncher {
	return s.dependencies.ProcessLauncher
}

func (s *Service) now() time.Time {
	return s.dependencies.Now()
}

func normalizeDependencies(dependencies Dependencies) Dependencies {
	if dependencies.Paths == nil {
		dependencies.Paths = filesystem.NewDefaultPlatformPaths()
	}
	if dependencies.Providers == nil {
		dependencies.Providers = []provider.Provider{provider.ClaudeProvider{}, provider.CodexProvider{}}
	}
	if dependencies.Editor == nil {
		dependencies.Editor = editor.VSCodeEditor{}
	}
	if dependencies.ProcessLauncher == nil {
		dependencies.ProcessLauncher = launcher.NativeProcessLauncher{}
	}
	if dependencies.ParentEnvironment == nil {
		dependencies.ParentEnvironment = os.Environ()
	} else {
		dependencies.ParentEnvironment = append([]string(nil), dependencies.ParentEnvironment...)
	}
	if dependencies.DetachMode == "" {
		dependencies.DetachMode = launcher.DetachModeDetached
	}
	if dependencies.Now == nil {
		dependencies.Now = time.Now
	}
	return dependencies
}
