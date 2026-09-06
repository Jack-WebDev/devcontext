import { useTheme } from "next-themes";

const appearanceOptions = [
	{ value: "system", label: "System" },
	{ value: "light", label: "Light" },
	{ value: "dark", label: "Dark" },
] as const;

function AppearanceSettings() {
	const { theme, setTheme } = useTheme();

	return (
		<section
			className="border-b border-border pb-6"
			aria-labelledby="settings-appearance"
		>
			<h3 id="settings-appearance" className="font-semibold">
				Appearance
			</h3>
			<p className="mt-1 text-sm text-muted-foreground">
				Choose how Dev Context matches your display.
			</p>
			<label className="mt-4 flex items-start justify-between gap-6">
				<span>
					<span className="block text-sm font-medium">Color theme</span>
					<span className="mt-1 block text-sm text-muted-foreground">
						System follows your device setting.
					</span>
				</span>
				<select
					className="rounded-md border border-input bg-background px-3 py-2 text-sm"
					value={theme ?? "system"}
					onChange={(event) => setTheme(event.target.value)}
					aria-label="Color theme"
				>
					{appearanceOptions.map((option) => (
						<option key={option.value} value={option.value}>
							{option.label}
						</option>
					))}
				</select>
			</label>
		</section>
	);
}

export { AppearanceSettings, appearanceOptions };
