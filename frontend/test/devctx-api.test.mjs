import assert from "node:assert/strict";
import test from "node:test";

import {createDevContextApi} from "../.tmp-test/src/lib/devctx-api.js";

const toolFixture = () => ({
  id: "second-tool",
  name: "Second Tool",
  status: "ready",
  message: "Second Tool is available for launch.",
});

const toolOptionFixture = () => ({id: "second-tool", name: "Second Tool"});

test("adapter normalizes successful Wails calls", async () => {
  const calls = [];
  const api = createDevContextApi({
    async getLaunchState(request) {
      calls.push(["getLaunchState", request]);
      return {
        project: {name: "api", path: "/work/api"},
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
                identity: {
                  status: "verified",
                  fields: [{label: "Email", value: "user@company.com"}, {label: "Plan", value: "Business"}, {label: "Account", value: "acct_123"}],
                },
              },
              {
                id: "claude",
                name: "Claude",
                enabled: true,
                state: "ready",
                identity: {
                  status: "verified",
                  fields: [{label: "Subscription", value: "Pro"}, {label: "Organization UUID", value: "e783"}, {label: "Organization", value: "Jishin Labs"}],
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
            metadata: {accent: "blue"},
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
            fields: [{label: "Email", value: "user@company.com"}, {label: "Plan", value: "Business"}, {label: "Account", value: "acct_123"}],
          },
          {
            providerId: "claude",
            name: "Claude",
            metadataAvailable: true,
            fields: [{label: "Subscription", value: "Pro"}, {label: "Organization UUID", value: "e783"}, {label: "Organization", value: "Jishin Labs"}],
          },
        ],
      };
    },
    async launchProject(request) {
      calls.push(["launchProject", request]);
      return {
        project: {name: "api", path: "/work/api"},
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
        project: {name: "api", path: "/work/api"},
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

  assert.deepEqual(await api.getLaunchState({projectPath: "/work/api"}), {
    ok: true,
    data: {
      project: {name: "api", path: "/work/api"},
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
              identity: {
                status: "verified",
                message: undefined,
                fields: [{label: "Email", value: "user@company.com"}, {label: "Plan", value: "Business"}, {label: "Account", value: "acct_123"}],
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
                fields: [{label: "Subscription", value: "Pro"}, {label: "Organization UUID", value: "e783"}, {label: "Organization", value: "Jishin Labs"}],
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
          metadata: {accent: "blue"},
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
          fields: [{label: "Email", value: "user@company.com"}, {label: "Plan", value: "Business"}, {label: "Account", value: "acct_123"}],
        },
        {
          providerId: "claude",
          name: "Claude",
          metadataAvailable: true,
          fields: [{label: "Subscription", value: "Pro"}, {label: "Organization UUID", value: "e783"}, {label: "Organization", value: "Jishin Labs"}],
        },
      ],
    },
  });

  assert.deepEqual(await api.preflightLaunchProject({projectPath: "/work/api", contextId: "personal"}), {
    ok: true,
    data: {
      project: {name: "api", path: "/work/api"},
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
  });
  assert.deepEqual(await api.launchProject({projectPath: "/work/api", contextId: "personal"}), {
    ok: true,
    data: {
      project: {name: "api", path: "/work/api"},
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
  });
  assert.deepEqual(await api.bindProject({projectPath: "/work/api", contextId: "personal"}), {
    ok: true,
    data: {
      projectPath: "/work/api",
      bound: true,
      contextId: "personal",
      dangling: false,
      missingContextId: undefined,
      recovery: undefined,
    },
  });
  assert.deepEqual(await api.unbindProject({projectPath: "/work/api"}), {
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
  assert.deepEqual(await api.createContext({contextId: "personal", importProviderIds: ["codex"]}), {
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
  });
  assert.deepEqual(calls, [
    ["getLaunchState", {projectPath: "/work/api"}],
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
    ["bindProject", {projectPath: "/work/api", contextId: "personal"}],
    ["unbindProject", {projectPath: "/work/api"}],
    ["createContext", {contextId: "personal", importProviderIds: ["codex"]}],
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

  assert.deepEqual(await api.launchProject({projectPath: "/work/api", contextId: "personal"}), {
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
  });

  assert.deepEqual(await api.bindProject({projectPath: "/work/api", contextId: "personal"}), {
    ok: false,
    error: {
      code: "unexpected_error",
      message: "Future failure.",
      recovery: "Retry later.",
      contextMismatch: undefined,
    },
  });

  assert.deepEqual(await api.createContext({contextId: "personal"}), {
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
          project: {name: "api", path: "/work/api"},
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
      project: {name: "api", path: "/work/api"},
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

  assert.deepEqual(await api.launchProject({contextId: "personal"}), {
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
      recovery: "Retry the action. If it keeps failing, include the error details in a bug report.",
    },
  });

  assert.deepEqual(await api.launchProject({contextId: "personal"}), {
    ok: false,
    error: {
      code: "launch_error",
      message: "Unable to launch editor.",
      recovery: "Check the editor command.",
      contextMismatch: undefined,
    },
  });

  assert.deepEqual(await api.bindProject({contextId: "personal"}), {
    ok: false,
    error: {
      code: "unexpected_error",
      message: "request failed",
      recovery: "Retry the action. If it keeps failing, include the error details in a bug report.",
    },
  });

  assert.deepEqual(await api.unbindProject(), {
    ok: false,
    error: {
      code: "unexpected_error",
      message: "Dev Context could not complete the request.",
      recovery: "Retry the action. If it keeps failing, include the error details in a bug report.",
    },
  });

  assert.deepEqual(await api.createContext({contextId: "personal"}), {
    ok: false,
    error: {
      code: "unexpected_error",
      message: "create failed",
      recovery: "Retry the action. If it keeps failing, include the error details in a bug report.",
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
            recovery: "Retry the action. If it keeps failing, include debug details in a bug report.",
          };
      }
    },
  });

  assert.deepEqual(await api.createContext({contextId: "duplicate"}), {
    ok: false,
    error: {
      code: "validation_error",
      message: "Unable to complete request.",
      recovery: "A context with this ID already exists.",
      contextMismatch: undefined,
    },
  });
  assert.deepEqual(await api.createContext({contextId: "permission"}), {
    ok: false,
    error: {
      code: "validation_error",
      message: "Unable to complete request.",
      recovery: "Check Dev Context storage permissions, then retry.",
      contextMismatch: undefined,
    },
  });
  assert.deepEqual(await api.createContext({contextId: "write-failure"}), {
    ok: false,
    error: {
      code: "internal_error",
      message: "Dev Context failed unexpectedly.",
      recovery: "Retry the action. If it keeps failing, include debug details in a bug report.",
      contextMismatch: undefined,
    },
  });
});
