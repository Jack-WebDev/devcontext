import assert from "node:assert/strict";
import test from "node:test";

import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import { ContextCard } from "../.tmp-test/src/components/selector/ContextCard.js";
import { ProjectIdentity } from "../.tmp-test/src/components/selector/ProjectIdentity.js";
import { SelectorActions } from "../.tmp-test/src/components/selector/SelectorActions.js";
import { SelectorConfidenceSummary } from "../.tmp-test/src/components/selector/SelectorConfidenceSummary.js";
import { SelectorLayout } from "../.tmp-test/src/components/selector/SelectorLayout.js";

test("UX acceptance: everyday launch information is understandable without technical settings", () => {
	const project = { name: "payments-api", path: "/work/payments-api" };
	const context = {
		id: "company",
		name: "Company",
		description: "Work projects and company accounts.",
		accent: "company",
		tool: {
			id: "cursor",
			name: "Cursor",
			status: "ready",
			message: "Cursor is ready for launch.",
		},
		providers: [
			{
				id: "codex",
				name: "Codex",
				enabled: true,
				state: "ready",
				explanation: "Available in this context.",
				identity: {
					status: "verified",
					fields: [{ label: "Email", value: "developer@company.test" }],
				},
			},
			{
				id: "claude",
				name: "Claude",
				enabled: true,
				state: "ready",
				explanation: "Available in this context.",
				identity: {
					status: "unavailable",
					fields: [],
				},
			},
		],
		confidence: {
			contextId: "company",
			status: "ready",
			checks: [
				{
					component: "provider",
					providerId: "codex",
					severity: "ready",
					label: "Codex",
					message: "Codex is ready.",
				},
				{
					component: "provider",
					providerId: "claude",
					severity: "ready",
					label: "Claude",
					message: "Claude is ready.",
				},
				{
					component: "tool",
					toolId: "cursor",
					severity: "ready",
					label: "Cursor",
					message: "Cursor is ready.",
				},
				{
					component: "isolation",
					severity: "ready",
					label: "Context storage",
					message: "Provider and coding-tool storage are isolated.",
				},
			],
		},
	};

	const html = renderToStaticMarkup(
		createElement(SelectorLayout, {
			projectIdentity: createElement(ProjectIdentity, { project }),
			contextCards: createElement(ContextCard, {
				context,
				selected: true,
				onSelect() {},
			}),
			confidenceSummary: createElement(SelectorConfidenceSummary, {
				project,
				context,
			}),
			rememberControl: createElement(
				"p",
				null,
				"Company is the remembered context for this project.",
			),
			launchActions: createElement(SelectorActions, {
				launchDisabled: false,
				launchPending: false,
				projectName: project.name,
				contextName: context.name,
				confidence: context.confidence,
				onLaunch() {},
				onCancel() {},
			}),
		}),
	);

	assert.ok(html.includes(project.name));
	assert.ok(html.includes(project.path));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Selected"));
	assert.ok(html.includes("Codex"));
	assert.ok(html.includes("Email: developer@company.test"));
	assert.ok(html.includes("Claude"));
	assert.ok(html.includes("Account identity unavailable"));
	assert.ok(html.includes("Cursor"));
	assert.ok(html.includes("Isolation"));
	assert.ok(html.includes("Protected"));
	assert.ok(html.includes("Ready"));
	assert.ok(html.includes("Launch Company"));
	assert.doesNotMatch(html, /CODEX_HOME|CLAUDE_CONFIG_DIR|user-data-dir/i);
});
