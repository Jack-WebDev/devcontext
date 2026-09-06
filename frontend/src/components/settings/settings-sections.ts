export type SupportedSetting =
	| "closeAfterLaunch"
	| "launchVerification"
	| "rememberProjects";

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
] as const;
