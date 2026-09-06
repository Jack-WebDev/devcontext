import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const sourceRoot = new URL("../src/", import.meta.url);

async function source(path) {
	return readFile(new URL(path, sourceRoot), "utf8");
}

function contrastRatio(first, second) {
	const luminance = (value) => {
		const channels = value
			.slice(1)
			.match(/.{2}/g)
			.map((channel) => Number.parseInt(channel, 16) / 255)
			.map((channel) =>
				channel <= 0.04045
					? channel / 12.92
					: ((channel + 0.055) / 1.055) ** 2.4,
			);
		return 0.2126 * channels[0] + 0.7152 * channels[1] + 0.0722 * channels[2];
	};
	const [lighter, darker] = [luminance(first), luminance(second)].sort(
		(a, b) => b - a,
	);
	return (lighter + 0.05) / (darker + 0.05);
}

test("accessibility: primary keyboard paths expose visible focus and conventional controls", async () => {
	const [styles, shell, contextChoices, contextCard, button, input] =
		await Promise.all([
			source("style.css"),
			source("components/shell/AppShell.tsx"),
			source("components/selector/ContextChoiceList.tsx"),
			source("components/selector/ContextCard.tsx"),
			source("components/ui/button.tsx"),
			source("components/ui/input.tsx"),
		]);

	assert.match(styles, /button:not\(\[data-slot="button"\]\):focus-visible/);
	assert.match(styles, /outline: 2px solid var\(--ring\)/);
	assert.match(shell, /focus-visible:outline-2 focus-visible:outline-ring/);
	assert.match(contextChoices, /<Input/);
	assert.match(contextCard, /contextNavigationDirectionForKey/);
	assert.match(contextCard, /event\.key === "Enter" && selected/);
	assert.match(button, /focus-visible:ring-2/);
	assert.match(input, /focus-visible:ring-2/);
});

test("accessibility: launcher state changes, errors, and dialogs have announced names", async () => {
	const [resolving, progress, error, dialog, mismatch, surface] =
		await Promise.all([
			source("components/launcher/ProjectResolvingView.tsx"),
			source("components/selector/LaunchProgressView.tsx"),
			source("components/selector/GuiErrorNotice.tsx"),
			source("components/ui/dialog.tsx"),
			source("components/selector/ContextMismatchDialog.tsx"),
			source("components/launcher/LauncherSurface.tsx"),
		]);

	assert.match(resolving, /aria-live="polite"/);
	assert.match(progress, /aria-live="polite"/);
	assert.match(error, /<Alert/);
	assert.match(dialog, /DialogPrimitive\.Title/);
	assert.match(dialog, /DialogPrimitive\.Description/);
	assert.match(mismatch, /aria-modal="true"/);
	assert.match(mismatch, /aria-labelledby=/);
	assert.match(surface, /aria-labelledby="launcher-heading"/);
});

test("accessibility: reduced motion and core text/focus colors meet the audit baseline", async () => {
	const styles = await source("style.css");

	assert.match(styles, /@media \(prefers-reduced-motion: reduce\)/);
	assert.match(styles, /animation-duration: 0\.01ms !important/);
	assert.match(styles, /transition-duration: 0\.01ms !important/);
	assert.ok(contrastRatio("#252521", "#f8f7f4") >= 4.5);
	assert.ok(contrastRatio("#6f6d68", "#f8f7f4") >= 4.5);
	assert.ok(contrastRatio("#5f594e", "#f8f7f4") >= 3);
	assert.ok(contrastRatio("#f4f1e9", "#1b1b19") >= 4.5);
	assert.ok(contrastRatio("#e8e4db", "#1b1b19") >= 3);
});
