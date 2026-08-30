import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const sourceRoot = new URL("../src/", import.meta.url);

async function source(path) {
	return readFile(new URL(path, sourceRoot), "utf8");
}

test("UXF-05 overlay primitives define modal focus, calm surfaces, and disclosures", async () => {
	const [
		dialog,
		sheet,
		alertDialog,
		menu,
		disclosure,
		failure,
		explanation,
		diagnostics,
	] = await Promise.all([
		source("components/ui/dialog.tsx"),
		source("components/ui/sheet.tsx"),
		source("components/ui/alert-dialog.tsx"),
		source("components/ui/dropdown-menu.tsx"),
		source("components/ui/disclosure.tsx"),
		source("components/selector/LaunchFailureView.tsx"),
		source("components/selector/ContextRecommendationExplanation.tsx"),
		source("components/diagnostics/DiagnosticsView.tsx"),
	]);

	assert.match(dialog, /modal = true/);
	assert.match(sheet, /modal = true/);
	for (const overlay of [dialog, alertDialog]) {
		assert.match(overlay, /rounded-2xl/);
	}
	assert.match(sheet, /rounded-[tb]-2xl/);
	assert.match(dialog, /DialogPrimitive\.Close/);
	assert.match(sheet, /SheetPrimitive\.Close/);
	assert.match(alertDialog, /AlertDialogPrimitive\.Close/);
	assert.match(menu, /rounded-xl/);
	assert.match(menu, /data-\[variant=destructive\]/);
	assert.match(disclosure, /<details/);
	assert.match(disclosure, /<summary/);
	assert.match(failure, /<Disclosure/);
	assert.match(explanation, /<Disclosure/);
	assert.match(diagnostics, /<Disclosure/);
});
