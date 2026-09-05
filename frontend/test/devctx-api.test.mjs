import assert from "node:assert/strict";
import test from "node:test";

import { createDevContextApi } from "../.tmp-test/src/lib/devctx-api.js";

const toolFixture = () => ({
	id: "second-tool",
	name: "Second Tool",
	status: "ready",
	message: "Second Tool is available for launch.",
});

const toolOptionFixture = () => ({ id: "second-tool", name: "Second Tool" });

test("adapter exposes the host-selected application mode", async () => {
	const api = createDevContextApi({
		async getApplicationMode() {
			return { type: "launcher", projectPath: "/work/api" };
		},
	});

	assert.deepEqual(await api.getApplicationMode(), {
		ok: true,
		data: { type: "launcher", projectPath: "/work/api" },
	});
});

test("adapter rejects an invalid application mode", async () => {
	const api = createDevContextApi({
		async getApplicationMode() {
			return { type: "launcher" };
		},
	});

	const result = await api.getApplicationMode();
	assert.equal(result.ok, false);
	if (!result.ok) {
		assert.equal(result.error.code, "unexpected_error");
	}
});

test("adapter preserves the typed project-path recovery category", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			return {
				code: "validation_error",
				message: "Project path does not exist.",
				recovery: "Choose an existing project directory.",
				projectPathIssue: "not_found",
			};
		},
	});

	const result = await api.getLaunchState({ projectPath: "/missing/project" });
	assert.equal(result.ok, false);
	if (!result.ok) {
		assert.equal(result.error.projectPathIssue, "not_found");
	}
});

test("adapter returns the selected project directory or a canceled selection", async () => {
	const selectedApi = createDevContextApi({
		async chooseProjectDirectory() {
			return "/work/recovered-project";
		},
	});
	const canceledApi = createDevContextApi({
		async chooseProjectDirectory() {
			return "";
		},
	});

	assert.deepEqual(await selectedApi.chooseProjectDirectory(), {
		ok: true,
		data: "/work/recovered-project",
	});
	assert.deepEqual(await canceledApi.chooseProjectDirectory(), {
		ok: true,
		data: undefined,
	});
});

test("adapter normalizes the Home dashboard contract", async () => {
	const api = createDevContextApi({
		async getHomeDashboard(request) {
			assert.deepEqual(request, { projectPath: "/work/api" });
			return {
				project: { name: "api", path: "/work/api" },
				currentContext: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					confidence: { contextId: "personal", status: "ready", checks: [] },
				},
				recentProjects: [],
				running: { count: 0, contextCounts: [], isolationProtected: false },
				activity: { count: 0 },
			};
		},
	});

	assert.deepEqual(await api.getHomeDashboard({ projectPath: "/work/api" }), {
		ok: true,
		data: {
			project: { name: "api", path: "/work/api" },
			currentContext: {
				id: "personal",
				name: "Personal",
				tool: toolFixture(),
				confidence: { contextId: "personal", status: "ready", checks: [] },
			},
			recentProjects: [],
			running: { count: 0, contextCounts: [], isolationProtected: false },
			activity: { count: 0 },
		},
	});
});

test("adapter normalizes safe context metadata export and import", async () => {
	const exported = {
		version: 1,
		context: {
			name: "Personal",
			metadata: { accent: "sage" },
			providers: [{ id: "fake", enabled: true, options: { region: "south" } }],
			launchTarget: {
				defaultTool: "second-tool",
				tools: [{ id: "second-tool", options: { profile: "personal" } }],
			},
		},
	};
	const api = createDevContextApi({
		async exportContextMetadata(request) {
			assert.deepEqual(request, { contextId: "personal" });
			return exported;
		},
		async importContextMetadata(request) {
			assert.deepEqual(request, { contextId: "imported", export: exported });
			return {
				context: {
					id: "imported",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [],
					providers: [],
					confidence: { contextId: "imported", status: "ready", checks: [] },
				},
			};
		},
	});

	assert.deepEqual(await api.exportContextMetadata({ contextId: "personal" }), {
		ok: true,
		data: exported,
	});
	assert.deepEqual(
		await api.importContextMetadata({
			contextId: "imported",
			export: exported,
		}),
		{
			ok: true,
			data: {
				context: {
					id: "imported",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [],
					providers: [],
					confidence: { contextId: "imported", status: "ready", checks: [] },
					metadata: undefined,
				},
			},
		},
	);
});

test("adapter normalizes Trust Center protection boundaries", async () => {
	const api = createDevContextApi({
		async getTrustCenter() {
			return {
				contexts: [
					{
						id: "personal",
						name: "Personal",
						providers: [],
						tool: {
							id: "second-tool",
							name: "Second Tool",
							isolation: { status: "ready", message: "Ready" },
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
				credentialSync: { enabled: false, message: "Credentials stay local." },
				integrationBoundaries: [
					{
						toolId: "second-tool",
						toolName: "Second Tool",
						statusDataAvailable: false,
						message: "No integration data.",
					},
				],
			};
		},
	});
	const result = await api.getTrustCenter();
	assert.equal(result.ok, true);
	if (!result.ok) return;
	assert.equal(result.data.contexts[0].tool.isolation.status, "ready");
	assert.equal(result.data.projectMappings[0].contextName, "Personal");
	assert.equal(result.data.credentialSync.enabled, false);
});

test("adapter accepts backend-owned account identity confidence checks", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			return {
				project: { name: "api", path: "/work/api" },
				contexts: [
					{
						id: "company",
						name: "Company",
						tool: toolFixture(),
						availableTools: [],
						providers: [],
						confidence: {
							contextId: "company",
							status: "needs_attention",
							checks: [
								{
									component: "identity",
									severity: "needs_attention",
									label: "Account identity",
									message:
										"Verified provider email identities do not match for this context.",
									actionHint:
										"Review provider account configuration before launch.",
								},
							],
						},
					},
				],
				binding: { projectPath: "/work/api", bound: false, dangling: false },
				selectionRequired: false,
				firstRun: false,
			};
		},
	});

	const state = await api.getLaunchState();
	assert.equal(state.ok, true);
	if (!state.ok) return;
	assert.equal(
		state.data.contexts[0].confidence.checks[0].component,
		"identity",
	);
});

test("adapter normalizes active running environments", async () => {
	const api = createDevContextApi({
		async getRunningEnvironments() {
			return {
				environments: [
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
				],
			};
		},
	});

	assert.deepEqual(await api.getRunningEnvironments(), {
		ok: true,
		data: {
			environments: [
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
			],
		},
	});
});

test("adapter normalizes recent projects", async () => {
	const api = createDevContextApi({
		async getRecentProjects() {
			return {
				projects: [
					{
						project: { name: "api", path: "/work/api" },
						contextId: "personal",
						contextName: "Personal",
						lastLaunchedAt: "2026-08-13T12:30:00Z",
					},
				],
			};
		},
	});

	assert.deepEqual(await api.getRecentProjects(), {
		ok: true,
		data: {
			projects: [
				{
					project: { name: "api", path: "/work/api" },
					contextId: "personal",
					contextName: "Personal",
					lastLaunchedAt: "2026-08-13T12:30:00Z",
				},
			],
		},
	});
});

test("adapter normalizes context list summaries", async () => {
	const api = createDevContextApi({
		async getContexts() {
			return {
				contexts: [
					{
						context: {
							id: "personal",
							name: "Personal",
							tool: toolFixture(),
							availableTools: [toolOptionFixture()],
							providers: [],
							confidence: {
								contextId: "personal",
								status: "ready",
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
					},
				],
			};
		},
	});

	const result = await api.getContexts();
	assert.equal(result.ok, true);
	assert.deepEqual(result.ok ? result.data.contexts[0] : undefined, {
		context: {
			id: "personal",
			name: "Personal",
			tool: toolFixture(),
			availableTools: [toolOptionFixture()],
			providers: [],
			confidence: { contextId: "personal", status: "ready", checks: [] },
			metadata: undefined,
		},
		enabledProviders: [
			{
				id: "provider",
				name: "Provider",
				enabled: true,
				state: "ready",
				explanation: undefined,
				identity: { status: "none", message: undefined, fields: [] },
			},
		],
		projectCount: 2,
		lastUsedAt: "2026-08-28T10:30:00Z",
	});
});

test("adapter normalizes context details", async () => {
	const api = createDevContextApi({
		async getContextDetails(request) {
			assert.deepEqual(request, { contextId: "personal" });
			return {
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: { contextId: "personal", status: "ready", checks: [] },
				},
				location: "/contexts/personal",
				createdAt: "2026-08-01T10:30:00Z",
				projectCount: 2,
				enabledProviders: [],
			};
		},
	});

	assert.deepEqual(await api.getContextDetails({ contextId: "personal" }), {
		ok: true,
		data: {
			context: {
				id: "personal",
				name: "Personal",
				tool: toolFixture(),
				availableTools: [toolOptionFixture()],
				providers: [],
				confidence: { contextId: "personal", status: "ready", checks: [] },
				metadata: undefined,
			},
			location: "/contexts/personal",
			createdAt: "2026-08-01T10:30:00Z",
			projectCount: 2,
			enabledProviders: [],
		},
	});
});

test("adapter normalizes successful Wails calls", async () => {
	const calls = [];
	const api = createDevContextApi({
		async getLaunchState(request) {
			calls.push(["getLaunchState", request]);
			return {
				project: { name: "api", path: "/work/api" },
				contexts: [
					{
						id: "personal",
						name: "Personal",
						tool: toolFixture(),
						availableTools: [toolOptionFixture()],
						providers: [
							{
								id: "codex",
								name: "Codex",
								enabled: true,
								state: "ready",
								setupAction: {
									state: "verified",
									label: "Verified",
									message:
										"Codex account identity is verified for this context.",
								},
								identity: {
									status: "verified",
									fields: [
										{ label: "Email", value: "user@company.com" },
										{ label: "Plan", value: "Business" },
										{ label: "Account", value: "acct_123" },
									],
								},
							},
							{
								id: "claude",
								name: "Claude",
								enabled: true,
								state: "ready",
								identity: {
									status: "verified",
									fields: [
										{ label: "Subscription", value: "Pro" },
										{ label: "Organization UUID", value: "e783" },
										{ label: "Organization", value: "Jishin Labs" },
									],
								},
							},
							{
								id: "internal",
								name: "Internal Tool",
								enabled: true,
								state: "ready",
								identity: {
									status: "mismatch_evidence",
									message: "Different account identity detected.",
									fields: [],
								},
							},
						],
						confidence: {
							contextId: "personal",
							status: "needs_attention",
							checks: [
								{
									component: "provider",
									providerId: "codex",
									severity: "needs_attention",
									label: "Codex",
									message: "Codex is not authenticated for this context.",
									actionHint: "Sign in to Codex.",
								},
							],
						},
						metadata: { accent: "blue" },
					},
				],
				binding: {
					projectPath: "/work/api",
					bound: true,
					contextId: "personal",
					dangling: false,
				},
				confidence: {
					contextId: "personal",
					status: "needs_attention",
					checks: [
						{
							component: "provider",
							providerId: "codex",
							severity: "needs_attention",
							label: "Codex",
							message: "Codex is not authenticated for this context.",
							actionHint: "Sign in to Codex.",
						},
						{
							component: "tool",
							toolId: "vscode",
							severity: "ready",
							label: "VS Code",
							message: "VS Code is available for launch.",
						},
					],
				},
				selectedContextId: "personal",
				selectionRequired: false,
				resolutionSource: "project_binding",
				firstRun: false,
				providerCredentialSessions: [
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
							{ label: "Organization UUID", value: "e783" },
							{ label: "Organization", value: "Jishin Labs" },
						],
					},
				],
			};
		},
		async launchProject(request) {
			calls.push(["launchProject", request]);
			return {
				project: { name: "api", path: "/work/api" },
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
				},
			};
		},
		async preflightLaunchProject(request) {
			calls.push(["preflightLaunchProject", request]);
			return {
				project: { name: "api", path: "/work/api" },
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
				},
				confidence: {
					contextId: "personal",
					status: "ready",
					checks: [],
				},
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
				],
				verificationSteps: [
					{
						id: "prepare_environment",
						label: "Prepare isolated environment",
						status: "ready",
						message: "Prepare isolated environment is ready.",
					},
					{
						id: "start_tool",
						label: "Start VS Code",
						status: "pending",
						message: "VS Code will start after launch verification completes.",
					},
				],
			};
		},
		async bindProject(request) {
			calls.push(["bindProject", request]);
			return {
				projectPath: "/work/api",
				bound: true,
				contextId: "personal",
				dangling: false,
			};
		},
		async unbindProject(request) {
			calls.push(["unbindProject", request]);
			return {
				projectPath: "/work/api",
				bound: false,
				dangling: false,
			};
		},
		async createContext(request) {
			calls.push(["createContext", request]);
			return {
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
				},
			};
		},
	});

	assert.deepEqual(await api.getLaunchState({ projectPath: "/work/api" }), {
		ok: true,
		data: {
			project: { name: "api", path: "/work/api" },
			contexts: [
				{
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [
						{
							id: "codex",
							name: "Codex",
							enabled: true,
							state: "ready",
							explanation: undefined,
							setupAction: {
								state: "verified",
								label: "Verified",
								message: "Codex account identity is verified for this context.",
							},
							identity: {
								status: "verified",
								message: undefined,
								fields: [
									{ label: "Email", value: "user@company.com" },
									{ label: "Plan", value: "Business" },
									{ label: "Account", value: "acct_123" },
								],
							},
						},
						{
							id: "claude",
							name: "Claude",
							enabled: true,
							state: "ready",
							explanation: undefined,
							identity: {
								status: "verified",
								message: undefined,
								fields: [
									{ label: "Subscription", value: "Pro" },
									{ label: "Organization UUID", value: "e783" },
									{ label: "Organization", value: "Jishin Labs" },
								],
							},
						},
						{
							id: "internal",
							name: "Internal Tool",
							enabled: true,
							state: "ready",
							explanation: undefined,
							identity: {
								status: "mismatch_evidence",
								message: "Different account identity detected.",
								fields: [],
							},
						},
					],
					confidence: {
						contextId: "personal",
						status: "needs_attention",
						checks: [
							{
								component: "provider",
								providerId: "codex",
								severity: "needs_attention",
								label: "Codex",
								message: "Codex is not authenticated for this context.",
								actionHint: "Sign in to Codex.",
							},
						],
					},
					metadata: { accent: "blue" },
				},
			],
			binding: {
				projectPath: "/work/api",
				bound: true,
				contextId: "personal",
				dangling: false,
				missingContextId: undefined,
				recovery: undefined,
			},
			confidence: {
				contextId: "personal",
				status: "needs_attention",
				checks: [
					{
						component: "provider",
						providerId: "codex",
						severity: "needs_attention",
						label: "Codex",
						message: "Codex is not authenticated for this context.",
						actionHint: "Sign in to Codex.",
					},
					{
						component: "tool",
						toolId: "vscode",
						severity: "ready",
						label: "VS Code",
						message: "VS Code is available for launch.",
						actionHint: undefined,
					},
				],
			},
			selectedContextId: "personal",
			selectionRequired: false,
			resolutionSource: "project_binding",
			warnings: [],
			firstRun: false,
			providerCredentialSessions: [
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
						{ label: "Organization UUID", value: "e783" },
						{ label: "Organization", value: "Jishin Labs" },
					],
				},
			],
		},
	});

	assert.deepEqual(
		await api.preflightLaunchProject({
			projectPath: "/work/api",
			contextId: "personal",
		}),
		{
			ok: true,
			data: {
				project: { name: "api", path: "/work/api" },
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
					metadata: undefined,
				},
				confidence: {
					contextId: "personal",
					status: "ready",
					checks: [],
				},
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
								actionHint: undefined,
							},
						],
					},
				],
				verificationSteps: [
					{
						id: "prepare_environment",
						label: "Prepare isolated environment",
						status: "ready",
						message: "Prepare isolated environment is ready.",
					},
					{
						id: "start_tool",
						label: "Start VS Code",
						status: "pending",
						message: "VS Code will start after launch verification completes.",
					},
				],
				warnings: [],
			},
		},
	);
	assert.deepEqual(
		await api.launchProject({
			projectPath: "/work/api",
			contextId: "personal",
		}),
		{
			ok: true,
			data: {
				project: { name: "api", path: "/work/api" },
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
					metadata: undefined,
				},
				warnings: [],
			},
		},
	);
	assert.deepEqual(
		await api.bindProject({ projectPath: "/work/api", contextId: "personal" }),
		{
			ok: true,
			data: {
				projectPath: "/work/api",
				bound: true,
				contextId: "personal",
				dangling: false,
				missingContextId: undefined,
				recovery: undefined,
			},
		},
	);
	assert.deepEqual(await api.unbindProject({ projectPath: "/work/api" }), {
		ok: true,
		data: {
			projectPath: "/work/api",
			bound: false,
			contextId: undefined,
			dangling: false,
			missingContextId: undefined,
			recovery: undefined,
		},
	});
	assert.deepEqual(
		await api.createContext({
			contextId: "personal",
			importProviderIds: ["codex"],
		}),
		{
			ok: true,
			data: {
				context: {
					id: "personal",
					name: "Personal",
					tool: toolFixture(),
					availableTools: [toolOptionFixture()],
					providers: [],
					confidence: {
						contextId: "personal",
						status: "ready",
						checks: [],
					},
					metadata: undefined,
				},
			},
		},
	);
	assert.deepEqual(calls, [
		["getLaunchState", { projectPath: "/work/api" }],
		[
			"preflightLaunchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: false,
			},
		],
		[
			"launchProject",
			{
				projectPath: "/work/api",
				contextId: "personal",
				confirmContextMismatch: false,
			},
		],
		["bindProject", { projectPath: "/work/api", contextId: "personal" }],
		["unbindProject", { projectPath: "/work/api" }],
		["createContext", { contextId: "personal", importProviderIds: ["codex"] }],
	]);
});

test("adapter normalizes resolved application errors", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			throw new Error("not used");
		},
		async launchProject() {
			return {
				code: "context_mismatch_requires_confirmation",
				message: "Context mismatch requires confirmation.",
				recovery: "Confirm the mismatch intentionally.",
				contextMismatch: {
					projectPath: "/work/api",
					boundContextId: "company",
					requestedContextId: "personal",
				},
			};
		},
		async preflightLaunchProject() {
			return {
				code: "context_mismatch_requires_confirmation",
				message: "Context mismatch requires confirmation.",
				recovery: "Confirm the mismatch intentionally.",
				contextMismatch: {
					projectPath: "/work/api",
					boundContextId: "company",
					requestedContextId: "personal",
				},
			};
		},
		async bindProject() {
			return {
				code: "future_code",
				message: "Future failure.",
				recovery: "Retry later.",
			};
		},
		async unbindProject() {
			throw new Error("not used");
		},
		async createContext() {
			return {
				code: "validation_error",
				message: "Unable to complete request.",
				recovery: "Check the selected project and context, then retry.",
			};
		},
	});

	assert.deepEqual(
		await api.launchProject({
			projectPath: "/work/api",
			contextId: "personal",
		}),
		{
			ok: false,
			error: {
				code: "context_mismatch_requires_confirmation",
				message: "Context mismatch requires confirmation.",
				recovery: "Confirm the mismatch intentionally.",
				contextMismatch: {
					projectPath: "/work/api",
					boundContextId: "company",
					requestedContextId: "personal",
				},
			},
		},
	);

	assert.deepEqual(
		await api.bindProject({ projectPath: "/work/api", contextId: "personal" }),
		{
			ok: false,
			error: {
				code: "unexpected_error",
				message: "Future failure.",
				recovery: "Retry later.",
				contextMismatch: undefined,
			},
		},
	);

	assert.deepEqual(await api.createContext({ contextId: "personal" }), {
		ok: false,
		error: {
			code: "validation_error",
			message: "Unable to complete request.",
			recovery: "Check the selected project and context, then retry.",
			contextMismatch: undefined,
		},
	});
});

test("adapter unwraps tuple-shaped Wails responses", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			return [
				{
					project: { name: "api", path: "/work/api" },
					contexts: [],
					binding: {
						projectPath: "/work/api",
						bound: false,
						dangling: false,
					},
					selectionRequired: true,
					firstRun: true,
				},
				null,
			];
		},
		async launchProject() {
			return [
				{},
				{
					code: "launch_error",
					message: "Unable to launch editor.",
					recovery: "Check the editor command.",
				},
			];
		},
		async preflightLaunchProject() {
			return [
				{},
				{
					code: "launch_error",
					message: "Unable to launch editor.",
					recovery: "Check the editor command.",
				},
			];
		},
		async bindProject() {
			throw new Error("not used");
		},
		async unbindProject() {
			throw new Error("not used");
		},
		async createContext() {
			throw new Error("not used");
		},
	});

	assert.deepEqual(await api.getLaunchState(), {
		ok: true,
		data: {
			project: { name: "api", path: "/work/api" },
			contexts: [],
			binding: {
				projectPath: "/work/api",
				bound: false,
				contextId: undefined,
				dangling: false,
				missingContextId: undefined,
				recovery: undefined,
			},
			confidence: undefined,
			selectedContextId: undefined,
			selectionRequired: true,
			resolutionSource: undefined,
			warnings: [],
			firstRun: true,
			providerCredentialSessions: [],
		},
	});

	assert.deepEqual(await api.launchProject({ contextId: "personal" }), {
		ok: false,
		error: {
			code: "launch_error",
			message: "Unable to launch editor.",
			recovery: "Check the editor command.",
			contextMismatch: undefined,
		},
	});
});

test("adapter normalizes rejected promises into displayable errors", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			throw new Error("Wails bridge unavailable");
		},
		async launchProject() {
			throw {
				code: "launch_error",
				message: "Unable to launch editor.",
				recovery: "Check the editor command.",
			};
		},
		async preflightLaunchProject() {
			throw {
				code: "launch_error",
				message: "Unable to launch editor.",
				recovery: "Check the editor command.",
			};
		},
		async bindProject() {
			throw "request failed";
		},
		async unbindProject() {
			throw null;
		},
		async createContext() {
			throw new Error("create failed");
		},
	});

	assert.deepEqual(await api.getLaunchState(), {
		ok: false,
		error: {
			code: "unexpected_error",
			message: "Wails bridge unavailable",
			recovery:
				"Retry the action. If it keeps failing, include the error details in a bug report.",
		},
	});

	assert.deepEqual(await api.launchProject({ contextId: "personal" }), {
		ok: false,
		error: {
			code: "launch_error",
			message: "Unable to launch editor.",
			recovery: "Check the editor command.",
			contextMismatch: undefined,
		},
	});

	assert.deepEqual(await api.bindProject({ contextId: "personal" }), {
		ok: false,
		error: {
			code: "unexpected_error",
			message: "request failed",
			recovery:
				"Retry the action. If it keeps failing, include the error details in a bug report.",
		},
	});

	assert.deepEqual(await api.unbindProject(), {
		ok: false,
		error: {
			code: "unexpected_error",
			message: "Dev Context could not complete the request.",
			recovery:
				"Retry the action. If it keeps failing, include the error details in a bug report.",
		},
	});

	assert.deepEqual(await api.createContext({ contextId: "personal" }), {
		ok: false,
		error: {
			code: "unexpected_error",
			message: "create failed",
			recovery:
				"Retry the action. If it keeps failing, include the error details in a bug report.",
		},
	});
});

test("adapter normalizes create context onboarding failure responses", async () => {
	const api = createDevContextApi({
		async getLaunchState() {
			throw new Error("not used");
		},
		async launchProject() {
			throw new Error("not used");
		},
		async preflightLaunchProject() {
			throw new Error("not used");
		},
		async bindProject() {
			throw new Error("not used");
		},
		async unbindProject() {
			throw new Error("not used");
		},
		async createContext(request) {
			switch (request.contextId) {
				case "duplicate":
					return {
						code: "validation_error",
						message: "Unable to complete request.",
						recovery: "A context with this ID already exists.",
					};
				case "permission":
					return {
						code: "validation_error",
						message: "Unable to complete request.",
						recovery: "Check Dev Context storage permissions, then retry.",
					};
				default:
					return {
						code: "internal_error",
						message: "Dev Context failed unexpectedly.",
						recovery:
							"Retry the action. If it keeps failing, include debug details in a bug report.",
					};
			}
		},
	});

	assert.deepEqual(await api.createContext({ contextId: "duplicate" }), {
		ok: false,
		error: {
			code: "validation_error",
			message: "Unable to complete request.",
			recovery: "A context with this ID already exists.",
			contextMismatch: undefined,
		},
	});
	assert.deepEqual(await api.createContext({ contextId: "permission" }), {
		ok: false,
		error: {
			code: "validation_error",
			message: "Unable to complete request.",
			recovery: "Check Dev Context storage permissions, then retry.",
			contextMismatch: undefined,
		},
	});
	assert.deepEqual(await api.createContext({ contextId: "write-failure" }), {
		ok: false,
		error: {
			code: "internal_error",
			message: "Dev Context failed unexpectedly.",
			recovery:
				"Retry the action. If it keeps failing, include debug details in a bug report.",
			contextMismatch: undefined,
		},
	});
});

test("adapter normalizes structured diagnostics and preserves path disclosure metadata", async () => {
	const api = createDevContextApi({
		async getDiagnostics(request) {
			assert.deepEqual(request, { contextId: "personal" });
			return {
				groups: [
					{
						id: "context-storage",
						label: "Context storage",
						checks: [
							{
								id: "context-root",
								severity: "needs_attention",
								label: "Context directory",
								message: "The context directory needs repair.",
								details: [
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
			};
		},
	});

	assert.deepEqual(await api.getDiagnostics({ contextId: "personal" }), {
		ok: true,
		data: {
			groups: [
				{
					id: "context-storage",
					label: "Context storage",
					checks: [
						{
							id: "context-root",
							severity: "needs_attention",
							label: "Context directory",
							message: "The context directory needs repair.",
							details: [
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
	});
});

test("adapter normalizes repair actions and defaults destructive confirmation to false", async () => {
	const requests = [];
	const api = createDevContextApi({
		async getRepairActions(request) {
			assert.deepEqual(request, { contextId: "personal" });
			return {
				actions: [
					{
						id: "reset-provider:future",
						label: "Reset Future Provider storage",
						description: "Remove provider-owned files.",
						destructive: true,
						requiresConfirmation: true,
						targets: [
							{
								label: "Future Provider storage",
								path: "/contexts/personal/providers/future/auth.json",
								kind: "file",
							},
						],
					},
				],
			};
		},
		async runRepairAction(request) {
			requests.push(request);
			return { actionId: request.actionId, diagnostics: { groups: [] } };
		},
	});

	const actions = await api.getRepairActions({ contextId: "personal" });
	assert.equal(actions.ok, true);
	assert.deepEqual(
		actions.ok ? actions.data.actions[0].targets[0] : undefined,
		{
			label: "Future Provider storage",
			path: "/contexts/personal/providers/future/auth.json",
			kind: "file",
		},
	);
	assert.deepEqual(
		await api.runRepairAction({
			contextId: "personal",
			actionId: "reset-provider:future",
		}),
		{
			ok: true,
			data: { actionId: "reset-provider:future", diagnostics: { groups: [] } },
		},
	);
	assert.deepEqual(requests, [
		{
			contextId: "personal",
			actionId: "reset-provider:future",
			confirmDestructive: false,
		},
	]);
});
