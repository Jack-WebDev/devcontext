package codingtool

import "fmt"

// RegisteredTool describes one registered editor integration and its user-facing name.
// The name belongs in the registry so callers do not need integration-specific
// display-name switches.
type RegisteredTool struct {
	Integration  CodingTool
	DisplayName  string
	Capabilities []Capability
}

// Capability identifies an optional editor feature supported by an
// integration. Concrete capabilities will be introduced with the tool
// implementations that need them.
type Capability string

// Registry owns the editor integrations available to Dev Context.
//
// It preserves registration order for presentation, resolves persisted editor
// IDs, and identifies the default integration for newly created contexts.
type Registry struct {
	tools       []RegisteredTool
	toolsByID   map[ID]RegisteredTool
	defaultTool ID
}

// NewRegistry creates a registry from an ordered list of editor tools.
func NewRegistry(tools []RegisteredTool, defaultTool ID) (Registry, error) {
	if defaultTool == "" {
		return Registry{}, fmt.Errorf("editor registry has an empty default tool ID")
	}

	ordered := make([]RegisteredTool, 0, len(tools))
	byID := make(map[ID]RegisteredTool, len(tools))
	for i, tool := range tools {
		if tool.Integration == nil {
			return Registry{}, fmt.Errorf("editor registry contains nil tool at index %d", i)
		}
		id := tool.Integration.ID()
		if id == "" {
			return Registry{}, fmt.Errorf("editor registry contains tool with empty ID at index %d", i)
		}
		if _, exists := byID[id]; exists {
			return Registry{}, fmt.Errorf("editor registry contains duplicate tool ID %q", id)
		}
		if tool.DisplayName == "" {
			return Registry{}, fmt.Errorf("editor registry contains tool %q with empty display name", id)
		}
		tool.Capabilities = append([]Capability(nil), tool.Capabilities...)
		ordered = append(ordered, tool)
		byID[id] = tool
	}
	if _, ok := byID[defaultTool]; !ok {
		return Registry{}, fmt.Errorf("editor registry default tool %q is not registered", defaultTool)
	}

	return Registry{tools: ordered, toolsByID: byID, defaultTool: defaultTool}, nil
}

// MustNewRegistry creates a registry and panics when its definition is invalid.
func MustNewRegistry(tools []RegisteredTool, defaultTool ID) Registry {
	registry, err := NewRegistry(tools, defaultTool)
	if err != nil {
		panic(err)
	}
	return registry
}

// BuiltInRegistry returns the currently available built-in editor tools.
func BuiltInRegistry() Registry {
	return MustNewRegistry([]RegisteredTool{{Integration: VSCodeEditor{}, DisplayName: "VS Code"}}, VSCodeID)
}

// IsZero reports whether the registry has not been initialized.
func (r Registry) IsZero() bool {
	return r.tools == nil && r.toolsByID == nil && r.defaultTool == ""
}

// All returns tools in stable presentation order.
func (r Registry) All() []RegisteredTool {
	tools := make([]RegisteredTool, len(r.tools))
	for i, tool := range r.tools {
		tools[i] = tool
		tools[i].Capabilities = append([]Capability(nil), tool.Capabilities...)
	}
	return tools
}

// Get returns the registered integration for id.
func (r Registry) Get(id ID) (CodingTool, bool) {
	tool, ok := r.toolsByID[id]
	return tool.Integration, ok
}

// DefaultID returns the ID selected for newly created contexts.
func (r Registry) DefaultID() ID {
	return r.defaultTool
}

// DisplayName returns the registered display name, or the raw ID when the
// tool is not registered.
func (r Registry) DisplayName(id ID) string {
	tool, ok := r.toolsByID[id]
	if !ok {
		return string(id)
	}
	return tool.DisplayName
}

// HasCapability reports whether a registered tool supports capability.
func (r Registry) HasCapability(id ID, capability Capability) bool {
	tool, ok := r.toolsByID[id]
	if !ok {
		return false
	}
	for _, supported := range tool.Capabilities {
		if supported == capability {
			return true
		}
	}
	return false
}
