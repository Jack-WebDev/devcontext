package application

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	codingtool "devctx/packages/core/codingtool"
	"devctx/packages/core/config"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

// Dependencies contains the core collaborators used by application use cases.
type Dependencies struct {
	Contexts         devcontext.Repository
	Projects         project.Repository
	RecentProjects   project.RecentRepository
	Paths            filesystem.PlatformPaths
	ProviderRegistry provider.Registry
	ToolRegistry     codingtool.Registry
	// Tool is retained temporarily for callers that have not yet moved to the
	// registry contract. New code must provide ToolRegistry.
	Tool               codingtool.CodingTool
	ProcessLauncher    launcher.ProcessLauncher
	StoragePermissions filesystem.StoragePermissions
	ParentEnvironment  []string
	WorkingDirectory   string
	DetachMode         launcher.DetachMode
	Now                func() time.Time
	Logger             devlog.Logger
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
		Contexts:           devcontext.NewRepository(layout.ContextsDir),
		Projects:           project.NewRepository(filepath.Join(layout.HomeDir, "projects.toml"), paths),
		RecentProjects:     project.NewRecentRepository(filepath.Join(layout.HomeDir, "recents.toml")),
		Paths:              paths,
		ProviderRegistry:   provider.BuiltInRegistry(),
		ToolRegistry:       codingtool.BuiltInRegistry(),
		ProcessLauncher:    launcher.NativeProcessLauncher{},
		StoragePermissions: filesystem.NewDefaultStoragePermissions(),
		ParentEnvironment:  options.ParentEnvironment,
		WorkingDirectory:   workingDirectory,
		DetachMode:         launcher.DetachModeDetached,
		Now:                options.Now,
		Logger:             devlog.NewLocalLogger(layout.LogsDir, filesystem.NewDefaultStoragePermissions(), options.Now),
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
		ProviderRegistry:  s.dependencies.ProviderRegistry,
		ToolRegistry:      s.dependencies.ToolRegistry,
		ParentEnvironment: s.dependencies.ParentEnvironment,
	}
}

func (s *Service) processLauncher() launcher.ProcessLauncher {
	return s.dependencies.ProcessLauncher
}

func (s *Service) now() time.Time {
	return s.dependencies.Now()
}

func (s *Service) logger() devlog.Logger {
	return s.dependencies.Logger
}

func normalizeDependencies(dependencies Dependencies) Dependencies {
	if dependencies.Paths == nil {
		dependencies.Paths = filesystem.NewDefaultPlatformPaths()
	}
	if dependencies.ProviderRegistry.IsZero() {
		dependencies.ProviderRegistry = provider.BuiltInRegistry()
	}
	if dependencies.ToolRegistry.IsZero() {
		if dependencies.Tool != nil {
			dependencies.ToolRegistry = codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: dependencies.Tool, DisplayName: string(dependencies.Tool.ID())}}, dependencies.Tool.ID())
		} else {
			dependencies.ToolRegistry = codingtool.BuiltInRegistry()
		}
	}
	if dependencies.ProcessLauncher == nil {
		dependencies.ProcessLauncher = launcher.NativeProcessLauncher{}
	}
	if dependencies.StoragePermissions == nil {
		dependencies.StoragePermissions = filesystem.NewDefaultStoragePermissions()
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
	if dependencies.Logger == nil {
		dependencies.Logger = devlog.NoopLogger{}
	}
	return dependencies
}
