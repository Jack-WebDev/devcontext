import type { DisplayError, SettingsState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Label } from "../ui/label.js";
import { Switch } from "../ui/switch.js";
import { PageHeader } from "../ui/page-header.js";
import { AppearanceSettings } from "./AppearanceSettings.js";
import { AboutSettings } from "./AboutSettings.js";
import { AdvancedSettings } from "./AdvancedSettings.js";
import { PrivacySettings } from "./PrivacySettings.js";
import {
	settingsSections,
	type SafetySetting,
	type SupportedSetting,
} from "./settings-sections.js";

interface SettingsViewProps {
	settings: SettingsState;
	pending: boolean;
	error?: DisplayError;
	onChange: (settings: SettingsState) => void;
	onReplayOnboarding: () => void;
	onOpenDiagnostics: () => void;
}

const labels: Record<
	SupportedSetting | SafetySetting,
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
			"Allow the selected context to be remembered for a project when you choose it.",
	},
	warnOnContextMismatch: {
		label: "Confirm temporary context overrides",
		description:
			"Ask before launching a project with a context other than its remembered context.",
	},
};

function SettingsView({
	settings,
	pending,
	error,
	onChange,
	onReplayOnboarding,
	onOpenDiagnostics,
}: SettingsViewProps) {
	return (
		<section
			className="page-content page-section-stack max-w-none"
			aria-labelledby="settings-heading"
		>
			<div className="page-reading-column space-y-8">
				<PageHeader
					id="settings-heading"
					eyebrow="Preferences"
					title="Settings"
				/>
				{error ? (
					<p role="alert" className="text-sm text-destructive">
						{error.message}
					</p>
				) : null}
				<div className="collection-surface divide-y divide-border/50">
					<section className="p-6" aria-labelledby="settings-general">
						<h3 id="settings-general" className="font-semibold">
							General
						</h3>
						<p className="mt-1 text-sm text-muted-foreground">
							Revisit the introduction to development contexts and local
							isolation.
						</p>
						<OnboardingReplayAction
							disabled={pending}
							onReplay={onReplayOnboarding}
						/>
					</section>
					{settingsSections.map((section) => (
						<section
							key={section.title}
							className="p-6"
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
						</section>
					))}
					<AppearanceSettings />
					<PrivacySettings />
					<AdvancedSettings onOpenDiagnostics={onOpenDiagnostics} />
					<AboutSettings />
				</div>
			</div>
		</section>
	);
}

function SettingToggle({
	field,
	settings,
	disabled,
	onChange,
}: {
	field: SupportedSetting | SafetySetting;
	settings: SettingsState;
	disabled: boolean;
	onChange: (settings: SettingsState) => void;
}) {
	const presentation = labels[field];
	return (
		<Label className="flex items-start justify-between gap-6 rounded-xl px-3 py-2.5 transition-colors hover:bg-muted/45">
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
		<div className="mt-5 flex items-start justify-between gap-6 rounded-xl bg-muted/35 p-4">
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
