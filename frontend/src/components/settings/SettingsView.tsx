import type { DisplayError, SettingsState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Label } from "../ui/label";
import { Switch } from "../ui/switch.js";

interface SettingsViewProps {
	settings: SettingsState;
	pending: boolean;
	error?: DisplayError;
	onChange: (settings: SettingsState) => void;
	onReplayOnboarding: () => void;
}

const sections = [
	{
		title: "General",
		description: "Choose how Dev Context prepares launches.",
		fields: ["launchVerification", "closeAfterLaunch"] as const,
	},
	{
		title: "Projects",
		description: "Control project-context suggestions.",
		fields: ["rememberProjects"] as const,
	},
	{
		title: "Application",
		description: "Control background application behavior.",
		fields: ["trayEnabled"] as const,
	},
	{
		title: "Advanced",
		description:
			"Advanced preferences will appear here as capabilities are added.",
		fields: [] as const,
	},
	{
		title: "About",
		description:
			"Dev Context keeps coding-tool and provider state isolated per context.",
		fields: [] as const,
	},
];

const labels: Record<
	keyof SettingsState,
	{ label: string; description: string }
> = {
	closeAfterLaunch: {
		label: "Close after launch",
		description: "Close the selector after a successful launch.",
	},
	launchVerification: {
		label: "Show launch verification",
		description:
			"Show verification progress while Dev Context prepares a launch.",
	},
	rememberProjects: {
		label: "Remember project contexts",
		description:
			"Offer the last selected context as a suggestion for a project.",
	},
	trayEnabled: {
		label: "Enable system tray",
		description:
			"Keep Dev Context available from the system tray when supported.",
	},
};

function SettingsView({
	settings,
	pending,
	error,
	onChange,
	onReplayOnboarding,
}: SettingsViewProps) {
	return (
		<section className="max-w-3xl space-y-8" aria-labelledby="settings-heading">
			<div>
				<p className="text-sm text-muted-foreground">Settings</p>
				<h2 id="settings-heading" className="text-2xl font-semibold">
					Settings
				</h2>
			</div>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error.message}
				</p>
			) : null}
			{sections.map((section) => (
				<section
					key={section.title}
					className="border-b border-border pb-6"
					aria-labelledby={`settings-${section.title.toLowerCase()}`}
				>
					<h3
						id={`settings-${section.title.toLowerCase()}`}
						className="font-semibold"
					>
						{section.title}
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						{section.description}
					</p>
					{section.fields.length > 0 ? (
						<div className="mt-4 space-y-4">
							{section.fields.map((field) => (
								<SettingToggle
									key={field}
									field={field}
									settings={settings}
									disabled={pending}
									onChange={onChange}
								/>
							))}
						</div>
					) : null}
					{section.title === "General" ? (
						<OnboardingReplayAction
							disabled={pending}
							onReplay={onReplayOnboarding}
						/>
					) : null}
				</section>
			))}
		</section>
	);
}

function SettingToggle({
	field,
	settings,
	disabled,
	onChange,
}: {
	field: keyof SettingsState;
	settings: SettingsState;
	disabled: boolean;
	onChange: (settings: SettingsState) => void;
}) {
	const presentation = labels[field];
	return (
		<Label className="flex items-start justify-between gap-6">
			<span>
				<span className="block text-sm font-medium">{presentation.label}</span>
				<span className="mt-1 block text-sm text-muted-foreground">
					{presentation.description}
				</span>
			</span>
			<Switch
				checked={settings[field]}
				disabled={disabled}
				aria-label={presentation.label}
				onCheckedChange={(checked) =>
					onChange({ ...settings, [field]: checked })
				}
			/>
		</Label>
	);
}

function OnboardingReplayAction({
	disabled,
	onReplay,
}: {
	disabled: boolean;
	onReplay: () => void;
}) {
	return (
		<div className="mt-6 flex items-start justify-between gap-6 border-t border-border pt-4">
			<span>
				<span className="block text-sm font-medium">Review onboarding</span>
				<span className="mt-1 block text-sm text-muted-foreground">
					Show the introduction to development contexts and local isolation
					again.
				</span>
			</span>
			<Button
				type="button"
				variant="outline"
				disabled={disabled}
				onClick={onReplay}
			>
				Show onboarding
			</Button>
		</div>
	);
}

export { SettingsView };
