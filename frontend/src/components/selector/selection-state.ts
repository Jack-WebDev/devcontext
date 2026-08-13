import type { ContextState, LaunchState } from "../../lib/devctx-api";

function initialSelectedContextId(launchState: LaunchState): string | undefined {
  const boundContextId = launchState.binding.contextId;
  if (!launchState.binding.bound || launchState.binding.dangling || boundContextId === undefined) {
    return undefined;
  }

  return contextExists(launchState.contexts, boundContextId) ? boundContextId : undefined;
}

function nextSelectedContextId(contexts: ContextState[], contextId: string): string | undefined {
  return contextExists(contexts, contextId) ? contextId : undefined;
}

function contextExists(contexts: ContextState[], contextId: string): boolean {
  return contexts.some((context) => context.id === contextId);
}

export { initialSelectedContextId, nextSelectedContextId };
