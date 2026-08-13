package cli

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidCommand identifies command input with the wrong shape.
	ErrInvalidCommand = errors.New("invalid CLI command")

	// ErrUnknownCommand identifies an unrecognized command family or subcommand.
	ErrUnknownCommand = errors.New("unknown CLI command")
)

// CommandKind identifies the top-level command family.
type CommandKind string

const (
	// CommandRootLaunch identifies the root launch command, such as
	// "devctx", "devctx .", or "devctx --context personal .".
	CommandRootLaunch CommandKind = "root_launch"

	// CommandContext identifies context management commands.
	CommandContext CommandKind = "context"

	// CommandProject identifies project binding commands.
	CommandProject CommandKind = "project"

	// CommandVersion identifies "devctx --version".
	CommandVersion CommandKind = "version"
)

// ContextSubcommand identifies a context command.
type ContextSubcommand string

const (
	// ContextList identifies "devctx context list".
	ContextList ContextSubcommand = "list"

	// ContextCreate identifies the future-ready "devctx context create <id>"
	// command shape documented in the PRD.
	ContextCreate ContextSubcommand = "create"
)

// ProjectSubcommand identifies a project command.
type ProjectSubcommand string

const (
	// ProjectShow identifies "devctx project show".
	ProjectShow ProjectSubcommand = "show"

	// ProjectBind identifies "devctx project bind <context-id>".
	ProjectBind ProjectSubcommand = "bind"

	// ProjectUnbind identifies "devctx project unbind".
	ProjectUnbind ProjectSubcommand = "unbind"
)

// Command is one parsed CLI request.
type Command struct {
	Kind CommandKind

	RootLaunch RootLaunchCommand
	Context    ContextCommand
	Project    ProjectCommand
}

// RootLaunchCommand stores root launch arguments exactly as supplied after the
// executable name. ParseLaunchRequest interprets supported launch arguments
// when filesystem context is available.
type RootLaunchCommand struct {
	Arguments []string
}

// ContextCommand stores the parsed context subcommand and its raw operand.
type ContextCommand struct {
	Subcommand ContextSubcommand
	ContextID  string
}

// ProjectCommand stores the parsed project subcommand and its raw operand.
type ProjectCommand struct {
	Subcommand ProjectSubcommand
	ContextID  string
}

// Parse converts command-line arguments, excluding the executable name, into a
// typed command request.
func Parse(args []string) (Command, error) {
	if len(args) == 0 {
		return rootLaunchCommand(nil), nil
	}

	switch args[0] {
	case "--version":
		if len(args) != 1 {
			return Command{}, fmt.Errorf("%w: --version accepts no arguments", ErrInvalidCommand)
		}
		return Command{Kind: CommandVersion}, nil
	case string(CommandContext):
		return parseContextCommand(args[1:])
	case string(CommandProject):
		return parseProjectCommand(args[1:])
	default:
		return rootLaunchCommand(args), nil
	}
}

func parseContextCommand(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, fmt.Errorf("%w: context command requires a subcommand: list or create", ErrInvalidCommand)
	}

	switch args[0] {
	case string(ContextList):
		if len(args) != 1 {
			return Command{}, fmt.Errorf("%w: context list accepts no arguments", ErrInvalidCommand)
		}
		return Command{
			Kind: CommandContext,
			Context: ContextCommand{
				Subcommand: ContextList,
			},
		}, nil
	case string(ContextCreate):
		if len(args) != 2 {
			return Command{}, fmt.Errorf("%w: context create requires exactly one context ID", ErrInvalidCommand)
		}
		return Command{
			Kind: CommandContext,
			Context: ContextCommand{
				Subcommand: ContextCreate,
				ContextID:  args[1],
			},
		}, nil
	default:
		return Command{}, fmt.Errorf("%w: unknown context subcommand %q; expected list or create", ErrUnknownCommand, args[0])
	}
}

func parseProjectCommand(args []string) (Command, error) {
	if len(args) == 0 {
		return Command{}, fmt.Errorf("%w: project command requires a subcommand: show, bind, or unbind", ErrInvalidCommand)
	}

	switch args[0] {
	case string(ProjectShow):
		if len(args) != 1 {
			return Command{}, fmt.Errorf("%w: project show accepts no arguments", ErrInvalidCommand)
		}
		return Command{
			Kind: CommandProject,
			Project: ProjectCommand{
				Subcommand: ProjectShow,
			},
		}, nil
	case string(ProjectBind):
		if len(args) != 2 {
			return Command{}, fmt.Errorf("%w: project bind requires exactly one context ID", ErrInvalidCommand)
		}
		return Command{
			Kind: CommandProject,
			Project: ProjectCommand{
				Subcommand: ProjectBind,
				ContextID:  args[1],
			},
		}, nil
	case string(ProjectUnbind):
		if len(args) != 1 {
			return Command{}, fmt.Errorf("%w: project unbind accepts no arguments", ErrInvalidCommand)
		}
		return Command{
			Kind: CommandProject,
			Project: ProjectCommand{
				Subcommand: ProjectUnbind,
			},
		}, nil
	default:
		return Command{}, fmt.Errorf("%w: unknown project subcommand %q; expected show, bind, or unbind", ErrUnknownCommand, args[0])
	}
}

func rootLaunchCommand(args []string) Command {
	return Command{
		Kind: CommandRootLaunch,
		RootLaunch: RootLaunchCommand{
			Arguments: append([]string(nil), args...),
		},
	}
}
