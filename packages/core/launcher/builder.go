package launcher

import (
	"errors"
	"fmt"
	"sort"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
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

	// ErrMissingTool identifies a launch-plan builder without a coding-tool
	// implementation.
	ErrMissingTool = errors.New("missing coding tool")

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
	Resolver         ContextResolver
	PlatformPaths    filesystem.PlatformPaths
	ProviderRegistry provider.Registry
	ToolRegistry     codingtool.Registry
	// Tool is retained temporarily for callers that have not yet moved to the
	// registry contract. New code must provide ToolRegistry.
	Tool              codingtool.CodingTool
	ParentEnvironment []string
}

// Build validates the project, resolves the context, builds provider
// environment, and constructs the coding-tool command for one launch.
func (b LaunchPlanBuilder) Build(request LaunchRequest) (LaunchPlan, error) {
	if b.Resolver == nil {
		return LaunchPlan{}, ErrMissingContextResolver
	}
	if b.PlatformPaths == nil {
		return LaunchPlan{}, ErrMissingPlatformPaths
	}
	if b.ToolRegistry.IsZero() && b.Tool == nil {
		return LaunchPlan{}, ErrMissingTool
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
	providerRegistry := b.providerRegistry()
	toolRegistry := b.toolRegistry()
	if err := filesystem.ValidateContextDirectoryTreeWithRegistries(contextPaths, *resolution.Context, providerRegistry, toolRegistry); err != nil {
		return LaunchPlan{}, err
	}
	contextPaths = contextPaths.WithProviderStorageDirs(registeredEnabledProviderIDs(*resolution.Context, providerRegistry))
	toolID := resolution.Context.Tool.DefaultTool
	toolConfig := resolution.Context.Tool.ConfigFor(toolID)
	contextPaths = contextPaths.WithToolStorageDirs([]codingtool.ID{toolID})

	contributions, missingProviderIDs, err := b.providerContributions(*resolution.Context, contextPaths, providerRegistry)
	if err != nil {
		return LaunchPlan{}, err
	}
	variables, err := environment.BuildForContext(b.ParentEnvironment, resolution.Context.ID, contributions...)
	if err != nil {
		return LaunchPlan{}, err
	}

	registeredTool, ok := toolRegistry.Lookup(toolID)
	if !ok {
		return LaunchPlan{}, fmt.Errorf("%w: %s", ErrMissingTool, toolID)
	}
	executable, err := registeredTool.Integration.DetectExecutable(toolConfig)
	if err != nil {
		return LaunchPlan{}, err
	}
	command, err := registeredTool.Integration.BuildLaunchCommand(codingtool.CommandRequest{
		Config:      toolConfig,
		Executable:  executable,
		ProjectPath: string(request.ProjectPath),
		Paths: codingtool.ContextPaths{
			RootDir:    contextPaths.RootDir,
			StorageDir: contextPaths.ToolStorageDir(toolID),
		},
	})
	if err != nil {
		return LaunchPlan{}, err
	}

	return LaunchPlan{
		ProjectPath: request.ProjectPath,
		Context:     *resolution.Context,
		Tool: Tool{
			ID:          toolID,
			DisplayName: registeredTool.DisplayName,
		},
		Executable:         Executable(command.Executable),
		Arguments:          launchArguments(command.Arguments),
		WorkingDirectory:   WorkingDirectory(request.ProjectPath),
		Environment:        launchEnvironment(variables),
		ContextPaths:       contextPaths,
		Warnings:           resolution.Warnings,
		ResolutionSource:   resolution.Source,
		MissingProviderIDs: missingProviderIDs,
	}, nil
}

func (b LaunchPlanBuilder) providerContributions(ctxContext devcontext.Context, paths filesystem.ContextPaths, registry provider.Registry) ([]provider.EnvironmentContribution, []provider.ID, error) {
	providers := registry.All()
	knownProviderIDs := make(map[provider.ID]struct{}, len(providers))
	contributions := make([]provider.EnvironmentContribution, 0, len(providers))
	for _, integration := range providers {
		providerID := integration.ID()
		knownProviderIDs[providerID] = struct{}{}

		config, ok := ctxContext.Providers[providerID]
		if !ok || !config.Enabled {
			continue
		}
		contribution, err := integration.BuildEnvironment(provider.RuntimeContext{
			ContextID: ctxContext.ID.String(),
			Config:    config,
			Paths: provider.ContextPaths{
				RootDir:    paths.RootDir,
				StorageDir: paths.ProviderStorageDir(providerID),
			},
		})
		if err != nil {
			return nil, nil, err
		}
		contributions = append(contributions, contribution)
	}

	missingProviderIDs := make([]provider.ID, 0)
	for providerID, config := range ctxContext.Providers {
		if !config.Enabled {
			continue
		}
		if _, ok := knownProviderIDs[providerID]; !ok {
			missingProviderIDs = append(missingProviderIDs, providerID)
		}
	}
	sort.Slice(missingProviderIDs, func(i int, j int) bool {
		return missingProviderIDs[i] < missingProviderIDs[j]
	})
	if len(missingProviderIDs) == 0 {
		missingProviderIDs = nil
	}
	return contributions, missingProviderIDs, nil
}

func (b LaunchPlanBuilder) providerRegistry() provider.Registry {
	if b.ProviderRegistry.IsZero() {
		return provider.BuiltInRegistry()
	}
	return b.ProviderRegistry
}

func (b LaunchPlanBuilder) toolRegistry() codingtool.Registry {
	if !b.ToolRegistry.IsZero() {
		return b.ToolRegistry
	}
	return codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: b.Tool, DisplayName: string(b.Tool.ID())}}, b.Tool.ID())
}

func registeredEnabledProviderIDs(ctx devcontext.Context, registry provider.Registry) []provider.ID {
	ids := make([]provider.ID, 0, len(ctx.Providers))
	for _, integration := range registry.All() {
		providerID := integration.ID()
		if config, ok := ctx.Providers[providerID]; ok && config.Enabled {
			ids = append(ids, providerID)
		}
	}
	return ids
}

func launchArguments(arguments codingtool.Arguments) Arguments {
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
