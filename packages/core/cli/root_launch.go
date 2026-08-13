package cli

import (
	"fmt"
	"strings"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/filesystem"
	"devctx/packages/core/launcher"
	"devctx/packages/core/project"
)

const contextFlag = "--context"

type rootLaunchArguments struct {
	projectPath      string
	requestedContext *devcontext.ID
}

// ParseLaunchRequest converts root launch command arguments into an application
// launch request. args excludes the executable name. workingDirectory is the
// process working directory at invocation time.
func ParseLaunchRequest(args []string, workingDirectory string, paths filesystem.PlatformPaths) (launcher.LaunchRequest, error) {
	command, err := Parse(args)
	if err != nil {
		return launcher.LaunchRequest{}, err
	}
	if command.Kind != CommandRootLaunch {
		return launcher.LaunchRequest{}, fmt.Errorf("%w: launch requests use the root command, not %s", ErrInvalidCommand, command.Kind)
	}

	return launchRequestFromRootCommand(command.RootLaunch, workingDirectory, paths)
}

func launchRequestFromRootCommand(command RootLaunchCommand, workingDirectory string, paths filesystem.PlatformPaths) (launcher.LaunchRequest, error) {
	args, err := parseRootLaunchArguments(command.Arguments)
	if err != nil {
		return launcher.LaunchRequest{}, err
	}

	projectPathInput := args.projectPath
	if projectPathInput == "" {
		projectPathInput = "."
	}

	canonicalPath, err := project.CanonicalizePath(paths, projectPathInput, project.Path(workingDirectory))
	if err != nil {
		return launcher.LaunchRequest{}, err
	}
	if err := project.ValidateProjectDirectory(canonicalPath); err != nil {
		return launcher.LaunchRequest{}, err
	}

	return launcher.LaunchRequest{
		ProjectPath:      canonicalPath,
		RequestedContext: args.requestedContext,
		Interactive:      args.requestedContext == nil,
		Source:           launcher.InvocationSourceCLI,
	}, nil
}

func parseRootLaunchArguments(args []string) (rootLaunchArguments, error) {
	var parsed rootLaunchArguments

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == contextFlag:
			if parsed.requestedContext != nil {
				return rootLaunchArguments{}, fmt.Errorf("%w: %s can only be provided once", ErrInvalidCommand, contextFlag)
			}
			if i+1 >= len(args) {
				return rootLaunchArguments{}, fmt.Errorf("%w: %s requires a context ID", ErrInvalidCommand, contextFlag)
			}

			value := args[i+1]
			if strings.HasPrefix(value, "-") {
				return rootLaunchArguments{}, fmt.Errorf("%w: %s requires a context ID", ErrInvalidCommand, contextFlag)
			}

			contextID, err := devcontext.NewID(value)
			if err != nil {
				return rootLaunchArguments{}, fmt.Errorf("%w: %s: %w", ErrInvalidCommand, contextFlag, err)
			}
			parsed.requestedContext = &contextID
			i++
		case strings.HasPrefix(arg, "-"):
			return rootLaunchArguments{}, fmt.Errorf("%w: unknown root option %q", ErrInvalidCommand, arg)
		default:
			if parsed.projectPath != "" {
				return rootLaunchArguments{}, fmt.Errorf("%w: root launch accepts at most one project path", ErrInvalidCommand)
			}
			parsed.projectPath = arg
		}
	}

	return parsed, nil
}
