import assert from "node:assert/strict";
import test from "node:test";

import {renderToStaticMarkup} from "react-dom/server";

import {ContextMismatchDialog} from "../.tmp-test/src/components/selector/ContextMismatchDialog.js";
import {ContextCard} from "../.tmp-test/src/components/selector/ContextCard.js";
import {GuiErrorNotice} from "../.tmp-test/src/components/selector/GuiErrorNotice.js";
import {ProjectIdentity} from "../.tmp-test/src/components/selector/ProjectIdentity.js";
import {
  boundContextName,
  RememberProjectControl,
} from "../.tmp-test/src/components/selector/RememberProjectControl.js";
import {SelectorActions} from "../.tmp-test/src/components/selector/SelectorActions.js";
import {cancelSelector} from "../.tmp-test/src/components/selector/cancel-action.js";
import {
  createLaunchRequestGuard,
  launchSelectedContext,
} from "../.tmp-test/src/components/selector/launch-action.js";
import {
  initialRovingContextId,
  initialSelectedContextId,
  nextKeyboardContextId,
  nextSelectedContextId,
} from "../.tmp-test/src/components/selector/selection-state.js";
import {
  canLaunchSelectedContextFromKeyboard,
  escapeKeyboardAction,
} from "../.tmp-test/src/components/selector/selector-keyboard.js";
import {createDevContextWindow} from "../.tmp-test/src/lib/devctx-window.js";

test("project identity preserves full project names and paths", () => {
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
    const html = renderToStaticMarkup(ProjectIdentity({project}));

    assert.match(html, /class="[^"]*truncate[^"]*"/);
    assert.match(html, new RegExp(`title="${escapeRegExp(project.name)}"`));
    assert.match(html, new RegExp(`title="${escapeRegExp(project.path)}"`));
    assert.ok(html.includes(project.name));
    assert.ok(html.includes(project.path));
  }
});

test("context card renders generic context names and ids", () => {
  const contexts = [
    contextFixture("personal", "Personal"),
    contextFixture("company", "Company"),
    contextFixture("client-a", "Client A"),
  ];

  for (const context of contexts) {
    const html = renderToStaticMarkup(ContextCard({context}));

    assert.match(html, /<article/);
    assert.ok(html.includes(context.name));
    assert.ok(html.includes(context.id));
    assert.match(html, new RegExp(`title="${escapeRegExp(context.name)}"`));
    assert.match(html, new RegExp(`title="${escapeRegExp(context.id)}"`));
  }
});

test("context card can represent a selected context", () => {
  const context = contextFixture("personal", "Personal");
  const html = renderToStaticMarkup(ContextCard({context, selected: true}));

  assert.match(html, /data-selected="true"/);
  assert.match(html, /border-primary/);
  assert.match(html, />Selected</);
});

test("context card renders as a selectable control when wired", () => {
  const context = contextFixture("client-a", "Client A");
  const html = renderToStaticMarkup(ContextCard({context, tabIndex: -1, onSelect: () => {}}));

  assert.match(html, /<button/);
  assert.match(html, /aria-pressed="false"/);
  assert.match(html, /tabindex="-1"/);
  assert.match(html, />Not selected</);
});

test("context card renders enabled provider status variants with accessible names", () => {
  const context = contextFixture("personal", "Personal", [
    providerFixture("claude-ready", "Claude", true, "ready"),
    providerFixture("codex-not-configured", "Codex", true, "not_configured", "Codex context directory is empty"),
    providerFixture("claude-missing", "Claude", true, "directory_missing", "Claude context directory is missing"),
    providerFixture("codex-unavailable", "Codex", true, "unavailable", "Codex command was not found"),
    providerFixture("disabled", "Disabled Provider", false, "ready"),
  ]);
  const html = renderToStaticMarkup(ContextCard({context}));

  assert.match(html, /Claude local status: Ready/);
  assert.match(html, /Codex local status: Not configured/);
  assert.match(html, /Claude local status: Directory missing/);
  assert.match(html, /Codex local status: Unavailable/);
  assert.ok(html.includes("Codex context directory is empty"));
  assert.ok(!html.includes("Disabled Provider"));
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

  assert.equal(nextKeyboardContextId(contexts, "personal", "previous"), "personal");
  assert.equal(nextKeyboardContextId(contexts, "personal", "next"), "company");
  assert.equal(nextKeyboardContextId(contexts, "company", "previous"), "personal");
  assert.equal(nextKeyboardContextId(contexts, "company", "next"), "client-a");
  assert.equal(nextKeyboardContextId(contexts, "client-a", "previous"), "company");
  assert.equal(nextKeyboardContextId(contexts, "client-a", "next"), "client-a");
});

test("keyboard context navigation handles empty, single, and missing active lists", () => {
  const singleContext = [contextFixture("personal", "Personal")];
  const multipleContexts = [
    contextFixture("personal", "Personal"),
    contextFixture("company", "Company"),
  ];

  assert.equal(nextKeyboardContextId([], undefined, "next"), undefined);
  assert.equal(nextKeyboardContextId(singleContext, "personal", "next"), "personal");
  assert.equal(nextKeyboardContextId(singleContext, "personal", "previous"), "personal");
  assert.equal(nextKeyboardContextId(multipleContexts, undefined, "next"), "personal");
  assert.equal(nextKeyboardContextId(multipleContexts, "missing", "previous"), "personal");
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
  assert.doesNotMatch(html, /checked=""/);
  assert.doesNotMatch(html, /disabled=""/);
  assert.ok(html.includes("Remember this project"));
  assert.ok(html.includes("Use this context automatically for this project next time."));
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
  assert.ok(html.includes("This project is remembered for"));
  assert.ok(html.includes("Company"));
  assert.equal(boundContextName(state.binding, state.contexts), "Company");
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

test("launch action does nothing without a selected context", async () => {
  const calls = [];
  const result = await launchSelectedContext({
    projectPath: "/work/api",
    rememberProject: false,
    bindProject(request) {
      calls.push(["bindProject", request]);
      return Promise.resolve(projectBindingResult());
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
    rememberProject: false,
    bindProject(request) {
      calls.push(["bindProject", request]);
      return Promise.resolve(projectBindingResult());
    },
    launchProject(request) {
      calls.push(["launchProject", request]);
      return Promise.resolve(launchProjectResult());
    },
  });

  assert.deepEqual(result, launchProjectResult());
  assert.deepEqual(calls, [
    ["launchProject", {projectPath: "/work/api", contextId: "personal"}],
  ]);
});

test("launch action binds before launch when remember is on", async () => {
  const calls = [];
  const result = await launchSelectedContext({
    projectPath: "/work/api",
    selectedContextId: "company",
    rememberProject: true,
    bindProject(request) {
      calls.push(["bindProject", request]);
      return Promise.resolve(projectBindingResult());
    },
    launchProject(request) {
      calls.push(["launchProject", request]);
      return Promise.resolve(launchProjectResult());
    },
  });

  assert.deepEqual(result, launchProjectResult());
  assert.deepEqual(calls, [
    ["bindProject", {projectPath: "/work/api", contextId: "company"}],
    ["launchProject", {projectPath: "/work/api", contextId: "company"}],
  ]);
});

test("launch action returns binding errors without launching", async () => {
  const calls = [];
  const result = await launchSelectedContext({
    projectPath: "/work/api",
    selectedContextId: "company",
    rememberProject: true,
    bindProject(request) {
      calls.push(["bindProject", request]);
      return Promise.resolve(apiError("validation_error", "Unable to complete request.", "Check the selected project and context, then retry."));
    },
    launchProject(request) {
      calls.push(["launchProject", request]);
      return Promise.resolve(launchProjectResult());
    },
  });

  assert.deepEqual(result, apiError("validation_error", "Unable to complete request.", "Check the selected project and context, then retry."));
  assert.deepEqual(calls, [
    ["bindProject", {projectPath: "/work/api", contextId: "company"}],
  ]);
});

test("launch action resubmits explicit context mismatch confirmation", async () => {
  const calls = [];
  const result = await launchSelectedContext({
    projectPath: "/work/api",
    selectedContextId: "personal",
    rememberProject: false,
    confirmContextMismatch: true,
    bindProject(request) {
      calls.push(["bindProject", request]);
      return Promise.resolve(projectBindingResult());
    },
    launchProject(request) {
      calls.push(["launchProject", request]);
      return Promise.resolve(launchProjectResult());
    },
  });

  assert.deepEqual(result, launchProjectResult());
  assert.deepEqual(calls, [
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

test("context mismatch dialog shows risk details and deliberate actions", () => {
  const html = renderToStaticMarkup(
    ContextMismatchDialog({
      mismatch: {
        projectPath: "/work/api",
        boundContextId: "company",
        requestedContextId: "personal",
      },
      contexts: [contextFixture("company", "Company"), contextFixture("personal", "Personal")],
      launchPending: false,
      onCancel: () => {},
      onOpenAnyway: () => {},
    }),
  );

  assert.match(html, /role="dialog"/);
  assert.ok(html.includes("Context mismatch"));
  assert.ok(html.includes("/work/api"));
  assert.ok(html.includes("Company"));
  assert.ok(html.includes("Personal"));
  assert.ok(html.includes("expose project files"));
  assert.ok(html.includes("Cancel"));
  assert.ok(html.includes("Open Anyway"));
  assert.match(html, /focus-visible:ring-2/);
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
      onOpenAnyway: () => {},
    }),
  );

  assert.ok(html.includes("Opening..."));
  assert.match(html, /disabled=""/);
});

test("gui error notice renders failure and recovery guidance", () => {
  const errors = [
    apiError("validation_error", "Unable to complete request.", "Check the selected project and context, then retry."),
    apiError("launch_error", "Unable to launch editor.", "Check the editor command, project path, and permissions, then retry."),
    apiError("internal_error", "Dev Context failed unexpectedly.", "Retry the action. If it keeps failing, include debug details in a bug report."),
    apiError("unexpected_error", "Wails bridge unavailable", "Retry the action. If it keeps failing, include the error details in a bug report."),
  ];

  for (const error of errors) {
    const html = renderToStaticMarkup(GuiErrorNotice({error: error.error}));

    assert.match(html, /role="alert"/);
    assert.ok(html.includes(error.error.message));
    assert.ok(html.includes(error.error.recovery));
  }
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
});

function contextFixture(id, name, providers = []) {
  return {
    id,
    name,
    editor: {type: "vscode"},
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
  };
}

function launchStateFixture(overrides = {}) {
  return {
    project: {name: "api", path: "/work/api"},
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
    ...overrides,
  };
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
      project: {name: "api", path: "/work/api"},
      context: contextFixture("personal", "Personal"),
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

  return {promise, resolve, reject};
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
