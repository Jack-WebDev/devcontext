package environment

import (
	"errors"
	"sort"
	"strings"

	devcontext "devctx/packages/core/context"
	"devctx/packages/core/provider"
)

const (
	// ActiveContextEnvVar identifies the selected Dev Context to launched tools.
	ActiveContextEnvVar = "DEVCTX_CONTEXT"
)

var (
	// ErrMissingContextID identifies attempts to build a context environment
	// without a resolved context ID.
	ErrMissingContextID = errors.New("missing active context ID")
)

// Variables stores process environment values by variable name.
type Variables map[string]string

// FromParent copies parent process environment entries into a key map.
//
// When the same key appears more than once, the later entry wins. This matches
// the override behavior used when provider contributions are applied.
func FromParent(parent []string) Variables {
	variables := make(Variables, len(parent))
	for _, entry := range parent {
		key, value, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		variables[key] = value
	}
	return variables
}

// Build copies the parent environment and applies context-specific provider
// contributions.
func Build(parent []string, contributions ...provider.EnvironmentContribution) Variables {
	variables := FromParent(parent)
	for _, contribution := range contributions {
		variables.Apply(contribution)
	}
	return variables
}

// BuildForContext copies the parent environment, applies provider
// contributions, and marks the selected Dev Context.
func BuildForContext(parent []string, contextID devcontext.ID, contributions ...provider.EnvironmentContribution) (Variables, error) {
	if contextID.String() == "" {
		return nil, ErrMissingContextID
	}

	variables := Build(parent, contributions...)
	variables[ActiveContextEnvVar] = contextID.String()
	return variables, nil
}

// Apply overlays one provider environment contribution onto the variables.
func (v Variables) Apply(contribution provider.EnvironmentContribution) {
	for key, value := range contribution {
		if key == "" {
			continue
		}
		v[key] = value
	}
}

// Environ returns a deterministic KEY=value slice.
func (v Variables) Environ() []string {
	keys := make([]string, 0, len(v))
	for key := range v {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, key+"="+v[key])
	}
	return entries
}
