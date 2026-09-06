const privacyStatements = [
	{
		title: "Stored on this device",
		description:
			"Dev Context stores its configuration, contexts, remembered project mappings, activity records, and logs in its local Dev Context home.",
	},
	{
		title: "Project paths",
		description:
			"A project path is saved only when you choose to remember its context. Dev Context never changes or deletes your project folder.",
	},
	{
		title: "Credentials stay with integrations",
		description:
			"Provider credentials are handled in context-owned integration storage. Dev Context does not store passwords, tokens, or cloud accounts itself.",
	},
	{
		title: "Portable exports",
		description:
			"Context exports contain selected metadata and settings only. They do not include credentials, project paths, internal IDs, or runtime state.",
	},
] as const;

export { privacyStatements };
