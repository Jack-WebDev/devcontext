export type SupportedSetting =
	| "closeAfterLaunch"
	| "launchVerification"
	| "rememberProjects";

export type SafetySetting = "warnOnContextMismatch";

export const settingsSections = [
	{
		title: "Launching",
		description: "Choose how Dev Context handles a successful launch.",
		fields: [
			"launchVerification",
			"rememberProjects",
			"closeAfterLaunch",
		] as const,
	},
	{
		title: "Safety",
		description: "Keep context overrides deliberate.",
		fields: ["warnOnContextMismatch"] as const,
	},
] as const;
