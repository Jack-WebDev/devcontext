package cli

import (
	"fmt"
	"io"
	"strings"
	"time"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/project"
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
	Now              func() time.Time
}

// Run parses and executes one CLI command.
func (r Runner) Run(args []string) Result {
	command, err := Parse(args)
	if err != nil {
		return errorResult(err)
	}

	switch command.Kind {
	case CommandContext:
		return r.runContext(command.Context)
	case CommandProject:
		return r.runProject(command.Project)
	default:
		return errorResult(fmt.Errorf("%w: root launch execution is not implemented in this adapter", ErrInvalidCommand))
	}
}

func (r Runner) runContext(command ContextCommand) Result {
	switch command.Subcommand {
	case ContextList:
		contexts, err := r.Contexts.List()
		if err != nil {
			return errorResult(err)
		}
		return successResult(renderContextList(contexts))
	default:
		return errorResult(fmt.Errorf("%w: context %s is not implemented", ErrInvalidCommand, command.Subcommand))
	}
}

func (r Runner) runProject(command ProjectCommand) Result {
	switch command.Subcommand {
	case ProjectShow:
		lookup, err := r.projectLookup()
		if err != nil {
			return errorResult(err)
		}
		return successResult(renderProjectLookup(lookup))
	case ProjectBind:
		contextID, err := devcontext.NewID(command.ContextID)
		if err != nil {
			return errorResult(err)
		}

		binding, err := r.Projects.Bind(".", project.Path(r.WorkingDirectory), contextID, r.Contexts, r.now())
		if err != nil {
			return errorResult(err)
		}
		return successResult(renderProjectBind(binding))
	case ProjectUnbind:
		if _, err := r.projectLookup(); err != nil {
			return errorResult(err)
		}
		result, err := r.Projects.Unbind(".", project.Path(r.WorkingDirectory))
		if err != nil {
			return errorResult(err)
		}
		return successResult(renderProjectUnbind(result))
	default:
		return errorResult(fmt.Errorf("%w: project %s is not implemented", ErrInvalidCommand, command.Subcommand))
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

func successResult(output string) Result {
	return Result{
		Code:   ExitSuccess,
		Stdout: output,
	}
}

func errorResult(err error) Result {
	return Result{
		Code:   ExitCodeForError(err),
		Stderr: err.Error() + "\n",
	}
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
