import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import test from "node:test";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import { ContextCreateDevelopmentToolsScreen } from "../.tmp-test/src/components/contexts/ContextCreateDevelopmentToolsScreen.js";

const sourceRoot = new URL("../src/", import.meta.url);
const integrationSpecificTerms = /\b(?:claude|codex|vscode|visual\s+studio\s+code)\b/i;

async function frontendSourceFiles(path) {
	const entries = await readdir(new URL(path, sourceRoot), {
		withFileTypes: true,
	});
	const files = await Promise.all(
		entries.map(async (entry) => {
			const entryPath = `${path}/${entry.name}`;
			if (entry.isDirectory()) {
				return frontendSourceFiles(entryPath);
			}
			return /\.tsx?$/.test(entry.name) ? [entryPath] : [];
		}),
	);
	return files.flat();
}

test("generic frontend UI and API layers do not name concrete integrations", async () => {
	const componentFiles = await frontendSourceFiles("components");
	const files = [
		...componentFiles,
		"lib/devctx-api.ts",
		"lib/devctx-window.ts",
	];
	const sources = await Promise.all(
		files.map(async (path) => [path, await readFile(new URL(path, sourceRoot), "utf8")]),
	);

	for (const [path, source] of sources) {
		assert.doesNotMatch(
			source,
			integrationSpecificTerms,
			`${path} contains an integration-specific frontend decision`,
		);
	}
});

test("development-tool UI renders registry metadata for an arbitrary integration", () => {
	const html = renderToStaticMarkup(
		createElement(ContextCreateDevelopmentToolsScreen, {
			integrations: [
				{
					id: "atlas-assistant",
					name: "Atlas Assistant",
					category: "ai",
					status: "needs_sign_in",
					message: "Sign in to Atlas Assistant for this context.",
					recoveryHint: "Open Atlas Assistant to sign in.",
					enabled: true,
				},
			],
			onContinue() {},
		}),
	);

	assert.match(html, /Atlas Assistant/);
	assert.match(html, /Sign in to Atlas Assistant for this context/);
	assert.match(html, /Open Atlas Assistant to sign in/);
	assert.match(html, /Needs sign-in/);
});
