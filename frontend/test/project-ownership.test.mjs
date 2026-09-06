import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import { contextRecommendation } from "../.tmp-test/src/components/selector/recommendation.js";

const sourceRoot = new URL("../src/", import.meta.url);

test("project recommendations use project bindings instead of integration configuration", async () => {
	const source = await readFile(
		new URL("components/selector/recommendation.ts", sourceRoot),
		"utf8",
	);

	assert.match(source, /launchState\.binding\.contextId === context\.id/);
	assert.doesNotMatch(source, /context\.(?:tool|providers)\b/);
});

test("a remembered project recommends only its bound context", () => {
	const personal = {
		id: "personal",
		name: "Personal",
		confidence: { contextId: "personal", status: "needs_attention", checks: [] },
	};
	const company = {
		id: "company",
		name: "Company",
		confidence: { contextId: "company", status: "needs_attention", checks: [] },
	};
	const launchState = {
		project: { name: "api", path: "/work/api" },
		contexts: [personal, company],
		binding: {
			projectPath: "/work/api",
			bound: true,
			contextId: "company",
			dangling: false,
		},
		resolutionSource: "project_binding",
		warnings: [],
	};

	assert.deepEqual(contextRecommendation(launchState, company), {
		category: "remembered",
		label: "Remembered",
		detail: "Remembered for this project.",
		reasons: ["api is bound to Company."],
	});
	assert.equal(contextRecommendation(launchState, personal), undefined);
});
