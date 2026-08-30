import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const sourceRoot = new URL("../src/", import.meta.url);

async function source(path) {
	return readFile(new URL(path, sourceRoot), "utf8");
}

test("UXF-03 and UXF-04 define shared control states and surface purposes", async () => {
	const [
		button,
		input,
		textarea,
		switchControl,
		checkbox,
		card,
		contextCard,
		remember,
	] = await Promise.all([
		source("components/ui/button.tsx"),
		source("components/ui/input.tsx"),
		source("components/ui/textarea.tsx"),
		source("components/ui/switch.tsx"),
		source("components/ui/checkbox.tsx"),
		source("components/ui/card.tsx"),
		source("components/selector/ContextCard.tsx"),
		source("components/selector/RememberProjectControl.tsx"),
	]);

	for (const control of [button, input, textarea, switchControl, checkbox]) {
		assert.match(control, /focus-visible/);
		assert.match(control, /disabled/);
		assert.match(control, /aria-busy/);
	}
	assert.match(button, /variant: \{/);
	assert.match(button, /destructive/);
	for (const hierarchy of [
		"primary",
		"inset",
		"secondary",
		"tertiary",
		"selection",
	]) {
		assert.match(card, new RegExp(`"${hierarchy}"`));
	}
	assert.match(card, /data-hierarchy/);
	assert.match(contextCard, /hierarchy="selection"/);
	assert.match(remember, /hierarchy="inset"/);
	assert.match(remember, /<Checkbox/);
});
