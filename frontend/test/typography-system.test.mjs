import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const sourceRoot = new URL("../src/", import.meta.url);

async function source(path) {
	return readFile(new URL(path, sourceRoot), "utf8");
}

test("UXF-02 typography roles are defined and used by representative screens", async () => {
	const [styles, home, contexts, launcher, welcome] = await Promise.all([
		source("style.css"),
		source("components/home/HomeView.tsx"),
		source("components/contexts/ContextsView.tsx"),
		source("components/launcher/LauncherSurface.tsx"),
		source("components/selector/WelcomeView.tsx"),
	]);

	for (const role of [
		"text-page-title",
		"text-launcher-title",
		"text-section-title",
		"text-body",
		"text-secondary",
		"text-label",
		"text-status",
		"text-caption",
		"text-detail",
		"text-technical",
	]) {
		assert.match(styles, new RegExp(`\\.${role}\\s*\\{`));
	}

	assert.match(styles, /--font-sans: ui-sans-serif, system-ui/);
	assert.match(home, /text-page-title/);
	assert.match(contexts, /text-page-title/);
	assert.match(launcher, /text-launcher-title/);
	assert.match(launcher, /text-technical/);
	assert.match(welcome, /text-page-title/);
	assert.match(welcome, /text-detail/);
});
