import type { ContextState } from "../../lib/devctx-api";

const defaultContextIds = ["personal", "company"] as const;

type DefaultContextId = (typeof defaultContextIds)[number];

function missingDefaultContextIds(contexts: ContextState[]): DefaultContextId[] {
  const contextIds = new Set(contexts.map((context) => context.id));
  return defaultContextIds.filter((contextId) => !contextIds.has(contextId));
}

export { missingDefaultContextIds };
export type { DefaultContextId };
