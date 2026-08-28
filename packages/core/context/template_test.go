package context

import "testing"

func TestTemplatesIncludeSupportedContextKinds(t *testing.T) {
	want := []string{"personal", "company", "freelance", "client", "open-source", "custom"}
	templates := Templates()
	if len(templates) != len(want) {
		t.Fatalf("template count = %d, want %d", len(templates), len(want))
	}
	for index, id := range want {
		if templates[index].ID != id {
			t.Fatalf("template %d = %q, want %q", index, templates[index].ID, id)
		}
		if templates[index].Name == "" || templates[index].Accent == "" {
			t.Fatalf("template %q has incomplete defaults: %#v", id, templates[index])
		}
	}
}

func TestTemplateByID(t *testing.T) {
	if template, ok := TemplateByID("freelance"); !ok || template.Name != "Freelance" {
		t.Fatalf("freelance template = %#v, found = %t", template, ok)
	}
	if _, ok := TemplateByID("unknown"); ok {
		t.Fatal("unknown template found")
	}
}
