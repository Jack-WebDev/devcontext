package codingtool_test

import (
	"strings"
	"testing"

	codingtool "devctx/packages/core/codingtool"
)

func TestRegistryPreservesOrderAndResolvesToolMetadata(t *testing.T) {
	first := registryFakeEditor{id: "first"}
	second := registryFakeEditor{id: "second"}
	registry, err := codingtool.NewRegistry([]codingtool.RegisteredTool{
		{Integration: first, DisplayName: "First RegisteredTool", Capabilities: []codingtool.Capability{"profiles"}},
		{Integration: second, DisplayName: "Second RegisteredTool"},
	}, second.ID())
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}

	tools := registry.All()
	if len(tools) != 2 || tools[0].Integration.ID() != first.ID() || tools[1].Integration.ID() != second.ID() {
		t.Fatalf("tools = %#v, want registered order", tools)
	}
	if registry.DefaultID() != second.ID() {
		t.Fatalf("default ID = %q, want %q", registry.DefaultID(), second.ID())
	}
	resolved, ok := registry.Get(first.ID())
	if !ok || resolved.ID() != first.ID() {
		t.Fatalf("resolved tool = %#v, found = %t", resolved, ok)
	}
	if registry.DisplayName(second.ID()) != "Second RegisteredTool" {
		t.Fatalf("display name = %q, want Second RegisteredTool", registry.DisplayName(second.ID()))
	}
	if registry.DisplayName("missing") != "missing" {
		t.Fatalf("unknown display name = %q, want raw ID", registry.DisplayName("missing"))
	}
	if !registry.HasCapability(first.ID(), "profiles") || registry.HasCapability(second.ID(), "profiles") {
		t.Fatalf("capability lookup did not match registered tool support")
	}
}

func TestRegistryRejectsInvalidDefinitions(t *testing.T) {
	valid := registryFakeEditor{id: "valid"}
	tests := []struct {
		name  string
		tools []codingtool.RegisteredTool
		want  string
	}{
		{name: "nil tool", tools: []codingtool.RegisteredTool{{DisplayName: "Missing"}}, want: "nil tool"},
		{name: "duplicate ID", tools: []codingtool.RegisteredTool{{Integration: valid, DisplayName: "One"}, {Integration: valid, DisplayName: "Two"}}, want: "duplicate"},
		{name: "missing display name", tools: []codingtool.RegisteredTool{{Integration: valid}}, want: "display name"},
		{name: "unknown default", tools: []codingtool.RegisteredTool{{Integration: valid, DisplayName: "Valid"}}, want: "default tool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := codingtool.NewRegistry(tt.tools, "missing")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

type registryFakeEditor struct{ id codingtool.ID }

func (e registryFakeEditor) ID() codingtool.ID { return e.id }

func (registryFakeEditor) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	return "/usr/local/bin/fake", nil
}

func (registryFakeEditor) BuildLaunchCommand(request codingtool.CommandRequest) (codingtool.Command, error) {
	return codingtool.Command{Executable: request.Executable}, nil
}
