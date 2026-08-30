import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

const sourceRoot = new URL("../src/", import.meta.url);

async function source(path) {
	return readFile(new URL(path, sourceRoot), "utf8");
}

test("UXF-01 layout primitives are defined and used by representative screens", async () => {
	const [styles, shell, launcher, contexts, home] = await Promise.all([
		source("style.css"),
		source("components/shell/AppShell.tsx"),
		source("components/launcher/LauncherSurface.tsx"),
		source("components/contexts/ContextsView.tsx"),
		source("components/home/HomeView.tsx"),
	]);

	for (const token of [
		"--layout-page-max-width",
		"--layout-section-gap",
		"--layout-card-padding",
		"--layout-list-row-min-height",
		"--layout-sidebar-width",
		"--layout-launcher-width",
		"--layout-min-window-width",
	]) {
		assert.match(styles, new RegExp(`${token}:`));
	}

	assert.match(shell, /app-shell-grid/);
	assert.match(launcher, /launcher-container/);
	assert.match(contexts, /page-content page-section-stack/);
	assert.match(home, /home-content/);
});
