import assert from "node:assert/strict";
import test from "node:test";

import {renderToStaticMarkup} from "react-dom/server";

import {ContextCard} from "../.tmp-test/src/components/selector/ContextCard.js";
import {ProjectIdentity} from "../.tmp-test/src/components/selector/ProjectIdentity.js";

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

function contextFixture(id, name) {
  return {
    id,
    name,
    editor: {type: "vscode"},
    providers: [],
  };
}

function escapeRegExp(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
