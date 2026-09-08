const privacyStatements = [
	{
		title: "Stored on this device",
		description:
			"Dev Context stores its configuration, contexts, remembered project mappings, activity records, and logs in its local Dev Context home.",
	},
	{
		title: "Project paths",
		description:
			"Dev Context stores project paths for remembered contexts, recent launches, active workspaces, and local activity. Forgetting a project removes its remembered context, recent launch, and activity records. Dev Context never changes or deletes your project folder.",
	},
	{
		title: "Credential handling",
		description:
			"Normal launches use context-owned integration storage and never sync credentials. When you explicitly choose to import a provider session, Dev Context copies the provider's local credential file into that context and reads only allowlisted identity metadata; it never displays, logs, uploads, or exports credential contents.",
	},
	{
		title: "Portable exports",
		description:
			"Context exports contain selected metadata and settings only. They do not include credentials, project paths, internal IDs, or runtime state.",
	},
] as const;

export { privacyStatements };
