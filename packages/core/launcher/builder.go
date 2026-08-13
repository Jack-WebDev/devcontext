package launcher

import (
	"errors"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/editor"
	"devctx/packages/core/environment"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
)

var (
	// ErrMissingContextResolver identifies a launch-plan builder without context
	// resolution.
	ErrMissingContextResolver = errors.New("missing context resolver")

	// ErrMissingPlatformPaths identifies a launch-plan builder without platform
	// path resolution.
	ErrMissingPlatformPaths = errors.New("missing platform paths")

	// ErrMissingEditor identifies a launch-plan builder without an editor
	// implementation.
	ErrMissingEditor = errors.New("missing editor")

	// ErrLaunchSelectionRequired identifies a request that cannot produce a full
	// launch plan until the user selects a context.
	ErrLaunchSelectionRequired = errors.New("launch requires context selection")
)

// ContextResolver resolves one launch request to a selected context or a
// required-selection state.
type ContextResolver interface {
	Resolve(LaunchRequest) (ResolutionResult, error)
}

// LaunchPlanBuilder builds a complete launch plan without starting a process.
type LaunchPlanBuilder struct {
	Resolver          ContextResolver
	PlatformPaths     filesystem.PlatformPaths
	Providers         []provider.Provider
	Editor            editor.Editor
	ParentEnvironment []string
}

// Build validates the project, resolves the context, builds provider
// environment, and constructs the editor command for one launch.
func (b LaunchPlanBuilder) Build(request LaunchRequest) (LaunchPlan, error) {
	if b.Resolver == nil {
		return LaunchPlan{}, ErrMissingContextResolver
	}
	if b.PlatformPaths == nil {
		return LaunchPlan{}, ErrMissingPlatformPaths
	}
	if b.Editor == nil {
		return LaunchPlan{}, ErrMissingEditor
	}
	if err := project.ValidateProjectDirectory(request.ProjectPath); err != nil {
		return LaunchPlan{}, err
	}

	resolution, err := b.Resolver.Resolve(request)
	if err != nil {
		return LaunchPlan{}, err
	}
	if resolution.SelectionRequired || resolution.Context == nil {
		return LaunchPlan{}, ErrLaunchSelectionRequired
	}

	contextPaths, err := filesystem.DeriveContextPaths(b.PlatformPaths, resolution.Context.ID)
	if err != nil {
		return LaunchPlan{}, err
	}

	contributions, err := b.providerContributions(*resolution.Context, contextPaths)
	if err != nil {
		return LaunchPlan{}, err
	}
	variables, err := environment.BuildForContext(b.ParentEnvironment, resolution.Context.ID, contributions...)
	if err != nil {
		return LaunchPlan{}, err
	}

	executable, err := b.Editor.DetectExecutable(resolution.Context.Editor)
	if err != nil {
		return LaunchPlan{}, err
	}
	command, err := b.Editor.BuildLaunchCommand(editor.CommandRequest{
		Config:      resolution.Context.Editor,
		Executable:  executable,
		ProjectPath: string(request.ProjectPath),
		Paths: editor.ContextPaths{
			RootDir:     contextPaths.RootDir,
			DataDir:     contextPaths.VSCodeDir,
			UserDataDir: contextPaths.VSCodeUserDataDir,
		},
	})
	if err != nil {
		return LaunchPlan{}, err
	}

	return LaunchPlan{
		ProjectPath:      request.ProjectPath,
		Context:          *resolution.Context,
		Editor:           resolution.Context.Editor,
		Executable:       Executable(command.Executable),
		Arguments:        launchArguments(command.Arguments),
		WorkingDirectory: WorkingDirectory(request.ProjectPath),
		Environment:      launchEnvironment(variables),
		Warnings:         resolution.Warnings,
		ResolutionSource: resolution.Source,
	}, nil
}

func (b LaunchPlanBuilder) providerContributions(ctxContext devcontext.Context, paths filesystem.ContextPaths) ([]provider.EnvironmentContribution, error) {
	contributions := make([]provider.EnvironmentContribution, 0, len(b.Providers))
	for _, integration := range b.Providers {
		if integration == nil {
			continue
		}
		config, ok := ctxContext.Providers[integration.ID()]
		if !ok || !config.Enabled {
			continue
		}

		contribution, err := integration.BuildEnvironment(provider.RuntimeContext{
			ContextID: ctxContext.ID.String(),
			Config:    config,
			Paths: provider.ContextPaths{
				RootDir:           paths.RootDir,
				ClaudeDir:         paths.ClaudeDir,
				CodexDir:          paths.CodexDir,
				VSCodeDir:         paths.VSCodeDir,
				VSCodeUserDataDir: paths.VSCodeUserDataDir,
			},
		})
		if err != nil {
			return nil, err
		}
		contributions = append(contributions, contribution)
	}
	return contributions, nil
}

func launchArguments(arguments editor.Arguments) Arguments {
	values := make(Arguments, len(arguments))
	for i, value := range arguments {
		values[i] = string(value)
	}
	return values
}

func launchEnvironment(variables environment.Variables) Environment {
	values := make(Environment, len(variables))
	for key, value := range variables {
		values[key] = value
	}
	return values
}
