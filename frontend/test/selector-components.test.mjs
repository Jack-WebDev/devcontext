import assert from "node:assert/strict";
import test from "node:test";

import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import {
	launchContextActions,
	navigationActions,
} from "../.tmp-test/src/components/command-palette/actions.js";
import {
	contextPositionFromShortcut,
	isCommandPaletteShortcut,
	keyboardShortcuts,
} from "../.tmp-test/src/components/command-palette/shortcut.js";
import {
	ContextAccentIndicator,
	contextAccentFromMetadata,
} from "../.tmp-test/src/components/context-accent/ContextAccent.js";
import {
	ContextsView,
	providerSummary,
} from "../.tmp-test/src/components/contexts/ContextsView.js";
import { parseContextMetadataExport } from "../.tmp-test/src/components/contexts/context-transfer.js";
import { renderDiagnostics } from "../.tmp-test/src/components/diagnostics/DiagnosticsView.js";
import {
	filterHistoryEntries,
	formatHistoryEvent,
	groupHistoryEntriesByDate,
	HistoryView,
} from "../.tmp-test/src/components/history/HistoryView.js";
import {
	HomeView,
	homeConfidenceSummary,
} from "../.tmp-test/src/components/home/HomeView.js";
import { RecentProjectConfirmationDialog } from "../.tmp-test/src/components/home/RecentProjectConfirmationDialog.js";
import { notificationPresentation } from "../.tmp-test/src/components/notifications/notification-policy.js";
import {
	ProjectContextChangeDialog,
	safetyImplication,
} from "../.tmp-test/src/components/projects/ProjectContextChangeDialog.js";
import {
	formatProjectTime,
	ProjectsView,
} from "../.tmp-test/src/components/projects/ProjectsView.js";
import {
	formatRunningTime,
	RunningView,
} from "../.tmp-test/src/components/running/RunningView.js";
import { AccountIdentityMismatchDialog } from "../.tmp-test/src/components/selector/AccountIdentityMismatchDialog.js";
import { hasAccountIdentityMismatch } from "../.tmp-test/src/components/selector/account-identity-mismatch.js";
import { bindingReplacementForLaunch } from "../.tmp-test/src/components/selector/binding-replacement.js";
import { ContextCard } from "../.tmp-test/src/components/selector/ContextCard.js";
import {
	ContextMismatchDialog,
	mismatchDialogKeyboardAction,
} from "../.tmp-test/src/components/selector/ContextMismatchDialog.js";
import { DanglingBindingDialog } from "../.tmp-test/src/components/selector/DanglingBindingDialog.js";
import { cancelSelector } from "../.tmp-test/src/components/selector/cancel-action.js";
import { missingDefaultContextIds } from "../.tmp-test/src/components/selector/default-context-actions.js";
import {
	FirstRunWelcome,
	shouldRenderFirstRunWelcome,
} from "../.tmp-test/src/components/selector/FirstRunWelcome.js";
import { WelcomeView } from "../.tmp-test/src/components/selector/WelcomeView.js";
import { GuiErrorNotice } from "../.tmp-test/src/components/selector/GuiErrorNotice.js";
import { LaunchFailureView } from "../.tmp-test/src/components/selector/LaunchFailureView.js";
import { LaunchProgressView } from "../.tmp-test/src/components/selector/LaunchProgressView.js";
import { LauncherSurface } from "../.tmp-test/src/components/launcher/LauncherSurface.js";
import { ProjectResolvingView } from "../.tmp-test/src/components/launcher/ProjectResolvingView.js";
import {
	ProjectNotFoundView,
	projectRecoveryEscapeAction,
} from "../.tmp-test/src/components/launcher/ProjectNotFoundView.js";
import {
	ContextChoiceList,
	filterContexts,
	groupContexts,
	shouldShowContextSearch,
} from "../.tmp-test/src/components/selector/ContextChoiceList.js";
import {
	LaunchVerificationProgress,
	verificationStepPresentation,
} from "../.tmp-test/src/components/selector/LaunchVerificationProgress.js";
import { PreflightReviewView } from "../.tmp-test/src/components/selector/PreflightReviewView.js";
import {
	createLaunchRequestGuard,
	launchSelectedContext,
} from "../.tmp-test/src/components/selector/launch-action.js";
import {
	launchActionLabel,
	launchPendingLabel,
} from "../.tmp-test/src/components/selector/launch-copy.js";
import {
	defaultLaunchSuccessCloseBehavior,
	shouldCloseSelectorAfterLaunch,
} from "../.tmp-test/src/components/selector/launch-success-close-behavior.js";
import { createContextAndRefresh } from "../.tmp-test/src/components/contexts/context-creation.js";
import {
	beginContextCreation,
	completeContextCreation,
	initialContextCreateFlow,
	nextContextCreateStep,
	previousContextCreateStep,
	returnToContextCreateReview,
	updateContextCreateDraft,
} from "../.tmp-test/src/components/contexts/context-create-flow.js";
import { ProjectIdentity } from "../.tmp-test/src/components/selector/ProjectIdentity.js";
import { ProviderCredentialClassification } from "../.tmp-test/src/components/selector/ProviderCredentialClassification.js";
import {
	boundContextName,
	canRememberProject,
	RememberProjectControl,
} from "../.tmp-test/src/components/selector/RememberProjectControl.js";
import { ReplaceBindingDialog } from "../.tmp-test/src/components/selector/ReplaceBindingDialog.js";
import { contextRecommendation } from "../.tmp-test/src/components/selector/recommendation.js";
import {
	launchConfidenceFeedback,
	SelectorActions,
} from "../.tmp-test/src/components/selector/SelectorActions.js";
import {
	confidenceStatusPresentation,
	SelectorConfidenceSummary,
} from "../.tmp-test/src/components/selector/SelectorConfidenceSummary.js";
import { SelectorLayout } from "../.tmp-test/src/components/selector/SelectorLayout.js";
import {
	initialRovingContextId,
	initialSelectedContextId,
	nextKeyboardContextId,
	nextSelectedContextId,
} from "../.tmp-test/src/components/selector/selection-state.js";
import { SingleContextLaunchView } from "../.tmp-test/src/components/selector/SingleContextLaunchView.js";
import {
	launcherSelection,
	launcherStateIsPending,
	selectingLauncherState,
} from "../.tmp-test/src/components/selector/launcher-state.js";
import {
	canLaunchSelectedContextFromKeyboard,
	escapeKeyboardAction,
} from "../.tmp-test/src/components/selector/selector-keyboard.js";
import { AppShell } from "../.tmp-test/src/components/shell/AppShell.js";
import {
	appRouteFromHash,
	appRoutes,
} from "../.tmp-test/src/components/shell/routes.js";
import {
	StatusIndicator,
	statusPresentation,
} from "../.tmp-test/src/components/status/StatusIndicator.js";
import { TrustCenterView } from "../.tmp-test/src/components/trust/TrustCenterView.js";
import { Card } from "../.tmp-test/src/components/ui/card.js";
import { createDevContextWindow } from "../.tmp-test/src/lib/devctx-window.js";

test("context metadata import requires a JSON export document", () => {
	const exported = { version: 1, context: { name: "Personal" } };
	assert.deepEqual(
		parseContextMetadataExport(JSON.stringify(exported)),
		exported,
	);
	assert.throws(() => parseContextMetadataExport("not JSON"));
	assert.throws(() => parseContextMetadataExport("[]"));
	assert.throws(() =>
		parseContextMetadataExport(JSON.stringify({ version: "1", context: {} })),
	);
});

test("account identity mismatch review is limited to backend identity evidence", () => {
	assert.equal(
		hasAccountIdentityMismatch({
			confidence: {
				checks: [{ component: "identity", severity: "needs_attention" }],
			},
		}),
		true,
	);
	assert.equal(
		hasAccountIdentityMismatch({
			confidence: {
				checks: [{ component: "provider", severity: "needs_attention" }],
			},
		}),
		false,
	);
	const html = renderToStaticMarkup(
		createElement(AccountIdentityMismatchDialog, {
			contextName: "Company",
			launchPending: false,
			onCancel() {},
			onReviewConfiguration() {},
			onLaunchAnyway() {},
		}),
	);
	assert.ok(html.includes("Review configuration"));
	assert.ok(html.includes("Launch anyway"));
	assert.ok(html.includes("Cancel"));
});

test("launcher surface is focused on one project", () => {
	const html = renderToStaticMarkup(
		createElement(
			LauncherSurface,
			{ projectPath: "/work/api" },
			"Loading launch options...",
		),
	);

	assert.match(html, /data-launcher-flow/);
	assert.match(html, /Open project/);
	assert.match(html, /\/work\/api/);
	assert.doesNotMatch(html, /data-app-shell|aria-label="Main navigation"/);
});

test("project resolving view explains the launch information being loaded", () => {
	const html = renderToStaticMarkup(createElement(ProjectResolvingView));

	assert.match(html, /Preparing your project/);
	assert.match(html, /remembered context/);
	assert.match(html, /available contexts/);
	assert.match(html, /readiness/);
	assert.match(html, /role="status"/);
});

test("project-not-found recovery uses folder selection instead of raw errors", () => {
	const html = renderToStaticMarkup(
		createElement(ProjectNotFoundView, {
			choosingFolder: false,
			onChooseFolder() {},
			onCancel() {},
		}),
	);

	assert.match(html, /Project not found/);
	assert.match(html, /Choose folder/);
	assert.match(html, /Cancel/);
	assert.doesNotMatch(html, /Project path does not exist/);
	assert.equal(projectRecoveryEscapeAction(false), "close-selector");
	assert.equal(projectRecoveryEscapeAction(true), "none");
});

test("launcher state keeps progress and dialogs mutually exclusive", () => {
	const selection = {
		selectedContextId: "company",
		rovingContextId: "company",
		rememberProject: true,
	};

	assert.deepEqual(selectingLauncherState(selection), {
		status: "selecting",
		selection,
	});
	assert.equal(
		launcherStateIsPending({ status: "preflighting", selection }),
		true,
	);
	assert.equal(
		launcherStateIsPending({
			status: "identity_mismatch",
			selection,
			contextId: "company",
		}),
		false,
	);
	assert.equal(
		launcherStateIsPending({
			status: "preflight_review",
			selection,
			preflight: {},
			attempt: {},
		}),
		false,
	);
	assert.deepEqual(
		launcherSelection({
			status: "existing_workspace",
			selection,
			conflict: { kind: "same_context", environment: {} },
		}),
		selection,
	);
});

test("notification policy permits only meaningful provider, tool, and update events", () => {
	assert.deepEqual(
		notificationPresentation({
			kind: "provider_verified",
			providerName: "Codex",
			contextName: "Personal",
		}),
		{
			kind: "provider_verified",
			title: "Codex verified",
			description: "Codex is ready in Personal.",
			severity: "success",
		},
	);
	assert.deepEqual(
		notificationPresentation({
			kind: "provider_attention",
			providerName: "Claude",
			contextName: "Company",
			message: "Sign in again.",
		}),
		{
			kind: "provider_attention",
			title: "Claude needs attention",
			description: "Sign in again.",
			severity: "warning",
		},
	);
	assert.deepEqual(
		notificationPresentation({
			kind: "tool_launched",
			projectName: "API",
			contextName: "Company",
			toolName: "Cursor",
		}),
		{
			kind: "tool_launched",
			title: "Cursor launched",
			description: "API opened in Company.",
			severity: "success",
		},
	);
	assert.deepEqual(
		notificationPresentation({ kind: "update_available", version: "1.2.3" }),
		{
			kind: "update_available",
			title: "Update available",
			description: "Dev Context 1.2.3 is ready to install.",
			severity: "info",
		},
	);
});

test("command palette shortcut accepts Ctrl or Command K without modifiers", () => {
	assert.equal(
		isCommandPaletteShortcut({
			key: "k",
			ctrlKey: true,
			metaKey: false,
			altKey: false,
			shiftKey: false,
		}),
		true,
	);
	assert.equal(
		isCommandPaletteShortcut({
			key: "K",
			ctrlKey: false,
			metaKey: true,
			altKey: false,
			shiftKey: false,
		}),
		true,
	);
	assert.equal(
		isCommandPaletteShortcut({
			key: "k",
			ctrlKey: true,
			metaKey: false,
			altKey: false,
			shiftKey: true,
		}),
		false,
	);
	assert.equal(
		isCommandPaletteShortcut({
			key: "p",
			ctrlKey: true,
			metaKey: false,
			altKey: false,
			shiftKey: false,
		}),
		false,
	);
});

test("shortcut registry maps number keys to visible context positions", () => {
	assert.equal(keyboardShortcuts.command_palette.label, "Ctrl/Cmd+K");
	assert.equal(keyboardShortcuts.select_context.label, "1-9");
	assert.equal(
		contextPositionFromShortcut({
			key: "1",
			ctrlKey: false,
			metaKey: false,
			altKey: false,
			shiftKey: false,
		}),
		0,
	);
	assert.equal(
		contextPositionFromShortcut({
			key: "9",
			ctrlKey: false,
			metaKey: false,
			altKey: false,
			shiftKey: false,
		}),
		8,
	);
	assert.equal(
		contextPositionFromShortcut({
			key: "1",
			ctrlKey: true,
			metaKey: false,
			altKey: false,
			shiftKey: false,
		}),
		undefined,
	);
});

test("command palette actions use context names and configured routes", () => {
	const launched = [];
	const navigated = [];
	const launchActions = launchContextActions(
		[
			{
				id: "personal",
				name: "Personal",
				confidence: { status: "ready" },
				tool: { name: "VS Code" },
			},
			{
				id: "company",
				name: "Company",
				confidence: { status: "blocked" },
				tool: { name: "Cursor" },
			},
		],
		(contextId) => launched.push(contextId),
	);
	const routeActions = navigationActions(
		[
			{ id: "home", label: "Home" },
			{ id: "diagnostics", label: "Diagnostics" },
		],
		(route) => navigated.push(route),
	);

	assert.equal(launchActions[0].label, "Launch Personal");
	assert.equal(launchActions[1].disabled, true);
	launchActions[0].onSelect();
	assert.deepEqual(launched, ["personal"]);
	assert.equal(routeActions[1].label, "Open Diagnostics");
	routeActions[1].onSelect();
	assert.deepEqual(navigated, ["diagnostics"]);
});

test("history groups entries by date and presents project, context, event, and time", () => {
	const entries = [
		{
			event: "launch_succeeded",
			category: "launch",
			timestamp: "2026-08-14T08:30:00Z",
			projectPath: "/work/api",
			contextId: "company",
			message: "Launch succeeded.",
		},
		{
			event: "context_created",
			category: "configuration",
			timestamp: "2026-08-13T15:00:00Z",
			contextId: "personal",
			message: "Context created.",
		},
		{
			event: "project_binding_changed",
			category: "configuration",
			timestamp: "2026-08-14T12:45:00Z",
			projectPath: "/work/web",
			contextId: "personal",
			message: "Project context binding changed.",
		},
	];

	const groups = groupHistoryEntriesByDate(entries);
	assert.equal(groups.length, 2);
	assert.equal(groups[0].date, "2026-08-14");
	assert.equal(groups[0].entries[0].event, "project_binding_changed");
	assert.equal(formatHistoryEvent("provider_reset"), "Provider Reset");

	const html = renderToStaticMarkup(createElement(HistoryView, { entries }));
	assert.ok(html.includes("History"));
	assert.ok(html.includes("/work/api"));
	assert.ok(html.includes("company"));
	assert.ok(html.includes("Launch Succeeded"));
	assert.ok(html.includes("Project context binding changed."));
	assert.match(html, /<time[^>]*dateTime="2026-08-14T12:45:00Z"/);
});

test("history filters by backend category and searches only project and context", () => {
	const entries = [
		{
			event: "launch_succeeded",
			category: "launch",
			timestamp: "2026-08-14T08:30:00Z",
			projectPath: "/work/api",
			contextId: "company",
			message: "Launch succeeded.",
		},
		{
			event: "provider_reset",
			category: "configuration",
			timestamp: "2026-08-14T08:30:00Z",
			contextId: "personal",
			message: "Provider storage reset.",
		},
		{
			event: "launch_process_failure",
			category: "warning",
			timestamp: "2026-08-14T08:30:00Z",
			projectPath: "/work/web",
			contextId: "personal",
			message: "Launch could not start.",
		},
	];

	assert.deepEqual(filterHistoryEntries(entries, "launch", ""), [entries[0]]);
	assert.deepEqual(filterHistoryEntries(entries, "all", "PERSONAL"), [
		entries[1],
		entries[2],
	]);
	assert.deepEqual(filterHistoryEntries(entries, "warning", "api"), []);
});

test("history presents an empty activity state", () => {
	const html = renderToStaticMarkup(
		createElement(HistoryView, { entries: [] }),
	);

	assert.ok(html.includes("No activity has been recorded yet."));
});

test("running environments show immutable launch context and tool action entry points", () => {
	const environments = [
		{
			id: "environment-1",
			project: { name: "api", path: "/work/api" },
			context: { id: "company", name: "Company" },
			tool: { id: "second-tool", name: "Second Tool" },
			startedAt: "2026-08-28T10:30:00Z",
			process: { state: "running" },
			session: { state: "unknown" },
			launch: { source: "gui", resolutionSource: "explicit" },
		},
	];
	const html = renderToStaticMarkup(
		createElement(RunningView, { environments }),
	);

	assert.ok(html.includes("Running"));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Second Tool"));
	assert.ok(html.includes("Reveal"));
	assert.ok(html.includes("Switch to"));
	assert.ok(html.includes("Stop"));
	assert.match(html, /disabled=""/);
	assert.notEqual(formatRunningTime("invalid"), "Invalid Date");
});

test("project identity presents the current project name and path in a compact block", () => {
	const projects = [
		{
			name: "api",
			path: "/work/api",
		},
		{
			name: "very-long-project-name-that-should-truncate-visually",
			path: "/workspace/customer-platform/services/backend/packages/feature-flags/current",
		},
		{
			name: "Project With Spaces",
			path: "/Users/alex/Client Work/Project With Spaces",
		},
		{
			name: "Café Portal",
			path: "/Users/alex/équipe/Café Portal",
		},
	];

	for (const project of projects) {
		const html = renderToStaticMarkup(ProjectIdentity({ project }));

		assert.match(html, /data-selector-project-identity="true"/);
		assert.ok(html.includes("Current project"));
		assert.ok(html.includes("Git branch"));
		assert.ok(html.includes("Last opened"));
		assert.match(html, />Unavailable</);
		assert.match(html, /class="[^"]*truncate[^"]*"/);
		assert.match(html, new RegExp(`title="${escapeRegExp(project.name)}"`));
		assert.match(html, new RegExp(`title="${escapeRegExp(project.path)}"`));
		assert.ok(html.includes(project.name));
		assert.ok(html.includes(project.path));
	}
});

test("first-run welcome explains local identity boundaries and setup choices", () => {
	const state = launchStateFixture({
		contexts: [],
		firstRun: true,
	});
	const html = renderToStaticMarkup(
		FirstRunWelcome({
			launchState: state,
			onCreatePersonal: () => {},
			onCreateCompany: () => {},
		}),
	);

	assert.ok(html.includes("Create your first development context"));
	assert.ok(html.includes("Local first"));
	assert.ok(html.includes("Nothing is synced or uploaded by Dev Context."));
	assert.ok(html.includes("Isolated tools"));
	assert.ok(html.includes("Separate authentication"));
	assert.ok(
		html.includes("does not store passwords, tokens, or cloud accounts"),
	);
	assert.ok(html.includes("Create Personal"));
	assert.ok(html.includes("Create Company"));
	assert.ok(html.includes(state.project.path));
	assert.doesNotMatch(html, /disabled=""/);
});

test("welcome screen explains contexts and their local separation", () => {
	const html = renderToStaticMarkup(
		createElement(WelcomeView, { onCreateFirstContext() {} }),
	);

	assert.ok(html.includes("Welcome to Dev Context"));
	assert.ok(html.includes("Create Your First Context"));
	assert.ok(html.includes("development identity, not a provider profile"));
	assert.ok(html.includes("Personal work, Work, a Client, or Open Source"));
	assert.ok(html.includes("What stays separate"));
	assert.ok(html.includes("Projects open with the context you choose"));
	assert.ok(html.includes("Account sessions and tool data use separate local storage"));
	assert.ok(html.includes("Tool settings and launch environment"));
	assert.ok(html.includes("without changing another context"));
	assert.doesNotMatch(html, /Classify detected provider sessions/);
	assert.doesNotMatch(html, /Create Personal/);
	assert.doesNotMatch(html, /Create Company/);
});

test("first-run welcome disables setup actions until handlers are wired", () => {
	const state = launchStateFixture({
		contexts: [],
		firstRun: true,
	});
	const html = renderToStaticMarkup(FirstRunWelcome({ launchState: state }));
});

test("first-run welcome requires detected provider sessions to be classified before creation", () => {
	const state = launchStateFixture({
		contexts: [],
		firstRun: true,
		providerCredentialSessions: providerCredentialSessionsFixture(),
	});

	const unassigned = renderToStaticMarkup(
		FirstRunWelcome({
			launchState: state,
			providerCredentialSessions: state.providerCredentialSessions,
			providerSessionAssignments: {},
			onCreatePersonal: () => {},
			onCreateCompany: () => {},
			onClassifyProviderSession: () => {},
		}),
	);
	const assigned = renderToStaticMarkup(
		FirstRunWelcome({
			launchState: state,
			providerCredentialSessions: state.providerCredentialSessions,
			providerSessionAssignments: { codex: "company", claude: "personal" },
			onCreatePersonal: () => {},
			onCreateCompany: () => {},
			onClassifyProviderSession: () => {},
		}),
	);

	assert.ok(unassigned.includes("Classify detected provider sessions"));
	assert.ok(unassigned.includes("Email:"));
	assert.ok(unassigned.includes("user@company.com"));
	assert.ok(unassigned.includes("Plan:"));
	assert.ok(unassigned.includes("Business"));
	assert.ok(unassigned.includes("Subscription:"));
	assert.ok(unassigned.includes("Pro"));
	assert.ok(unassigned.includes("Organization:"));
	assert.ok(unassigned.includes("Jishin Labs"));
	assert.ok(unassigned.includes("Organization ID:"));
	assert.ok(unassigned.includes("e783"));
	assert.match(unassigned, /disabled=""/);
	assert.doesNotMatch(assigned, /disabled=""/);
});

test("provider credential classification renders only safe metadata fields", () => {
	const html = renderToStaticMarkup(
		ProviderCredentialClassification({
			sessions: [...providerCredentialSessionsFixture()],
			assignments: { codex: "personal" },
			onClassify: () => {},
		}),
	);

	assert.ok(html.includes("Codex"));
	assert.ok(html.includes("Claude"));
	assert.ok(html.includes("user@company.com"));
	assert.ok(html.includes("Business"));
	assert.ok(html.includes("acct_123"));
	assert.ok(html.includes("Pro"));
	assert.ok(html.includes("Jishin Labs"));
	assert.ok(html.includes("e783"));
});

test("provider credential classification renders a future provider without a provider-specific branch", () => {
	const html = renderToStaticMarkup(
		ProviderCredentialClassification({
			sessions: [
				{
					providerId: "future",
					name: "Future Provider",
					metadataAvailable: true,
					fields: [{ label: "Workspace", value: "Example" }],
				},
			],
			assignments: { future: "company" },
			onClassify: () => {},
		}),
	);

	assert.ok(html.includes("Future Provider"));
	assert.ok(html.includes("Workspace:"));
	assert.ok(html.includes("Example"));
	assert.ok(html.includes("Current global Future Provider session"));
});

test("first-run welcome shows pending and error states", () => {
	const state = launchStateFixture({
		contexts: [],
		firstRun: true,
	});
	const pending = renderToStaticMarkup(
		FirstRunWelcome({
			launchState: state,
			onCreatePersonal: () => {},
			onCreateCompany: () => {},
			pendingContextId: "company",
		}),
	);
	const failed = renderToStaticMarkup(
		FirstRunWelcome({
			launchState: state,
			error: apiError(
				"validation_error",
				"Unable to complete request.",
				"Check the selected project and context, then retry.",
			).error,
		}),
	);

	assert.ok(pending.includes("Creating Company context..."));
	assert.ok(pending.includes("Creating..."));
	assert.match(pending, /role="status"/);
	assert.match(pending, /disabled=""/);
	assert.match(failed, /role="alert"/);
	assert.ok(failed.includes("Unable to complete request."));
});

test("first-run predicate separates new and returning users", () => {
	assert.equal(
		shouldRenderFirstRunWelcome(
			launchStateFixture({
				contexts: [],
				firstRun: true,
			}),
		),
		true,
	);
	assert.equal(shouldRenderFirstRunWelcome(launchStateFixture()), false);
	assert.equal(
		shouldRenderFirstRunWelcome(
			launchStateFixture({ contexts: [], firstRun: false }),
		),
		true,
	);
});

test("a single unbound context is selected so it can be launched", () => {
	const state = launchStateFixture({
		contexts: [contextFixture("personal", "Personal")],
	});

	assert.equal(initialSelectedContextId(state), "personal");
});

test("single-context confirmation stays concise and offers alternatives", () => {
	const html = renderToStaticMarkup(
		createElement(SingleContextLaunchView, {
			context: contextFixture("personal", "Personal"),
			projectName: "api",
			onChooseAnother() {},
		}),
	);

	assert.match(html, /Open api with Personal/);
	assert.match(html, /Personal is ready to open this project/);
	assert.match(html, /Choose another context/);
});

test("context choices separate recommendations and search only larger collections", () => {
	const contexts = [
		contextFixture("personal", "Personal"),
		contextFixture("company", "Company"),
		contextFixture("client-a", "Client A"),
	];
	const rememberedContext = contexts[1];
	assert.deepEqual(
		groupContexts(
			contexts.map((context) => ({
				context,
				recommendation:
					context.id === rememberedContext.id
						? {
								category: "remembered",
								label: "Remembered",
								detail: "Remembered for this project.",
								reasons: ["api is bound to Company."],
							}
						: undefined,
			})),
		).map((group) => [
			group.label,
			group.contexts.map(({ context }) => context.id),
		]),
		[
			["Remembered context", ["company"]],
			["Other contexts", ["personal", "client-a"]],
		],
	);
	assert.deepEqual(filterContexts(contexts, "client"), [contexts[2]]);
	assert.equal(shouldShowContextSearch(contexts), false);
	assert.equal(
		shouldShowContextSearch([
			...contexts,
			contextFixture("four", "Four"),
			contextFixture("five", "Five"),
			contextFixture("six", "Six"),
		]),
		true,
	);

	const html = renderToStaticMarkup(
		createElement(ContextChoiceList, {
			launchState: launchStateFixture({
				contexts,
				selectedContextId: "company",
				resolutionSource: "project_binding",
				binding: {
					projectPath: "/work/api",
					bound: true,
					contextId: "company",
					dangling: false,
				},
			}),
			selectedContextId: "company",
			rovingContextId: "company",
			launchPending: false,
			keyboardLaunchAvailable: true,
			search: "",
			onSearchChange() {},
			buttonRef() {
				return null;
			},
			onSelect() {},
			onNavigate() {},
			onLaunch() {},
			onProviderSetup() {},
		}),
	);
	assert.match(html, /Remembered context/);
	assert.match(html, /Remembered/);
	assert.match(html, /Other contexts/);
	assert.doesNotMatch(html, /Recommended/);
	assert.doesNotMatch(html, /Enabled providers/);
});

test("onboarding replay does not offer duplicate context creation", () => {
	const replayHTML = renderToStaticMarkup(
		createElement(FirstRunWelcome, {
			launchState: launchStateFixture({ firstRun: false }),
			replay: true,
			onContinue() {},
		}),
	);

	assert.ok(replayHTML.includes("Development context setup"));
	assert.ok(replayHTML.includes("Continue to context selector"));
	assert.ok(!replayHTML.includes("Create Personal"));
	assert.ok(!replayHTML.includes("Create Company"));
});

test("default context setup finds only missing default contexts", () => {
	assert.deepEqual(
		missingDefaultContextIds([contextFixture("personal", "Personal")]),
		["company"],
	);
	assert.deepEqual(
		missingDefaultContextIds([contextFixture("company", "Company")]),
		["personal"],
	);
	assert.deepEqual(
		missingDefaultContextIds([
			contextFixture("personal", "Personal"),
			contextFixture("company", "Company"),
		]),
		[],
	);
	assert.deepEqual(
		missingDefaultContextIds([contextFixture("client-a", "Client A")]),
		["personal", "company"],
	);
});

test("context card renders generic context names and ids", () => {
	const contexts = [
		contextFixture("personal", "Personal"),
		contextFixture("company", "Company"),
		contextFixture("client-a", "Client A"),
	];

	for (const context of contexts) {
		const html = renderToStaticMarkup(ContextCard({ context }));

		assert.match(html, /<article/);
		assert.ok(html.includes(context.name));
		assert.ok(html.includes(context.id));
		assert.match(html, new RegExp(`title="${escapeRegExp(context.name)}"`));
		assert.match(html, new RegExp(`title="${escapeRegExp(context.id)}"`));
		assert.ok(html.includes("Enabled providers"));
		assert.ok(html.includes("No providers enabled."));
	}
});

test("context card can represent a selected context", () => {
	const context = contextFixture("personal", "Personal");
	const html = renderToStaticMarkup(
		ContextCard({ context, selected: true, onSelect: () => {} }),
	);

	assert.match(html, /data-selected="true"/);
	assert.match(html, /border-primary/);
	assert.match(html, /data-context-selection-marker="true"/);
	assert.match(html, /aria-current="true"/);
	assert.match(html, />Selected</);
});

test("context card presents an evidence-backed label with its reason", () => {
	const context = contextFixture("company", "Company");
	const html = renderToStaticMarkup(
		ContextCard({
			context,
			recommendation: {
				category: "remembered",
				label: "Remembered",
				detail: "Remembered for this project.",
				reasons: ["api is bound to Company."],
			},
		}),
	);

	assert.match(html, />Remembered</);
	assert.ok(html.includes("Remembered for this project."));
	assert.match(html, /Why this context/);
	assert.ok(html.includes("api is bound to Company."));
	assert.match(html, /<details/);
	assert.doesNotMatch(html, /<details(?:\s[^>]*)?\sopen(?:=|\s|>)/);
});

test("recommendation labels only reflect backend binding, verification, and conflict evidence", () => {
	const context = {
		...contextFixture("company", "Company"),
		confidence: { contextId: "company", status: "ready", checks: [] },
	};
	const remembered = launchStateFixture({
		contexts: [context],
		binding: {
			projectPath: "/work/api",
			bound: true,
			contextId: "company",
			dangling: false,
		},
		resolutionSource: "project_binding",
	});
	assert.deepEqual(contextRecommendation(remembered, context), {
		category: "remembered",
		label: "Remembered",
		detail: "Remembered for this project.",
		reasons: ["api is bound to Company."],
	});

	const verified = launchStateFixture({ contexts: [context] });
	assert.deepEqual(contextRecommendation(verified, context), {
		category: "verified",
		label: "Verified",
		detail: "Dev Context verified the required launch checks.",
		reasons: ["Required launch checks are ready."],
	});

	const conflict = launchStateFixture({
		contexts: [context],
		warnings: [
			{
				code: "context_mismatch",
				message: "Mismatch",
				requestedContextId: "company",
			},
		],
	});
	assert.deepEqual(contextRecommendation(conflict, context), {
		category: "conflict",
		label: "Conflict",
		detail: "This context conflicts with the project's remembered context.",
		reasons: ["Mismatch"],
	});
	assert.equal(
		contextRecommendation(launchStateFixture({ contexts: [context] }), {
			...context,
			confidence: {
				contextId: "company",
				status: "needs_attention",
				checks: [],
			},
		}),
		undefined,
	);
});

test("context card renders backend-provided description and accent metadata", () => {
	const context = {
		...contextFixture("company", "Company"),
		description: "Work environment",
		metadata: {
			accent: "slate-blue",
		},
	};
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.match(html, /data-context-accent="slate-blue"/);
	assert.ok(html.includes("Work environment"));
	assert.match(html, /bg-accent-company/);
});

test("context accents use shared semantic tokens without changing the full card theme", () => {
	assert.equal(contextAccentFromMetadata("sage"), "sage");
	assert.equal(contextAccentFromMetadata("future-accent"), "neutral");

	const html = renderToStaticMarkup(
		ContextAccentIndicator({ accent: "custom" }),
	);
	assert.match(html, /bg-accent-custom/);
	assert.doesNotMatch(html, /bg-card/);
});

test("status indicator renders approved status labels with non-color text", () => {
	const statuses = [
		"ready",
		"needs_attention",
		"not_configured",
		"blocked",
		"running",
		"protected",
		"failed",
	];
	const labels = statuses.map((status) => statusPresentation(status).label);

	assert.deepEqual(labels, [
		"Ready",
		"Needs attention",
		"Not configured",
		"Blocked",
		"Running",
		"Protected",
		"Failed",
	]);
	const html = renderToStaticMarkup(StatusIndicator({ status: "blocked" }));
	assert.ok(html.includes("Blocked"));
	assert.match(html, /bg-destructive/);
});

test("card hierarchy differentiates primary, inset, secondary, tertiary, and selection surfaces", () => {
	const primary = renderToStaticMarkup(Card({ hierarchy: "primary" }));
	const inset = renderToStaticMarkup(Card({ hierarchy: "inset" }));
	const secondary = renderToStaticMarkup(Card({ hierarchy: "secondary" }));
	const tertiary = renderToStaticMarkup(Card({ hierarchy: "tertiary" }));
	const selection = renderToStaticMarkup(Card({ hierarchy: "selection" }));

	assert.match(primary, /shadow-sm/);
	assert.match(inset, /surface-subtle/);
	assert.match(secondary, /bg-surface-muted/);
	assert.match(tertiary, /bg-transparent/);
	assert.match(selection, /data-hierarchy="selection"/);
});

test("app shell exposes stable navigation, current project state, and a responsive content boundary", () => {
	const html = renderToStaticMarkup(
		AppShell({
			activeRoute: "projects",
			onNavigate: () => {},
			currentProject: { name: "api", path: "/work/api" },
			children: "Content",
		}),
	);

	assert.match(html, /data-app-shell="true"/);
	assert.match(html, /aria-label="Primary navigation"/);
	assert.match(html, /aria-current="page"/);
	assert.match(html, /app-page-container/);
	assert.match(html, /min-w-0/);
	assert.ok(html.includes("Current project"));
	assert.ok(html.includes("api"));
	assert.ok(html.includes("/work/api"));
	assert.deepEqual(
		appRoutes.map((route) => route.label),
		[
			"Home",
			"Contexts",
			"Projects",
			"Running",
			"History",
			"Settings",
			"Trust Center",
		],
	);
	assert.equal(appRouteFromHash("#projects"), "projects");
	assert.equal(appRouteFromHash("#unknown"), "home");
});

test("Trust Center presents actual isolation, mappings, integration boundaries, and credential-sync state", () => {
	const html = renderToStaticMarkup(
		createElement(TrustCenterView, {
			state: {
				contexts: [
					{
						id: "personal",
						name: "Personal",
						providers: [
							{
								id: "codex",
								name: "Codex",
								isolation: {
									status: "ready",
									message: "Codex isolation storage is ready.",
								},
							},
						],
						tool: {
							id: "vscode",
							name: "VS Code",
							isolation: {
								status: "ready",
								message: "VS Code isolation storage is ready.",
							},
						},
					},
				],
				projectMappings: [
					{
						project: { name: "api", path: "/work/api" },
						contextId: "personal",
						contextName: "Personal",
					},
				],
				credentialSync: {
					enabled: false,
					message: "Dev Context does not sync credentials.",
				},
				integrationBoundaries: [
					{
						toolId: "vscode",
						toolName: "VS Code",
						statusDataAvailable: true,
						message: "Safe status data stays in tool storage.",
					},
				],
			},
		}),
	);
	assert.ok(html.includes("Trust Center"));
	assert.ok(html.includes("Credential sync"));
	assert.ok(html.includes("Codex isolation storage is ready."));
	assert.ok(html.includes("Suggested context: Personal"));
	assert.ok(html.includes("Safe status data available"));
});

test("Home shows project, selected context, and context-named quick launch", () => {
	const html = renderToStaticMarkup(
		HomeView({
			dashboard: {
				project: { name: "api", path: "/work/api" },
				currentContext: {
					id: "company",
					name: "Company",
					tool: {
						id: "tool",
						name: "Future Tool",
						status: "ready",
						message: "Ready",
					},
					confidence: { contextId: "company", status: "ready", checks: [] },
				},
				recentProjects: [],
				running: { count: 0, contextCounts: [], isolationProtected: false },
				activity: { count: 0 },
			},
			launchPending: false,
			onQuickLaunch: () => {},
			onReviewLaunchOptions: () => {},
		}),
	);

	assert.ok(html.includes("Selected project"));
	assert.ok(html.includes("/work/api"));
	assert.ok(html.includes("Git branch"));
	assert.ok(html.includes("Last opened"));
	assert.ok(html.includes("Current context"));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Future Tool"));
	assert.ok(html.includes("Launch Company"));
	assert.equal(
		homeConfidenceSummary("blocked"),
		"This context is blocked until its required setup is resolved.",
	);
});

test("Home lists recent projects for review and requires confirmation before launch", () => {
	const recentProject = {
		project: { name: "api", path: "/work/api" },
		contextId: "company",
		contextName: "Company",
		lastLaunchedAt: "2026-08-28T10:30:00Z",
	};
	const homeHtml = renderToStaticMarkup(
		HomeView({
			dashboard: {
				project: { name: "current", path: "/work/current" },
				recentProjects: [recentProject],
				running: { count: 0, contextCounts: [], isolationProtected: false },
				activity: { count: 0 },
			},
			launchPending: false,
			onQuickLaunch: () => {},
			onReviewLaunchOptions: () => {},
			onRecentProjectSelect: () => {},
		}),
	);
	const dialogHtml = renderToStaticMarkup(
		RecentProjectConfirmationDialog({
			project: recentProject,
			launchPending: false,
			onCancel: () => {},
			onConfirm: () => {},
		}),
	);

	assert.ok(homeHtml.includes("Recent projects"));
	assert.ok(homeHtml.includes("Review a project before launching it."));
	assert.ok(homeHtml.includes("/work/api"));
	assert.ok(homeHtml.includes("Company"));
	assert.ok(dialogHtml.includes("Launch recent project?"));
	assert.ok(
		dialogHtml.includes(
			"Dev Context will check this project and context before opening it.",
		),
	);
	assert.ok(dialogHtml.includes("Launch Company"));
	assert.match(dialogHtml, /role="dialog"/);
});

test("Projects lists known projects with safe launch and management entry points", () => {
	const html = renderToStaticMarkup(
		ProjectsView({
			projects: [
				{
					project: { name: "api", path: "/work/api" },
					contextId: "company",
					contextName: "Company",
					lastLaunchedAt: "2026-08-28T10:30:00Z",
					running: true,
				},
			],
			onLaunch: () => {},
			onChangeContext: () => {},
			onOpenFolder: () => {},
		}),
	);

	assert.ok(html.includes("Known projects"));
	assert.ok(html.includes("/work/api"));
	assert.ok(html.includes("Remembered context"));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Running"));
	assert.ok(html.includes("Launch Company"));
	assert.ok(html.includes("Change context"));
	assert.ok(html.includes("Open folder"));
	assert.ok(html.includes("Forget project"));
	assert.match(html, /Forget project<\/button>/);
	assert.match(html, /disabled=""/);
	assert.equal(formatProjectTime(undefined), "Never launched");
});

test("Project context changes are explicit and show backend safety implications", () => {
	const html = renderToStaticMarkup(
		createElement(ProjectContextChangeDialog, {
			project: {
				project: { name: "api", path: "/work/api" },
				contextId: "personal",
				contextName: "Personal",
				running: false,
			},
			contexts: [
				{
					id: "personal",
					name: "Personal",
					tool: {
						id: "tool",
						name: "Future Tool",
						status: "ready",
						message: "Ready",
					},
					availableTools: [],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "needs_attention",
						checks: [],
					},
				},
			],
			pending: false,
			onCancel: () => {},
			onConfirm: () => {},
		}),
	);

	assert.match(html, /role="dialog"/);
	assert.ok(html.includes("Change project context"));
	assert.ok(html.includes("Current context"));
	assert.ok(html.includes("Safety implications"));
	assert.ok(html.includes("can launch, but its setup needs attention"));
	assert.ok(html.includes("Use Personal"));
	assert.equal(
		safetyImplication("blocked", "Company"),
		"Company is blocked and cannot launch until its required setup is resolved.",
	);
});

test("Diagnostics groups backend checks and keeps paths in a disclosure", () => {
	const html = renderToStaticMarkup(
		renderDiagnostics(
			{
				status: "loaded",
				data: {
					groups: [
						{
							id: "context-filesystem",
							label: "Context filesystem",
							checks: [
								{
									id: "context-directory",
									severity: "ready",
									label: "Context directory",
									message: "Context directory is available.",
									details: [
										{ label: "Mode", value: "-rwx------", isPath: false },
										{
											label: "Location",
											value: "/contexts/personal",
											isPath: true,
										},
									],
								},
							],
						},
					],
				},
			},
			1,
		),
	);

	assert.ok(html.includes("Context filesystem"));
	assert.ok(html.includes("Context directory is available."));
	assert.ok(html.includes("Mode"));
	assert.ok(html.includes("Show paths"));
	assert.match(html, /<details/);
	assert.doesNotMatch(html, /<details open/);
});

test("Contexts screen lists backend-owned identity summaries and reserves creation", () => {
	const context = {
		context: {
			id: "company",
			name: "Company",
			description: "Work identity",
			tool: {
				id: "tool",
				name: "Future Tool",
				status: "ready",
				message: "Ready",
			},
			availableTools: [],
			providers: [],
			confidence: {
				contextId: "company",
				status: "needs_attention",
				checks: [],
			},
		},
		enabledProviders: [
			{
				id: "provider",
				name: "Provider",
				enabled: true,
				state: "ready",
				identity: { status: "none", fields: [] },
			},
		],
		projectCount: 2,
		lastUsedAt: "2026-08-28T10:30:00Z",
	};
	const html = renderToStaticMarkup(ContextsView({ contexts: [context] }));

	assert.ok(html.includes("Contexts"));
	assert.ok(html.includes("New context"));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Work identity"));
	assert.ok(html.includes("Future Tool"));
	assert.ok(html.includes("2 projects"));
	assert.ok(html.includes("Provider"));
	assert.equal(
		providerSummary({ ...context, enabledProviders: [] }),
		"No providers enabled",
	);
});

test("context card summarizes provider, tool, and isolation health from confidence checks", () => {
	const context = {
		...contextFixture("company", "Company", [
			providerFixture("provider", "Provider", true, "ready"),
		]),
		confidence: {
			contextId: "company",
			status: "needs_attention",
			checks: [
				{
					component: "provider",
					providerId: "provider",
					severity: "ready",
					label: "Provider",
					message: "Provider is ready.",
				},
				{
					component: "tool",
					toolId: "vscode",
					severity: "needs_attention",
					label: "VS Code",
					message: "Review VS Code.",
				},
				{
					component: "isolation",
					severity: "ready",
					label: "Context storage",
					message: "Context storage is ready.",
				},
			],
		},
	};
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.ok(html.includes("Context health"));
	assert.ok(html.includes("1 provider"));
	assert.ok(html.includes("VS Code"));
	assert.ok(html.includes("Needs attention"));
	assert.ok(html.includes("Isolation"));
	assert.ok(html.includes("Protected"));
});

test("context card renders backend-provided coding tool readiness and guidance", () => {
	const context = contextFixture("client-a", "Client A");
	context.tool = {
		id: "future-tool",
		name: "Future Tool",
		status: "blocked",
		message: "Future Tool is not available for launch.",
		actionHint: "Install Future Tool or configure its executable.",
	};

	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.ok(html.includes("Future Tool"));
	assert.ok(html.includes("Blocked"));
	assert.ok(html.includes("Future Tool is not available for launch."));
	assert.ok(html.includes("Install Future Tool or configure its executable."));
});

test("context card renders as a selectable control when wired", () => {
	const context = contextFixture("client-a", "Client A");
	const html = renderToStaticMarkup(
		ContextCard({ context, tabIndex: -1, onSelect: () => {} }),
	);

	assert.match(html, /<button/);
	assert.match(html, /aria-pressed="false"/);
	assert.match(html, /tabindex="-1"/);
	assert.match(html, /absolute inset-0/);
	assert.ok(html.includes("Select Client A"));
});

test("context card renders enabled provider status variants with accessible names", () => {
	const context = contextFixture("personal", "Personal", [
		providerFixture("claude-ready", "Claude", true, "ready"),
		providerFixture(
			"codex-not-configured",
			"Codex",
			true,
			"not_configured",
			"Codex isolated provider state was not found",
		),
		providerFixture(
			"claude-missing",
			"Claude",
			true,
			"directory_missing",
			"Claude context directory is missing",
		),
		providerFixture(
			"codex-unavailable",
			"Codex",
			true,
			"unavailable",
			"Codex context directory could not be inspected",
		),
		providerFixture("disabled", "Disabled Provider", false, "ready"),
	]);
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.match(html, /Claude local status: Ready/);
	assert.match(html, /Codex local status: Not configured/);
	assert.match(html, /Claude local status: Directory missing/);
	assert.match(html, /Codex local status: Unavailable/);
	assert.ok(html.includes("Codex isolated provider state was not found"));
	assert.ok(!html.includes("Disabled Provider"));
});

test("context card renders only backend-provided provider identity information", () => {
	const context = contextFixture("personal", "Personal", [
		{
			...providerFixture("verified", "Verified Provider", true, "ready"),
			identity: {
				status: "verified",
				fields: [{ label: "Email", value: "developer@example.com" }],
			},
		},
		{
			...providerFixture("unavailable", "Unavailable Provider", true, "ready"),
			identity: { status: "unavailable", fields: [] },
		},
	]);
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.ok(html.includes("Account: Email: developer@example.com"));
	assert.ok(html.includes("Account identity unavailable"));
	assert.ok(!html.includes("Personal account"));
});

test("context card renders generic setup guidance for an unconfigured provider", () => {
	const personal = contextFixture("personal", "Personal", [
		providerFixture(
			"codex",
			"Codex",
			true,
			"not_configured",
			"Codex isolated provider state was not found",
		),
		providerFixture("claude", "Claude", true, "ready"),
	]);
	const company = contextFixture("company", "Company", [
		providerFixture("codex", "Codex", true, "ready"),
		{
			...providerFixture("internal", "Internal Tool", true, "not_configured"),
			actionHint: "Connect Internal Tool to Company.",
		},
	]);
	const html = [
		renderToStaticMarkup(ContextCard({ context: personal })),
		renderToStaticMarkup(ContextCard({ context: company })),
	].join("");

	assert.ok(
		html.includes("Codex is enabled for Personal but is not configured."),
	);
	assert.ok(html.includes("Open Personal and complete Codex setup."));
	assert.ok(!html.includes("Claude is enabled for Personal"));
	assert.ok(!html.includes("Codex is enabled for Company"));
	assert.ok(html.includes("Connect Internal Tool to Company."));
});

test("context card offers the backend-supplied provider setup action", () => {
	const context = contextFixture("personal", "Personal", [
		{
			...providerFixture("codex", "Codex", true, "not_configured"),
			setupAction: {
				state: "open_and_configure",
				label: "Open and configure",
				message: "Sign in to Codex for this context.",
			},
		},
	]);
	const html = renderToStaticMarkup(
		ContextCard({ context, onProviderSetup: () => {} }),
	);

	assert.ok(html.includes("Sign in to Codex for this context."));
	assert.ok(html.includes("Open and configure"));
	assert.doesNotMatch(html, /Open and configure<\/button>[^]*disabled=""/);
});

test("context card renders the backend-supplied provider sign-in waiting state", () => {
	const context = contextFixture("company", "Company", [
		{
			...providerFixture("future", "Future Provider", true, "ready"),
			setupAction: {
				state: "waiting_for_sign_in",
				label: "Waiting for sign-in",
				message: "Waiting for Future Provider sign-in verification.",
			},
		},
	]);
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.match(html, /role="status"/);
	assert.ok(html.includes("Waiting for sign-in"));
	assert.ok(html.includes("Waiting for Future Provider sign-in verification."));
	assert.doesNotMatch(html, /Open and configure/);
});

test("context card shows connected provider identity only after backend verification", () => {
	const context = contextFixture("personal", "Personal", [
		{
			...providerFixture("future", "Future Provider", true, "ready"),
			identity: {
				status: "verified",
				fields: [{ label: "Workspace", value: "Example" }],
			},
			setupAction: {
				state: "verified",
				label: "Verified",
				message:
					"Future Provider account identity is verified for this context.",
			},
		},
	]);
	const html = renderToStaticMarkup(ContextCard({ context }));

	assert.match(html, /role="status"/);
	assert.ok(html.includes("Verified"));
	assert.ok(
		html.includes(
			"Future Provider account identity is verified for this context.",
		),
	);
	assert.ok(html.includes("Account: Workspace: Example"));
	assert.ok(!html.includes("Account identity unavailable"));
});

test("selection initializes from a valid bound context", () => {
	const state = launchStateFixture({
		binding: {
			projectPath: "/work/api",
			bound: true,
			contextId: "company",
			dangling: false,
		},
		selectedContextId: "company",
	});

	assert.equal(initialSelectedContextId(state), "company");
	assert.equal(initialRovingContextId(state), "company");
});

test("selection stays empty for unbound and dangling launch states", () => {
	assert.equal(
		initialSelectedContextId(
			launchStateFixture({
				binding: {
					projectPath: "/work/api",
					bound: false,
					dangling: false,
				},
			}),
		),
		undefined,
	);

	assert.equal(
		initialSelectedContextId(
			launchStateFixture({
				binding: {
					projectPath: "/work/api",
					bound: true,
					contextId: "missing",
					dangling: true,
					missingContextId: "missing",
				},
			}),
		),
		undefined,
	);
});

test("selection changes to one existing context at a time", () => {
	const contexts = [
		contextFixture("personal", "Personal"),
		contextFixture("company", "Company"),
		contextFixture("client-a", "Client A"),
	];

	assert.equal(nextSelectedContextId(contexts, "client-a"), "client-a");
	assert.equal(nextSelectedContextId(contexts, "missing"), undefined);
});

test("keyboard context navigation clamps at the start, middle, and end", () => {
	const contexts = [
		contextFixture("personal", "Personal"),
		contextFixture("company", "Company"),
		contextFixture("client-a", "Client A"),
	];

	assert.equal(
		nextKeyboardContextId(contexts, "personal", "previous"),
		"personal",
	);
	assert.equal(nextKeyboardContextId(contexts, "personal", "next"), "company");
	assert.equal(
		nextKeyboardContextId(contexts, "company", "previous"),
		"personal",
	);
	assert.equal(nextKeyboardContextId(contexts, "company", "next"), "client-a");
	assert.equal(
		nextKeyboardContextId(contexts, "client-a", "previous"),
		"company",
	);
	assert.equal(nextKeyboardContextId(contexts, "client-a", "next"), "client-a");
});

test("keyboard context navigation handles empty, single, and missing active lists", () => {
	const singleContext = [contextFixture("personal", "Personal")];
	const multipleContexts = [
		contextFixture("personal", "Personal"),
		contextFixture("company", "Company"),
	];

	assert.equal(nextKeyboardContextId([], undefined, "next"), undefined);
	assert.equal(
		nextKeyboardContextId(singleContext, "personal", "next"),
		"personal",
	);
	assert.equal(
		nextKeyboardContextId(singleContext, "personal", "previous"),
		"personal",
	);
	assert.equal(
		nextKeyboardContextId(multipleContexts, undefined, "next"),
		"personal",
	);
	assert.equal(
		nextKeyboardContextId(multipleContexts, "missing", "previous"),
		"personal",
	);
});

test("enter launches only a selected idle base selector", () => {
	assert.equal(
		canLaunchSelectedContextFromKeyboard({
			selectedContextId: "personal",
			launchPending: false,
			mismatchDialogOpen: false,
		}),
		true,
	);
	assert.equal(
		canLaunchSelectedContextFromKeyboard({
			launchPending: false,
			mismatchDialogOpen: false,
		}),
		false,
	);
	assert.equal(
		canLaunchSelectedContextFromKeyboard({
			selectedContextId: "personal",
			launchPending: true,
			mismatchDialogOpen: false,
		}),
		false,
	);
	assert.equal(
		canLaunchSelectedContextFromKeyboard({
			selectedContextId: "personal",
			launchPending: false,
			mismatchDialogOpen: true,
		}),
		false,
	);
});

test("escape cancels the active selector layer", () => {
	assert.equal(
		escapeKeyboardAction({
			launchPending: false,
			mismatchDialogOpen: false,
		}),
		"close-selector",
	);
	assert.equal(
		escapeKeyboardAction({
			selectedContextId: "personal",
			launchPending: false,
			mismatchDialogOpen: true,
		}),
		"close-dialog",
	);
	assert.equal(
		escapeKeyboardAction({
			selectedContextId: "personal",
			launchPending: false,
			mismatchDialogOpen: false,
			dialogOpen: true,
		}),
		"close-dialog",
	);
	assert.equal(
		escapeKeyboardAction({
			selectedContextId: "personal",
			launchPending: true,
			mismatchDialogOpen: true,
		}),
		"none",
	);
});

test("remember control renders unchecked for unbound selected projects", () => {
	const state = launchStateFixture();
	const html = renderToStaticMarkup(
		RememberProjectControl({
			binding: state.binding,
			contexts: state.contexts,
			rememberProject: false,
			selectedContextId: "personal",
		}),
	);

	assert.match(html, /type="checkbox"/);
	assert.match(html, /focus-visible:ring-2/);
	assert.match(html, /aria-checked="false"/);
	assert.doesNotMatch(html, /disabled=""/);
	assert.ok(html.includes("Remember Personal for this project"));
	assert.ok(html.includes("Dev Context will suggest this context next time"));
});

test("remember control renders checked user intent for unbound selected projects", () => {
	const state = launchStateFixture();
	const html = renderToStaticMarkup(
		RememberProjectControl({
			binding: state.binding,
			contexts: state.contexts,
			rememberProject: true,
			selectedContextId: "personal",
		}),
	);

	assert.match(html, /type="checkbox"/);
	assert.match(html, /checked=""/);
	assert.ok(html.includes("Remember Personal for this project"));
});

test("remember control renders existing binding without a checkbox", () => {
	const state = launchStateFixture({
		binding: {
			projectPath: "/work/api",
			bound: true,
			contextId: "company",
			dangling: false,
		},
	});
	const html = renderToStaticMarkup(
		RememberProjectControl({
			binding: state.binding,
			contexts: state.contexts,
			rememberProject: false,
			selectedContextId: "company",
		}),
	);

	assert.doesNotMatch(html, /type="checkbox"/);
	assert.ok(html.includes("Remembered context"));
	assert.ok(html.includes("Company"));
	assert.ok(
		html.includes("will be suggested the next time you open this project"),
	);
	assert.equal(boundContextName(state.binding, state.contexts), "Company");
	assert.equal(canRememberProject(state.binding), false);
});

test("remember control does not treat a dangling binding as unbound", () => {
	const state = launchStateFixture({
		binding: {
			projectPath: "/work/api",
			bound: false,
			dangling: true,
			missingContextId: "company",
		},
	});
	const html = renderToStaticMarkup(
		RememberProjectControl({
			binding: state.binding,
			contexts: state.contexts,
			rememberProject: false,
			selectedContextId: "personal",
		}),
	);

	assert.doesNotMatch(html, /type="checkbox"/);
	assert.ok(html.includes("Remembered context unavailable"));
	assert.ok(html.includes("without changing its remembered context"));
	assert.equal(canRememberProject(state.binding), false);
});

test("dangling binding recovery offers temporary launch, removal, and cancel", () => {
	const html = renderToStaticMarkup(
		DanglingBindingDialog({
			missingContextId: "company",
			pending: false,
			onChooseContext: () => {},
			onRemoveBinding: () => {},
			onCancel: () => {},
		}),
	);

	assert.match(html, /role="dialog"/);
	assert.ok(html.includes("Remembered context unavailable"));
	assert.ok(html.includes("company"));
	assert.ok(html.includes("Choose a context"));
	assert.ok(html.includes("Remove remembered context"));
	assert.ok(html.includes("Cancel"));
});

test("remember control is disabled when no context is selected", () => {
	const state = launchStateFixture();
	const html = renderToStaticMarkup(
		RememberProjectControl({
			binding: state.binding,
			contexts: state.contexts,
			rememberProject: false,
		}),
	);

	assert.match(html, /type="checkbox"/);
	assert.match(html, /disabled=""/);
	assert.ok(html.includes("Select a context before remembering this project."));
	assert.equal(canRememberProject(state.binding), true);
});

test("selector actions disable launch without a selected context", () => {
	const html = renderToStaticMarkup(
		SelectorActions({
			launchDisabled: true,
			launchPending: false,
			onLaunch: () => {},
			onCancel: () => {},
		}),
	);

	assert.ok(html.includes("Launch"));
	assert.ok(html.includes("Cancel"));
	assert.match(html, /disabled=""/);
	assert.match(html, /focus-visible:ring-2/);
});

test("selector actions show launch as enabled and pending", () => {
	const enabled = renderToStaticMarkup(
		SelectorActions({
			launchDisabled: false,
			launchPending: false,
			onLaunch: () => {},
			onCancel: () => {},
		}),
	);
	const pending = renderToStaticMarkup(
		SelectorActions({
			launchDisabled: false,
			launchPending: true,
			onLaunch: () => {},
			onCancel: () => {},
		}),
	);

	assert.doesNotMatch(enabled, /disabled=""/);
	assert.ok(enabled.includes("Launch"));
	assert.match(pending, /disabled=""/);
	assert.ok(pending.includes("Launching..."));
});

test("launch labels name the selected context and pending project", () => {
	assert.equal(launchActionLabel("Company"), "Launch Company");
	assert.equal(
		launchPendingLabel("devctx", "Company"),
		"Launching devctx as Company...",
	);
	assert.equal(
		launchPendingLabel(undefined, "Company"),
		"Launching Company...",
	);
});

test("launch verification progress renders a pending shell, groups, and backend stages", () => {
	const pending = renderToStaticMarkup(
		LaunchVerificationProgress({
			projectName: "devctx",
			contextName: "Company",
		}),
	);
	const staged = renderToStaticMarkup(
		LaunchVerificationProgress({
			projectName: "devctx",
			contextName: "Company",
			steps: [
				{
					id: "prepare_environment",
					label: "Prepare isolated environment",
					status: "ready",
					message: "Environment is ready.",
				},
				{
					id: "check_providers",
					label: "Check enabled providers",
					status: "needs_attention",
					message: "Review provider setup.",
				},
				{
					id: "start_tool",
					label: "Start coding tool",
					status: "pending",
					message: "Waiting to start.",
				},
			],
		}),
	);
	const grouped = renderToStaticMarkup(
		LaunchVerificationProgress({
			projectName: "devctx",
			contextName: "Company",
			groups: [
				{
					id: "project",
					label: "Project",
					status: "ready",
					blocking: false,
					message: "Project folder is ready.",
					checks: [
						{
							id: "project_directory",
							label: "Project folder",
							status: "ready",
							blocking: false,
							message: "Project folder is ready.",
						},
					],
				},
				{
					id: "tools",
					label: "Tools",
					status: "blocked",
					blocking: true,
					message: "VS Code is unavailable.",
					checks: [
						{
							id: "tool_0",
							label: "VS Code",
							status: "blocked",
							blocking: true,
							message: "VS Code is unavailable.",
							actionHint: "Install VS Code.",
						},
					],
				},
			],
		}),
	);

	assert.match(pending, /role="status"/);
	assert.ok(pending.includes("Launching devctx as Company..."));
	assert.ok(pending.includes("Preparing launch stages..."));
	assert.ok(staged.includes("Prepare isolated environment"));
	assert.ok(staged.includes("Needs attention"));
	assert.ok(staged.includes("Pending"));
	assert.ok(grouped.includes("Project folder is ready."));
	assert.ok(grouped.includes("VS Code is unavailable."));
	assert.ok(grouped.includes("Fix required"));
	assert.ok(grouped.includes("Show details"));
	assert.doesNotMatch(grouped, /<details[^>]* open=""/);
	assert.deepEqual(
		["ready", "needs_attention", "blocked", "pending"].map(
			(status) => verificationStepPresentation(status).label,
		),
		["Ready", "Needs attention", "Blocked", "Pending"],
	);
});

test("preflight review permits continuation only for non-blocking checks", () => {
	const preflight = {
		...preflightLaunchProjectResult().data,
		groups: [
			{
				id: "project",
				label: "Project",
				status: "needs_attention",
				blocking: false,
				message: "Review the project binding.",
				checks: [],
			},
		],
	};
	const review = renderToStaticMarkup(
		PreflightReviewView({
			projectName: "api",
			contextName: "Company",
			preflight,
			onFixFirst: () => {},
			onLaunchWithoutIt: () => {},
		}),
	);
	const blocked = renderToStaticMarkup(
		PreflightReviewView({
			projectName: "api",
			contextName: "Company",
			preflight: {
				...preflight,
				groups: [{ ...preflight.groups[0], blocking: true, status: "blocked" }],
			},
			onFixFirst: () => {},
		}),
	);

	assert.ok(review.includes("Ready to launch"));
	assert.ok(review.includes("api will open as Company"));
	assert.ok(review.includes("Fix first"));
	assert.ok(review.includes("Launch without it"));
	assert.ok(blocked.includes("Launch is blocked"));
	assert.ok(blocked.includes("Fix first"));
	assert.ok(!blocked.includes("Launch without it"));
});

test("selector actions block unsafe launches and explain the blocking checks", () => {
	const html = renderToStaticMarkup(
		SelectorActions({
			launchDisabled: true,
			launchPending: false,
			contextName: "Company",
			confidence: {
				contextId: "company",
				status: "blocked",
				checks: [
					{
						component: "tool",
						toolId: "tool",
						severity: "blocked",
						label: "Selected tool",
						message: "The selected tool is unavailable.",
						actionHint: "Install the selected tool.",
					},
				],
			},
			onLaunch: () => {},
			onCancel: () => {},
		}),
	);

	assert.match(html, /role="alert"/);
	assert.ok(html.includes("Launch blocked for Company"));
	assert.ok(html.includes("Selected tool:"));
	assert.ok(html.includes("Install the selected tool."));
	assert.ok(html.includes("Launch Company"));
	assert.match(html, /disabled=""/);
});

test("selector actions make warnings actionable and confirm ready launch state", () => {
	const warning = launchConfidenceFeedback(
		{
			contextId: "company",
			status: "needs_attention",
			checks: [
				{
					component: "provider",
					providerId: "provider",
					severity: "needs_attention",
					label: "Provider",
					message: "Provider needs review.",
					actionHint: "Review provider setup.",
				},
			],
		},
		"Company",
	);
	const ready = launchConfidenceFeedback(
		{ contextId: "company", status: "ready", checks: [] },
		"Company",
	);

	assert.deepEqual(warning, {
		status: "needs_attention",
		title: "Review Company before launch",
		message: "Launch is available, but these items need your attention.",
		checks: [
			{
				component: "provider",
				providerId: "provider",
				severity: "needs_attention",
				label: "Provider",
				message: "Provider needs review.",
				actionHint: "Review provider setup.",
			},
		],
	});
	assert.deepEqual(ready, {
		status: "ready",
		title: "Company is ready to launch",
		message: "Everything required for a safe launch is available.",
		checks: [],
	});
});

test("selector layout keeps project, contexts, confidence, remember, and actions in order", () => {
	const html = renderToStaticMarkup(
		SelectorLayout({
			projectIdentity: "Project identity",
			contextCards: "Context cards",
			confidenceSummary: "Confidence summary",
			rememberControl: "Remember control",
			launchActions: "Launch actions",
		}),
	);

	const sections = [
		'data-selector-layout-section="project-identity"',
		'data-selector-layout-section="context-cards"',
		'data-selector-layout-section="confidence-summary"',
		'data-selector-layout-section="remember-control"',
		'data-selector-layout-section="launch-actions"',
	];
	assert.deepEqual(
		sections.map((section) => html.indexOf(section) >= 0),
		[true, true, true, true, true],
	);
	assert.deepEqual(
		sections.map((section) => html.indexOf(section)),
		[...sections.map((section) => html.indexOf(section))].sort(
			(left, right) => left - right,
		),
	);
});

test("selector confidence summary renders selected context readiness", () => {
	const selected = renderToStaticMarkup(
		SelectorConfidenceSummary({
			project: { name: "api", path: "/work/api" },
			context: {
				...contextFixture("personal", "Personal"),
				confidence: {
					contextId: "personal",
					status: "needs_attention",
					checks: [
						{
							component: "provider",
							providerId: "provider",
							severity: "ready",
							label: "Provider",
							message: "Provider is ready.",
						},
						{
							component: "tool",
							toolId: "vscode",
							severity: "needs_attention",
							label: "VS Code",
							message: "Review VS Code.",
						},
						{
							component: "isolation",
							severity: "ready",
							label: "Context storage",
							message: "Context storage is ready.",
						},
					],
				},
			},
		}),
	);
	const empty = renderToStaticMarkup(SelectorConfidenceSummary({}));

	assert.ok(selected.includes("Launch confidence"));
	assert.ok(selected.includes("Project"));
	assert.ok(selected.includes("api"));
	assert.ok(selected.includes("Personal"));
	assert.ok(selected.includes("Provider"));
	assert.ok(selected.includes("VS Code"));
	assert.ok(selected.includes("Isolation"));
	assert.ok(selected.includes("Protected"));
	assert.ok(selected.includes("Needs attention"));
	assert.ok(empty.includes("Select a context to review launch readiness."));
	assert.deepEqual(
		["ready", "needs_attention", "blocked"].map(
			(status) => confidenceStatusPresentation(status).label,
		),
		["Ready", "Needs attention", "Blocked"],
	);
});

test("selector critical path renders selected context and submits remembered launch", async () => {
	const launchState = launchStateFixture({
		binding: {
			projectPath: "/work/api",
			bound: true,
			contextId: "personal",
			dangling: false,
		},
		selectedContextId: "personal",
		selectionRequired: false,
		resolutionSource: "project_binding",
	});
	const html = [
		renderToStaticMarkup(ProjectIdentity({ project: launchState.project })),
		renderToStaticMarkup(
			ContextCard({
				context: launchState.contexts[0],
				selected: true,
				onSelect: () => {},
			}),
		),
		renderToStaticMarkup(
			RememberProjectControl({
				binding: launchState.binding,
				contexts: launchState.contexts,
				rememberProject: false,
				selectedContextId: "personal",
			}),
		),
		renderToStaticMarkup(
			SelectorActions({
				launchDisabled: false,
				launchPending: true,
				onLaunch: () => {},
				onCancel: () => {},
			}),
		),
	].join("");

	assert.ok(html.includes("/work/api"));
	assert.ok(html.includes("Personal"));
	assert.match(html, /data-selected="true"/);
	assert.ok(html.includes("Remembered context"));
	assert.ok(html.includes("Launching..."));

	const calls = [];
	const selectedContextId = nextSelectedContextId(
		launchState.contexts,
		"company",
	);
	const result = await launchSelectedContext({
		projectPath: launchState.project.path,
		selectedContextId,
		bindingContextId: "company",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, launchProjectResult());
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "company" },
		],
		["bindProject", { projectPath: "/work/api", contextId: "company" }],
		["launchProject", { projectPath: "/work/api", contextId: "company" }],
	]);
});

test("launch action does nothing without a selected context", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.equal(result, undefined);
	assert.deepEqual(calls, []);
});

test("launch action launches the selected context when remember is off", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "personal",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, launchProjectResult());
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "personal" },
		],
		["launchProject", { projectPath: "/work/api", contextId: "personal" }],
	]);
});

test("launch action exposes preflight verification steps before starting the coding tool", async () => {
	const calls = [];
	const verificationSteps = [
		{
			id: "prepare_environment",
			label: "Prepare isolated environment",
			status: "ready",
			message: "Environment is ready.",
		},
	];

	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "personal",
		bindProject: () => Promise.resolve(projectBindingResult()),
		preflightLaunchProject(request) {
			calls.push(["preflight", request]);
			return Promise.resolve({
				ok: true,
				data: { ...preflightLaunchProjectResult().data, verificationSteps },
			});
		},
		onPreflightComplete(preflight) {
			calls.push(["verification", preflight.verificationSteps]);
		},
		launchProject(request) {
			calls.push(["launch", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.equal(result?.ok, true);
	assert.deepEqual(calls, [
		["preflight", { projectPath: "/work/api", contextId: "personal" }],
		["verification", verificationSteps],
		["launch", { projectPath: "/work/api", contextId: "personal" }],
	]);
});

test("launch action preflights before binding when remember is on", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "company",
		bindingContextId: "company",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, launchProjectResult());
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "company" },
		],
		["bindProject", { projectPath: "/work/api", contextId: "company" }],
		["launchProject", { projectPath: "/work/api", contextId: "company" }],
	]);
});

test("launch action pauses before binding when preflight needs review", async () => {
	const calls = [];
	const preflight = {
		...preflightLaunchProjectResult().data,
		groups: [
			{
				id: "tools",
				label: "Tools",
				status: "needs_attention",
				blocking: false,
				message: "Provider setup needs review.",
				checks: [],
			},
		],
	};
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "company",
		bindingContextId: "company",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve({ ok: true, data: preflight });
		},
		onPreflightComplete() {
			return false;
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, { preflightReview: preflight });
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "company" },
		],
	]);
});

test("launch action rejects a binding target that differs from the launch context", async () => {
	const calls = [];

	await assert.rejects(
		launchSelectedContext({
			projectPath: "/work/api",
			selectedContextId: "personal",
			bindingContextId: "company",
			bindProject(request) {
				calls.push(["bindProject", request]);
				return Promise.resolve(projectBindingResult());
			},
			preflightLaunchProject(request) {
				calls.push(["preflightLaunchProject", request]);
				return Promise.resolve(preflightLaunchProjectResult());
			},
			launchProject(request) {
				calls.push(["launchProject", request]);
				return Promise.resolve(launchProjectResult());
			},
		}),
		/Binding context must match the selected launch context/,
	);

	assert.deepEqual(calls, []);
});

test("launch action returns binding errors without launching", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "company",
		bindingContextId: "company",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(
				apiError(
					"validation_error",
					"Unable to complete request.",
					"Check the selected project and context, then retry.",
				),
			);
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(
		result,
		apiError(
			"validation_error",
			"Unable to complete request.",
			"Check the selected project and context, then retry.",
		),
	);
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "company" },
		],
		["bindProject", { projectPath: "/work/api", contextId: "company" }],
	]);
});

test("launch action returns preflight errors without launching", async () => {
	const calls = [];
	const error = apiError(
		"launch_error",
		"Unable to launch editor.",
		"Check the editor command.",
	);
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "personal",
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(error);
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, error);
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{ projectPath: "/work/api", contextId: "personal" },
		],
	]);
});

test("launch action resubmits explicit context mismatch confirmation", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "personal",
		confirmContextMismatch: true,
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, launchProjectResult());
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: true,
			},
		],
		[
			"launchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: true,
			},
		],
	]);
});

test("context mismatch dialog explains and prioritizes the remembered context", () => {
	const html = renderToStaticMarkup(
		ContextMismatchDialog({
			mismatch: {
				projectPath: "/work/api",
				boundContextId: "company",
				requestedContextId: "personal",
			},
			contexts: [
				contextFixture("company", "Company"),
				contextFixture("personal", "Personal"),
			],
			launchPending: false,
			onCancel: () => {},
			onUseRememberedContext: () => {},
			onOpenAnyway: () => {},
		}),
	);

	assert.match(html, /role="dialog"/);
	assert.ok(html.includes("This project normally opens in Company"));
	assert.ok(html.includes("/work/api"));
	assert.ok(html.includes("Company"));
	assert.ok(html.includes("Personal"));
	assert.ok(html.includes("signed-in accounts, tools, and environment"));
	assert.ok(
		html.includes(
			"will not change which context this project normally opens in",
		),
	);
	assert.match(html, /aria-label="Context comparison"/);
	assert.ok(html.includes("Normal choice"));
	assert.ok(html.includes("Open only if intended"));
	assert.ok(html.includes("Cancel"));
	assert.ok(html.includes("Open Personal anyway"));
	assert.ok(html.includes("Use remembered context"));
	assert.match(html, /autofocus=""/);
	assert.ok(
		html.indexOf("Open Personal anyway") <
			html.indexOf("Use remembered context"),
	);
	assert.match(html, /bg-destructive/);
	assert.match(html, /bg-primary/);
	assert.match(html, /focus-visible:ring-2/);
});

test("context mismatch keyboard behavior only cancels with Escape", () => {
	assert.equal(mismatchDialogKeyboardAction("Escape", false), "cancel");
	assert.equal(mismatchDialogKeyboardAction("Enter", false), "none");
	assert.equal(mismatchDialogKeyboardAction("Escape", true), "none");
});

test("context mismatch open anyway pending state disables actions", () => {
	const html = renderToStaticMarkup(
		ContextMismatchDialog({
			mismatch: {
				projectPath: "/work/api",
				boundContextId: "company",
				requestedContextId: "personal",
			},
			contexts: [],
			launchPending: true,
			onCancel: () => {},
			onUseRememberedContext: () => {},
			onOpenAnyway: () => {},
		}),
	);

	assert.ok(html.includes("Opening..."));
	assert.match(html, /disabled=""/);
});

test("successful temporary launches offer an explicit binding replacement", () => {
	const binding = {
		projectPath: "/work/api",
		bound: true,
		contextId: "company",
		dangling: false,
	};
	assert.deepEqual(bindingReplacementForLaunch(binding, "personal"), {
		boundContextId: "company",
		replacementContextId: "personal",
	});
	assert.equal(bindingReplacementForLaunch(binding, "company"), undefined);
	assert.equal(
		bindingReplacementForLaunch({ ...binding, dangling: true }, "personal"),
		undefined,
	);

	const html = renderToStaticMarkup(
		ReplaceBindingDialog({
			boundContextName: "Company",
			replacementContextName: "Personal",
			pending: false,
			onKeepCurrent: () => {},
			onReplace: () => {},
		}),
	);

	assert.match(html, /role="dialog"/);
	assert.ok(html.includes("still remembered for Company"));
	assert.ok(html.includes("Keep Company"));
	assert.ok(html.includes("Remember Personal"));
});

test("context mismatch cancel exits without launch side effects", () => {
	const calls = [];
	const props = {
		mismatch: {
			projectPath: "/work/api",
			boundContextId: "company",
			requestedContextId: "personal",
		},
		contexts: [
			contextFixture("company", "Company"),
			contextFixture("personal", "Personal"),
		],
		launchPending: false,
		onCancel: () => calls.push("cancel"),
		onUseRememberedContext: () => calls.push("useRememberedContext"),
		onOpenAnyway: () => calls.push("launchProject"),
	};
	const html = renderToStaticMarkup(ContextMismatchDialog(props));

	assert.ok(html.includes("Cancel"));
	assert.ok(html.includes("Open Personal anyway"));
	assert.ok(html.includes("Use remembered context"));
	props.onCancel();
	assert.deepEqual(calls, ["cancel"]);
});

test("context mismatch open anyway submits exactly one confirmed launch", async () => {
	const calls = [];
	const result = await launchSelectedContext({
		projectPath: "/work/api",
		selectedContextId: "personal",
		confirmContextMismatch: true,
		bindProject(request) {
			calls.push(["bindProject", request]);
			return Promise.resolve(projectBindingResult());
		},
		preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return Promise.resolve(preflightLaunchProjectResult());
		},
		launchProject(request) {
			calls.push(["launchProject", request]);
			return Promise.resolve(launchProjectResult());
		},
	});

	assert.deepEqual(result, launchProjectResult());
	assert.deepEqual(calls, [
		[
			"preflightLaunchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: true,
			},
		],
		[
			"launchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: true,
			},
		],
	]);
});

test("gui error notice renders failure and recovery guidance", () => {
	const errors = [
		apiError(
			"validation_error",
			"Unable to complete request.",
			"Check the selected project and context, then retry.",
		),
		apiError(
			"launch_error",
			"Unable to launch editor.",
			"Check the editor command, project path, and permissions, then retry.",
		),
		apiError(
			"internal_error",
			"Dev Context failed unexpectedly.",
			"Retry the action. If it keeps failing, include debug details in a bug report.",
		),
		apiError(
			"unexpected_error",
			"Wails bridge unavailable",
			"Retry the action. If it keeps failing, include the error details in a bug report.",
		),
	];

	for (const error of errors) {
		const html = renderToStaticMarkup(GuiErrorNotice({ error: error.error }));

		assert.match(html, /role="alert"/);
		assert.ok(html.includes("What happened"));
		assert.ok(html.includes("Why it matters"));
		assert.ok(html.includes("What to do"));
		assert.ok(html.includes(error.error.message));
		assert.ok(html.includes(error.error.recovery));
	}
});

test("launch success close behavior defaults to closing the selector", () => {
	assert.equal(defaultLaunchSuccessCloseBehavior, "close_selector");
	assert.equal(shouldCloseSelectorAfterLaunch("keep_open"), false);
	assert.equal(shouldCloseSelectorAfterLaunch("close_selector"), true);
});

test("launch progress is a focused surface without launcher controls", () => {
	const html = renderToStaticMarkup(
		LaunchProgressView({
			projectName: "api",
			contextName: "Company",
			showVerification: true,
			steps: [
				{
					id: "start_tool",
					label: "Start VS Code",
					status: "pending",
					message: "Starting VS Code.",
				},
			],
		}),
	);

	assert.match(html, /aria-label="Launching project"/);
	assert.ok(html.includes("Opening api"));
	assert.ok(html.includes("Using Company"));
	assert.ok(html.includes("Start VS Code"));
	assert.doesNotMatch(html, /Launch Company|Choose another|Remember this project/);
});

test("launch failure view keeps recovery actions available without exposing technical details", () => {
	const error = apiError(
		"launch_error",
		"Unable to launch editor.",
		"Check the editor command, project path, and permissions, then retry.",
	).error;
	const html = renderToStaticMarkup(
		LaunchFailureView({
			error,
			onRetry: () => {},
			onRunDiagnostics: () => {},
			onChooseAnotherContext: () => {},
			onCancel: () => {},
		}),
	);

	assert.match(html, /role="alert"/);
	assert.ok(html.includes("Dev Context is still open"));
	assert.ok(html.includes("Retry"));
	assert.ok(html.includes("Run diagnostics"));
	assert.ok(html.includes("Choose another context"));
	assert.ok(html.includes("Cancel"));
	assert.doesNotMatch(html, /Technical details/);
});

test("launch failure view hides technical details until requested", () => {
	const error = {
		...apiError(
			"launch_error",
			"Unable to launch editor.",
			"Check the editor command, then retry.",
		).error,
		launchFailureDetails: {
			executable: "/usr/bin/code",
			exitCode: 1,
			timestamp: "2026-09-05T10:30:00Z",
			logs: "starting tool in /work/api failed: permission denied",
		},
	};
	const html = renderToStaticMarkup(
		LaunchFailureView({
			error,
			onRetry: () => {},
			onChooseAnotherContext: () => {},
			onCancel: () => {},
		}),
	);

	assert.ok(html.includes("Technical details"));
	assert.ok(html.includes("Executable"));
	assert.ok(html.includes("Exit code"));
	assert.ok(html.includes("Timestamp"));
	assert.ok(html.includes("/usr/bin/code"));
	assert.ok(html.includes("/work/api"));
	assert.match(html, /<details/);
	assert.doesNotMatch(html, /<details(?:\s[^>]*)?\sopen(?:=|\s|>)/);
});

test("launch progress guard allows only one in-flight launch and restores after rejection", async () => {
	const guard = createLaunchRequestGuard();
	const deferred = createDeferred();
	const calls = [];

	const first = guard.run(async () => {
		calls.push("first");
		await deferred.promise;
		return "done";
	});
	const duplicate = await guard.run(async () => {
		calls.push("duplicate");
		return "duplicate";
	});

	assert.equal(duplicate, undefined);
	assert.deepEqual(calls, ["first"]);

	deferred.reject(new Error("launch failed"));
	await assert.rejects(first, /launch failed/);

	const afterFailure = await guard.run(async () => {
		calls.push("afterFailure");
		return "restored";
	});

	assert.equal(afterFailure, "restored");
	assert.deepEqual(calls, ["first", "afterFailure"]);
});

test("cancel closes the selector without launch or binding side effects", async () => {
	const calls = [];
	const window = createDevContextWindow({
		quit() {
			calls.push("closeSelector");
		},
	});

	await cancelSelector({
		closeSelector: () => window.closeSelector(),
		launchProject() {
			calls.push("launchProject");
		},
		bindProject() {
			calls.push("bindProject");
		},
	});

	assert.deepEqual(calls, ["closeSelector"]);

	const blocked = await cancelSelector({
		canCancel: false,
		closeSelector: () => window.closeSelector(),
	});
	assert.equal(blocked, false);
	assert.deepEqual(calls, ["closeSelector"]);
});

test("context creation refreshes launch state after success", async () => {
	const calls = [];
	const refreshedState = launchStateFixture({
		contexts: [contextFixture("personal", "Personal")],
		firstRun: false,
	});

	const request = {
		contextId: "personal",
		name: "Personal",
		importProviderIds: ["codex"],
	};
	const result = await createContextAndRefresh({
		request,
		createContext(receivedRequest) {
			calls.push(["createContext", receivedRequest]);
			return Promise.resolve({
				ok: true,
				data: {
					context: contextFixture("personal", "Personal"),
				},
			});
		},
		getLaunchState() {
			calls.push(["getLaunchState"]);
			return Promise.resolve({
				ok: true,
				data: refreshedState,
			});
		},
	});

	assert.deepEqual(result, {
		ok: true,
		created: {
			context: contextFixture("personal", "Personal"),
		},
		launchState: refreshedState,
	});
	assert.deepEqual(calls, [
		["createContext", request],
		["getLaunchState"],
	]);
});

test("context creation returns failures without refreshing", async () => {
	const calls = [];
	const error = apiError(
		"validation_error",
		"Unable to complete request.",
		"Check the selected project and context, then retry.",
	);

	const request = { contextId: "company" };
	const result = await createContextAndRefresh({
		request,
		createContext(receivedRequest) {
			calls.push(["createContext", receivedRequest]);
			return Promise.resolve(error);
		},
		getLaunchState() {
			calls.push(["getLaunchState"]);
			return Promise.resolve({ ok: true, data: launchStateFixture() });
		},
	});

	assert.deepEqual(result, {
		ok: false,
		error: error.error,
	});
	assert.deepEqual(calls, [["createContext", request]]);
});

test("context creation flow preserves its draft across forward and back steps", () => {
	let flow = initialContextCreateFlow({ name: "Personal" });
	flow = updateContextCreateDraft(flow, {
		description: "Personal repositories and experiments",
		enabledProviderIds: ["codex"],
	});
	flow = nextContextCreateStep(flow);
	flow = nextContextCreateStep(flow);
	flow = previousContextCreateStep(flow);

	assert.equal(flow.status, "projects");
	assert.deepEqual(flow.draft, {
		name: "Personal",
		description: "Personal repositories and experiments",
		enabledProviderIds: ["codex"],
	});
});

test("context creation flow permits creation only from review", () => {
	let flow = initialContextCreateFlow();
	assert.equal(beginContextCreation(flow), flow);

	flow = nextContextCreateStep(flow);
	flow = nextContextCreateStep(flow);
	flow = nextContextCreateStep(flow);
	flow = beginContextCreation(flow);
	assert.equal(flow.status, "creating");

	flow = returnToContextCreateReview(flow);
	assert.equal(flow.status, "review");
	flow = beginContextCreation(flow);
	flow = completeContextCreation(flow);
	assert.equal(flow.status, "success");
});

function contextFixture(id, name, providers = []) {
	return {
		id,
		name,
		tool: {
			id: "vscode",
			name: "VS Code",
			status: "ready",
			message: "VS Code is available for launch.",
		},
		providers,
	};
}

function providerFixture(id, name, enabled, state, explanation) {
	return {
		id,
		name,
		enabled,
		state,
		explanation,
		identity:
			enabled && state === "ready"
				? {
						status: "unavailable",
						message: "Account identity unavailable.",
						fields: [],
					}
				: { status: "none", fields: [] },
	};
}

function launchStateFixture(overrides = {}) {
	return {
		project: { name: "api", path: "/work/api" },
		contexts: [
			contextFixture("personal", "Personal"),
			contextFixture("company", "Company"),
			contextFixture("client-a", "Client A"),
		],
		binding: {
			projectPath: "/work/api",
			bound: false,
			dangling: false,
		},
		selectionRequired: true,
		warnings: [],
		firstRun: false,
		providerCredentialSessions: [],
		...overrides,
	};
}

function providerCredentialSessionsFixture() {
	return [
		{
			providerId: "codex",
			name: "Codex",
			metadataAvailable: true,
			fields: [
				{ label: "Email", value: "user@company.com" },
				{ label: "Plan", value: "Business" },
				{ label: "Account", value: "acct_123" },
			],
		},
		{
			providerId: "claude",
			name: "Claude",
			metadataAvailable: true,
			fields: [
				{ label: "Subscription", value: "Pro" },
				{ label: "Organization", value: "Jishin Labs" },
				{ label: "Organization ID", value: "e783" },
			],
		},
	];
}

function projectBindingResult() {
	return {
		ok: true,
		data: {
			projectPath: "/work/api",
			bound: true,
			contextId: "personal",
			dangling: false,
		},
	};
}

function launchProjectResult() {
	return {
		ok: true,
		data: {
			project: { name: "api", path: "/work/api" },
			context: contextFixture("personal", "Personal"),
			warnings: [],
		},
	};
}

function preflightLaunchProjectResult() {
	return {
		ok: true,
		data: {
			project: { name: "api", path: "/work/api" },
			context: contextFixture("personal", "Personal"),
			confidence: {
				contextId: "personal",
				status: "ready",
				checks: [],
			},
			warnings: [],
		},
	};
}

function apiError(code, message, recovery, contextMismatch) {
	return {
		ok: false,
		error: {
			code,
			message,
			recovery,
			contextMismatch,
		},
	};
}

function createDeferred() {
	let resolve;
	let reject;
	const promise = new Promise((promiseResolve, promiseReject) => {
		resolve = promiseResolve;
		reject = promiseReject;
	});

	return { promise, resolve, reject };
}

function escapeRegExp(value) {
	return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
