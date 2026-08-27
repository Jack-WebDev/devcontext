package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	codingtool "devctx/packages/core/codingtool"
	devcontext "devctx/packages/core/context"
	"devctx/packages/core/environment"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	devlog "devctx/packages/core/logging"
	"devctx/packages/core/project"
	"devctx/packages/core/provider"
	"devctx/packages/core/version"
)

// Result is the rendered outcome of a CLI command.
type Result struct {
	Code   ExitCode
	Stdout string
	Stderr string
}

// Write writes rendered command output to the supplied streams.
func (r Result) Write(stdout io.Writer, stderr io.Writer) error {
	if r.Stdout != "" && stdout != nil {
		if _, err := io.WriteString(stdout, r.Stdout); err != nil {
			return err
		}
	}
	if r.Stderr != "" && stderr != nil {
		if _, err := io.WriteString(stderr, r.Stderr); err != nil {
			return err
		}
	}
	return nil
}

// Runner executes implemented CLI commands against core repositories.
type Runner struct {
	Contexts         devcontext.Repository
	Projects         project.Repository
	WorkingDirectory string
	Paths            filesystem.PlatformPaths
	ProviderRegistry provider.Registry
	ToolRegistry     codingtool.Registry
	// Tool is retained temporarily for callers that have not yet moved to the
	// registry contract. New code must provide ToolRegistry.
	Tool               codingtool.CodingTool
	ProcessLauncher    launcher.ProcessLauncher
	ParentEnvironment  []string
	DetachMode         launcher.DetachMode
	StoragePermissions filesystem.StoragePermissions
	Now                func() time.Time
	Debug              bool
	Logger             devlog.Logger
}

// Run parses and executes one CLI command.
func (r Runner) Run(args []string) Result {
	args, debug := SplitDebugFlag(args)
	if debug {
		r.Debug = true
	}

	command, err := Parse(args)
	if err != nil {
		return r.errorResult(err)
	}

	switch command.Kind {
	case CommandContext:
		return r.runContext(command.Context)
	case CommandProject:
		return r.runProject(command.Project)
	case CommandVersion:
		return successResult(version.Render(version.Current()))
	case CommandRootLaunch:
		return r.runRootLaunch(command.RootLaunch)
	default:
		return r.errorResult(fmt.Errorf("%w: root launch execution is not implemented in this adapter", ErrInvalidCommand))
	}
}

func (r Runner) runRootLaunch(command RootLaunchCommand) Result {
	paths := r.paths()
	request, err := launchRequestFromRootCommand(command, r.WorkingDirectory, paths)
	if err != nil {
		r.recordLaunchEvent(devlog.NewEvent(devlog.EventInput{
			Name:             devlog.LaunchEventNameForError(err),
			Timestamp:        r.now(),
			Err:              err,
			KnownEnvironment: r.parentEnvironment(),
		}))
		return r.errorResult(err)
	}

	builder := launcher.LaunchPlanBuilder{
		Resolver:          launcher.NewResolver(r.Contexts, r.Projects),
		PlatformPaths:     paths,
		ProviderRegistry:  r.providerRegistry(),
		ToolRegistry:      r.toolRegistry(),
		ParentEnvironment: r.parentEnvironment(),
	}
	plan, err := builder.Build(request)
	if err != nil {
		r.recordLaunchEvent(devlog.NewEvent(devlog.EventInput{
			Name:             devlog.LaunchEventNameForError(err),
			Timestamp:        r.now(),
			ProjectPath:      string(request.ProjectPath),
			ContextID:        requestedContextID(request),
			Err:              err,
			KnownEnvironment: r.parentEnvironment(),
		}))
		return r.errorResult(err)
	}

	r.recordLaunchEvent(eventFromPlan(devlog.EventContextResolution, plan, nil, r.now()))
	for range plan.MissingProviderIDs {
		event := eventFromPlan(devlog.EventLaunchProviderMissing, plan, nil, r.now())
		event.ErrorCategory = devlog.ErrorCategoryProvider
		r.recordLaunchEvent(event)
	}

	if err := r.processLauncher().Launch(processRequestFromLaunchPlan(plan, r.detachMode())); err != nil {
		r.recordLaunchEvent(eventFromPlan(devlog.EventLaunchProcessFailure, plan, err, r.now()))
		return r.errorResult(err)
	}

	r.recordLaunchEvent(eventFromPlan(devlog.EventLaunchSucceeded, plan, nil, r.now()))

	output := renderLaunchPlan(plan)
	if r.Debug {
		output += "\n" + renderDebugLaunchPlan(plan)
	}
	return successResult(output)
}

func (r Runner) runContext(command ContextCommand) Result {
	switch command.Subcommand {
	case ContextList:
		contexts, err := r.Contexts.List()
		if err != nil {
			return r.errorResult(err)
		}
		return successResult(renderContextList(contexts))
	case ContextCreate:
		contextID, err := devcontext.NewID(command.ContextID)
		if err != nil {
			return r.errorResult(err)
		}
		ctx, err := devcontext.DefaultContextForIDWithRegistries(contextID, r.now(), r.providerRegistry(), r.toolRegistry())
		if err != nil {
			return r.errorResult(err)
		}
		paths := r.paths()
		contextPaths, err := filesystem.DeriveContextPaths(paths, contextID)
		if err != nil {
			return r.errorResult(err)
		}
		if err := filesystem.CreateContextDirectoryTreeWithProviderRegistryCredentialsAndPermissions(paths, contextPaths, ctx, r.providerRegistry(), nil, r.storagePermissions()); err != nil {
			return r.errorResult(err)
		}
		return successResult(renderContextCreate(ctx))
	default:
		return r.errorResult(fmt.Errorf("%w: context %s is not implemented", ErrInvalidCommand, command.Subcommand))
	}
}

func (r Runner) runProject(command ProjectCommand) Result {
	switch command.Subcommand {
	case ProjectShow:
		lookup, err := r.projectLookup()
		if err != nil {
			return r.errorResult(err)
		}
		return successResult(renderProjectLookup(lookup))
	case ProjectBind:
		contextID, err := devcontext.NewID(command.ContextID)
		if err != nil {
			return r.errorResult(err)
		}

		binding, err := r.Projects.Bind(".", project.Path(r.WorkingDirectory), contextID, r.Contexts, r.now())
		if err != nil {
			return r.errorResult(err)
		}
		return successResult(renderProjectBind(binding))
	case ProjectUnbind:
		if _, err := r.projectLookup(); err != nil {
			return r.errorResult(err)
		}
		result, err := r.Projects.Unbind(".", project.Path(r.WorkingDirectory))
		if err != nil {
			return r.errorResult(err)
		}
		return successResult(renderProjectUnbind(result))
	default:
		return r.errorResult(fmt.Errorf("%w: project %s is not implemented", ErrInvalidCommand, command.Subcommand))
	}
}

func (r Runner) projectLookup() (project.BindingLookup, error) {
	lookup, err := r.Projects.LookupWithContextValidation(".", project.Path(r.WorkingDirectory), r.Contexts)
	if err != nil {
		return project.BindingLookup{}, err
	}
	if err := project.ValidateProjectDirectory(lookup.ProjectPath); err != nil {
		return project.BindingLookup{}, err
	}
	return lookup, nil
}

func (r Runner) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func (r Runner) paths() filesystem.PlatformPaths {
	if r.Paths != nil {
		return r.Paths
	}
	return filesystem.NewDefaultPlatformPaths()
}

func (r Runner) providerRegistry() provider.Registry {
	if !r.ProviderRegistry.IsZero() {
		return r.ProviderRegistry
	}
	return provider.BuiltInRegistry()
}

func (r Runner) toolRegistry() codingtool.Registry {
	if !r.ToolRegistry.IsZero() {
		return r.ToolRegistry
	}
	if r.Tool != nil {
		return codingtool.MustNewRegistry([]codingtool.RegisteredTool{{Integration: r.Tool, DisplayName: string(r.Tool.ID())}}, r.Tool.ID())
	}
	return codingtool.BuiltInRegistry()
}

func (r Runner) processLauncher() launcher.ProcessLauncher {
	if r.ProcessLauncher != nil {
		return r.ProcessLauncher
	}
	return launcher.NativeProcessLauncher{}
}

func (r Runner) parentEnvironment() []string {
	if r.ParentEnvironment != nil {
		return append([]string(nil), r.ParentEnvironment...)
	}
	return os.Environ()
}

func (r Runner) detachMode() launcher.DetachMode {
	if r.DetachMode != "" {
		return r.DetachMode
	}
	return launcher.DetachModeDetached
}

func (r Runner) storagePermissions() filesystem.StoragePermissions {
	if r.StoragePermissions != nil {
		return r.StoragePermissions
	}
	return filesystem.NewDefaultStoragePermissions()
}

func (r Runner) logger() devlog.Logger {
	if r.Logger != nil {
		return r.Logger
	}
	return devlog.NoopLogger{}
}

func (r Runner) recordLaunchEvent(event devlog.Event) {
	_ = r.logger().Record(event)
}

func processRequestFromLaunchPlan(plan launcher.LaunchPlan, detachMode launcher.DetachMode) launcher.ProcessRequest {
	return launcher.ProcessRequest{
		Executable:       plan.Executable,
		Arguments:        append(launcher.Arguments(nil), plan.Arguments...),
		Environment:      cloneLaunchEnvironment(plan.Environment),
		WorkingDirectory: plan.WorkingDirectory,
		DetachMode:       detachMode,
	}
}

func cloneLaunchEnvironment(environment launcher.Environment) launcher.Environment {
	cloned := make(launcher.Environment, len(environment))
	for key, value := range environment {
		cloned[key] = value
	}
	return cloned
}

func successResult(output string) Result {
	return Result{
		Code:   ExitSuccess,
		Stdout: output,
	}
}

func (r Runner) errorResult(err error) Result {
	return Result{
		Code:   ExitCodeForError(err),
		Stderr: RenderError(err, r.Debug),
	}
}

// SplitDebugFlag removes --debug from command arguments and reports whether it
// was present.
func SplitDebugFlag(args []string) ([]string, bool) {
	filtered := make([]string, 0, len(args))
	debug := false
	for _, arg := range args {
		if arg == debugFlag {
			debug = true
			continue
		}
		filtered = append(filtered, arg)
	}
	return filtered, debug
}

func renderContextList(contexts []devcontext.Context) string {
	if len(contexts) == 0 {
		return "No contexts configured.\n"
	}

	nameWidth := len("NAME")
	for _, ctx := range contexts {
		if len(ctx.Name) > nameWidth {
			nameWidth = len(ctx.Name)
		}
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "%-*s  %s\n", nameWidth, "NAME", "ID")
	for _, ctx := range contexts {
		fmt.Fprintf(&builder, "%-*s  %s\n", nameWidth, ctx.Name, ctx.ID.String())
	}
	return builder.String()
}

func renderContextCreate(ctx devcontext.Context) string {
	return fmt.Sprintf("Context:\n%s\n\nStatus:\ncreated\n", ctx.ID.String())
}

func renderProjectLookup(lookup project.BindingLookup) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Project:\n%s\n\n", lookup.ProjectPath)

	switch {
	case lookup.Dangling:
		fmt.Fprintf(&builder, "Context:\nmissing: %s\n\n", lookup.MissingContextID.String())
		fmt.Fprintf(&builder, "Binding:\ndangling\n\n")
		fmt.Fprintf(&builder, "Recovery:\n%s\n", lookup.Recovery)
	case lookup.Bound:
		fmt.Fprintf(&builder, "Context:\n%s\n", lookup.Binding.ContextID.String())
	default:
		fmt.Fprintf(&builder, "Context:\nunbound\n")
	}

	return builder.String()
}

func renderProjectBind(binding project.Binding) string {
	return fmt.Sprintf("Project:\n%s\n\nContext:\n%s\n\nStatus:\nbound\n", binding.ProjectPath, binding.ContextID.String())
}

func renderProjectUnbind(result project.UnbindResult) string {
	if result.Removed {
		return fmt.Sprintf("Project:\n%s\n\nRemoved context:\n%s\n\nStatus:\nunbound\n", result.ProjectPath, result.Binding.ContextID.String())
	}
	return fmt.Sprintf("Project:\n%s\n\nContext:\nunbound\n\nStatus:\nunchanged\n", result.ProjectPath)
}

func renderLaunchPlan(plan launcher.LaunchPlan) string {
	return fmt.Sprintf("Project:\n%s\n\nContext:\n%s\n\nStatus:\nlaunched\n", plan.ProjectPath, plan.Context.ID.String())
}

func renderDebugLaunchPlan(plan launcher.LaunchPlan) string {
	var builder strings.Builder
	builder.WriteString("Debug:\n")
	fmt.Fprintf(&builder, "resolution_source: %s\n", plan.ResolutionSource)
	fmt.Fprintf(&builder, "editor_id: %s\n", plan.Tool.Type)
	fmt.Fprintf(&builder, "editor_executable: %s\n", plan.Executable)
	builder.WriteString("context_directories:\n")
	fmt.Fprintf(&builder, "  root: %s\n", plan.ContextPaths.RootDir)
	builder.WriteString("  providers:\n")
	for _, providerID := range sortedProviderPathIDs(plan.ContextPaths.ProviderStorageDirs) {
		fmt.Fprintf(&builder, "    %s: %s\n", providerID, plan.ContextPaths.ProviderStorageDirs[providerID])
	}
	builder.WriteString("arguments:\n")
	for i, argument := range plan.Arguments {
		fmt.Fprintf(&builder, "  %d: %s\n", i, argument)
	}
	builder.WriteString("environment:\n")
	for _, entry := range redactedEnvironment(plan.Environment) {
		fmt.Fprintf(&builder, "  %s\n", entry)
	}
	return builder.String()
}

func sortedProviderPathIDs(paths map[provider.ID]string) []provider.ID {
	ids := make([]provider.ID, 0, len(paths))
	for providerID := range paths {
		ids = append(ids, providerID)
	}
	sort.Slice(ids, func(i int, j int) bool {
		return ids[i] < ids[j]
	})
	return ids
}

func redactedEnvironment(variables launcher.Environment) []string {
	keys := make([]string, 0, len(variables))
	for key := range variables {
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		value := variables[key]
		if environment.IsSensitiveKey(key) {
			value = environment.RedactedValue
		}
		entries = append(entries, key+"="+value)
	}
	return entries
}

func eventFromPlan(name devlog.EventName, plan launcher.LaunchPlan, err error, timestamp time.Time) devlog.Event {
	return devlog.NewEvent(devlog.EventInput{
		Name:             name,
		Timestamp:        timestamp,
		ProjectPath:      string(plan.ProjectPath),
		ContextID:        plan.Context.ID.String(),
		ToolID:           string(plan.Tool.Type),
		ResolutionSource: string(plan.ResolutionSource),
		Err:              err,
		KnownEnvironment: plan.Environment.Environ(),
	})
}

func requestedContextID(request launcher.LaunchRequest) string {
	if request.RequestedContext == nil {
		return ""
	}
	return request.RequestedContext.String()
}
