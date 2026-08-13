import assert from "node:assert/strict";
import test from "node:test";

import {renderToStaticMarkup} from "react-dom/server";

import {ContextCard} from "../.tmp-test/src/components/selector/ContextCard.js";
import {ProjectIdentity} from "../.tmp-test/src/components/selector/ProjectIdentity.js";
import {
  initialSelectedContextId,
  nextSelectedContextId,
} from "../.tmp-test/src/components/selector/selection-state.js";

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
});

test("context card renders as a selectable control when wired", () => {
  const context = contextFixture("client-a", "Client A");
  const html = renderToStaticMarkup(ContextCard({context, onSelect: () => {}}));

  assert.match(html, /<button/);
  assert.match(html, /aria-pressed="false"/);
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

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
