export type SupportedSetting = "closeAfterLaunch" | "launchVerification";

export const settingsSections = [
	{
		title: "Launching",
		description: "Choose how Dev Context handles a successful launch.",
		fields: ["launchVerification", "closeAfterLaunch"] as const,
	},
] as const;
