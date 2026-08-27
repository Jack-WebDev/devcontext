package editor_test

import (
	"strings"
	"testing"

	"devctx/packages/core/editor"
)

func TestRegistryPreservesOrderAndResolvesToolMetadata(t *testing.T) {
	first := registryFakeEditor{id: "first"}
	second := registryFakeEditor{id: "second"}
	registry, err := editor.NewRegistry([]editor.Tool{
		{Integration: first, DisplayName: "First Tool", Capabilities: []editor.Capability{"profiles"}},
		{Integration: second, DisplayName: "Second Tool"},
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
	if registry.DisplayName(second.ID()) != "Second Tool" {
		t.Fatalf("display name = %q, want Second Tool", registry.DisplayName(second.ID()))
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
		tools []editor.Tool
		want  string
	}{
		{name: "nil tool", tools: []editor.Tool{{DisplayName: "Missing"}}, want: "nil tool"},
		{name: "duplicate ID", tools: []editor.Tool{{Integration: valid, DisplayName: "One"}, {Integration: valid, DisplayName: "Two"}}, want: "duplicate"},
		{name: "missing display name", tools: []editor.Tool{{Integration: valid}}, want: "display name"},
		{name: "unknown default", tools: []editor.Tool{{Integration: valid, DisplayName: "Valid"}}, want: "default tool"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := editor.NewRegistry(tt.tools, "missing")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

type registryFakeEditor struct{ id editor.ID }

func (e registryFakeEditor) ID() editor.ID { return e.id }

func (registryFakeEditor) DetectExecutable(editor.Config) (editor.Executable, error) {
	return "/usr/local/bin/fake", nil
}

func (registryFakeEditor) BuildLaunchCommand(request editor.CommandRequest) (editor.Command, error) {
	return editor.Command{Executable: request.Executable}, nil
}
