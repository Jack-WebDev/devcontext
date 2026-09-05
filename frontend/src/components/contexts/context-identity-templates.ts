import type { CreateContextRequest } from "../../lib/devctx-api";

interface ContextIdentityTemplate {
	id: "personal" | "work" | "client" | "open-source" | "custom";
	name: string;
	description: string;
	icon: string;
	accent: string;
	backendTemplateId: string;
}

const contextIdentityTemplates: ContextIdentityTemplate[] = [
	{
		id: "personal",
		name: "Personal",
		description: "Personal development work.",
		icon: "user",
		accent: "sage",
		backendTemplateId: "personal",
	},
	{
		id: "work",
		name: "Work",
		description: "Development work for your employer.",
		icon: "building",
		accent: "slate-blue",
		backendTemplateId: "company",
	},
	{
		id: "client",
		name: "Client",
		description: "A dedicated client identity.",
		icon: "users",
		accent: "amber",
		backendTemplateId: "client",
	},
	{
		id: "open-source",
		name: "Open Source",
		description: "Open source contributions.",
		icon: "code",
		accent: "sage",
		backendTemplateId: "open-source",
	},
	{
		id: "custom",
		name: "Custom",
		description: "A custom development identity.",
		icon: "",
		accent: "custom",
		backendTemplateId: "custom",
	},
];

// The request intentionally omits contextId. The application derives its
// filesystem-safe internal ID from this editable display name.
function draftFromContextIdentityTemplate(
	template: ContextIdentityTemplate,
): CreateContextRequest {
	return {
		templateId: template.backendTemplateId,
		name: template.name,
		description: template.description,
		icon: template.icon,
		accent: template.accent,
	};
}

export type { ContextIdentityTemplate };
export { contextIdentityTemplates, draftFromContextIdentityTemplate };
