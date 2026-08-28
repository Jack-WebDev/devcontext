package context

// Template is a safe set of suggested metadata for a new development
// identity. Templates never contain credentials, provider settings, or coding
// tool settings.
type Template struct {
	ID          string
	Name        string
	Description string
	Icon        string
	Accent      string
}

// Templates returns the built-in context templates in their display order.
func Templates() []Template {
	return []Template{
		{ID: "personal", Name: "Personal", Description: "Personal development work.", Icon: "user", Accent: "sage"},
		{ID: "company", Name: "Company", Description: "Work for your employer.", Icon: "building", Accent: "slate-blue"},
		{ID: "freelance", Name: "Freelance", Description: "Independent client work.", Icon: "briefcase", Accent: "amber"},
		{ID: "client", Name: "Client", Description: "A dedicated client identity.", Icon: "users", Accent: "amber"},
		{ID: "open-source", Name: "Open Source", Description: "Open source contributions.", Icon: "code", Accent: "sage"},
		{ID: "custom", Name: "Custom", Description: "A custom development identity.", Icon: "", Accent: "custom"},
	}
}

// TemplateByID returns one built-in template.
func TemplateByID(id string) (Template, bool) {
	for _, template := range Templates() {
		if template.ID == id {
			return template, true
		}
	}
	return Template{}, false
}
