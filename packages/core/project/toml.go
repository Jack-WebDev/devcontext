package project

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	devcontext "devctx/packages/core/context"
)

var (
	// ErrInvalidProjectBindings identifies structurally invalid project binding data.
	ErrInvalidProjectBindings = errors.New("invalid project bindings")

	// ErrDuplicateProjectBinding identifies multiple bindings for the same
	// project path.
	ErrDuplicateProjectBinding = errors.New("duplicate project binding")
)

type projectBindingsTOML struct {
	Projects []projectBindingTOML `toml:"projects"`
}

type projectBindingTOML struct {
	Path      *string    `toml:"path"`
	Context   *string    `toml:"context"`
	CreatedAt *time.Time `toml:"created_at"`
}

// DecodeProjectBindingsTOML decodes projects.toml bytes into project bindings.
func DecodeProjectBindingsTOML(data []byte) ([]Binding, error) {
	var raw projectBindingsTOML
	metadata, err := toml.Decode(string(data), &raw)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidProjectBindings, err)
	}

	if undecoded := metadata.Undecoded(); len(undecoded) > 0 {
		return nil, fmt.Errorf("%w: unsupported field %q", ErrInvalidProjectBindings, undecoded[0].String())
	}

	return bindingsFromTOML(raw)
}

// EncodeProjectBindingsTOML encodes project bindings in deterministic TOML.
func EncodeProjectBindingsTOML(bindings []Binding) ([]byte, error) {
	if err := validateProjectBindings(bindings); err != nil {
		return nil, err
	}

	ordered := append([]Binding(nil), bindings...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].ProjectPath < ordered[j].ProjectPath
	})

	var builder strings.Builder
	for i, binding := range ordered {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("[[projects]]\n")
		writeProjectStringValue(&builder, "path", string(binding.ProjectPath))
		writeProjectStringValue(&builder, "context", binding.ContextID.String())
		writeProjectTimeValue(&builder, "created_at", binding.CreatedAt)
	}

	return []byte(builder.String()), nil
}

func bindingsFromTOML(raw projectBindingsTOML) ([]Binding, error) {
	bindings := make([]Binding, 0, len(raw.Projects))
	seen := make(map[Path]struct{}, len(raw.Projects))

	for index, rawProject := range raw.Projects {
		binding, err := bindingFromTOML(index, rawProject)
		if err != nil {
			return nil, err
		}
		if _, exists := seen[binding.ProjectPath]; exists {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateProjectBinding, binding.ProjectPath)
		}
		seen[binding.ProjectPath] = struct{}{}
		bindings = append(bindings, binding)
	}

	return bindings, nil
}

func bindingFromTOML(index int, raw projectBindingTOML) (Binding, error) {
	if raw.Path == nil {
		return Binding{}, fmt.Errorf("%w: missing projects[%d].path", ErrInvalidProjectBindings, index)
	}
	if *raw.Path == "" {
		return Binding{}, fmt.Errorf("%w: projects[%d].path cannot be empty", ErrInvalidProjectBindings, index)
	}

	if raw.Context == nil {
		return Binding{}, fmt.Errorf("%w: missing projects[%d].context", ErrInvalidProjectBindings, index)
	}
	contextID, err := devcontext.NewID(*raw.Context)
	if err != nil {
		return Binding{}, fmt.Errorf("%w: projects[%d].context: %w", ErrInvalidProjectBindings, index, err)
	}

	if raw.CreatedAt == nil {
		return Binding{}, fmt.Errorf("%w: missing projects[%d].created_at", ErrInvalidProjectBindings, index)
	}
	if raw.CreatedAt.IsZero() {
		return Binding{}, fmt.Errorf("%w: projects[%d].created_at cannot be zero", ErrInvalidProjectBindings, index)
	}

	return Binding{
		ProjectPath: Path(*raw.Path),
		ContextID:   contextID,
		CreatedAt:   raw.CreatedAt.UTC(),
	}, nil
}

func validateProjectBindings(bindings []Binding) error {
	seen := make(map[Path]struct{}, len(bindings))
	for _, binding := range bindings {
		if binding.ProjectPath == "" {
			return fmt.Errorf("%w: project path cannot be empty", ErrInvalidProjectBindings)
		}
		if binding.ContextID.String() == "" {
			return fmt.Errorf("%w: %w: context cannot be empty", ErrInvalidProjectBindings, devcontext.ErrInvalidID)
		}
		if binding.CreatedAt.IsZero() {
			return fmt.Errorf("%w: created_at cannot be zero", ErrInvalidProjectBindings)
		}
		if _, exists := seen[binding.ProjectPath]; exists {
			return fmt.Errorf("%w: %q", ErrDuplicateProjectBinding, binding.ProjectPath)
		}
		seen[binding.ProjectPath] = struct{}{}
	}
	return nil
}

func writeProjectStringValue(builder *strings.Builder, key string, value string) {
	builder.WriteString(key)
	builder.WriteString(" = ")
	builder.WriteString(strconv.Quote(value))
	builder.WriteString("\n")
}

func writeProjectTimeValue(builder *strings.Builder, key string, value time.Time) {
	builder.WriteString(key)
	builder.WriteString(" = ")
	builder.WriteString(value.UTC().Format(time.RFC3339Nano))
	builder.WriteString("\n")
}
