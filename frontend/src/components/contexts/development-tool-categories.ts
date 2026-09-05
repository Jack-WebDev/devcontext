const developmentToolCategories = [
	{
		id: "coding",
		name: "Coding",
		description: "Editors and coding environments used to open projects.",
	},
	{
		id: "ai",
		name: "AI",
		description: "AI assistants used as part of your development work.",
	},
	{
		id: "version-control",
		name: "Version control",
		description: "Tools that manage local source history.",
	},
	{
		id: "source-hosting",
		name: "Source hosting",
		description: "Services that host and review source code.",
	},
	{
		id: "cloud-registries",
		name: "Cloud & registries",
		description: "Cloud services, package registries, and deployment tools.",
	},
	{
		id: "other",
		name: "Other",
		description:
			"Other development integrations that do not fit a category above.",
	},
] as const;

type DevelopmentToolCategory = (typeof developmentToolCategories)[number];
type DevelopmentToolCategoryID = DevelopmentToolCategory["id"];

export type { DevelopmentToolCategory, DevelopmentToolCategoryID };
export { developmentToolCategories };
