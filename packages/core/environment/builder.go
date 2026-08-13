package environment

import (
	"sort"
	"strings"

	"devctx/packages/core/provider"
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
