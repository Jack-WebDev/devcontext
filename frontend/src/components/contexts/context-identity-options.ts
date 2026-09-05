const contextIconOptions = [
	{ id: "user", label: "Person", symbol: "●" },
	{ id: "building", label: "Building", symbol: "▦" },
	{ id: "users", label: "People", symbol: "◉" },
	{ id: "code", label: "Code", symbol: "⌘" },
	{ id: "heart", label: "Heart", symbol: "♥" },
	{ id: "sparkles", label: "Sparkles", symbol: "✦" },
] as const;

const contextAccentOptions = [
	{ id: "sage", label: "Sage", swatchClassName: "bg-accent-personal" },
	{
		id: "slate-blue",
		label: "Slate blue",
		swatchClassName: "bg-accent-company",
	},
	{ id: "amber", label: "Amber", swatchClassName: "bg-accent-freelance" },
	{ id: "custom", label: "Orchid", swatchClassName: "bg-accent-custom" },
] as const;

function contextIconOption(icon: string | undefined) {
	return contextIconOptions.find((option) => option.id === icon);
}

function contextAccentOption(accent: string | undefined) {
	return contextAccentOptions.find((option) => option.id === accent);
}

export {
	contextAccentOption,
	contextAccentOptions,
	contextIconOption,
	contextIconOptions,
};
