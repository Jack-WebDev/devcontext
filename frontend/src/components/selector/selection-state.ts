import type { ContextState, LaunchState } from "../../lib/devctx-api";

type ContextNavigationDirection = "previous" | "next";

function initialSelectedContextId(launchState: LaunchState): string | undefined {
  const boundContextId = launchState.binding.contextId;
  if (!launchState.binding.bound || launchState.binding.dangling || boundContextId === undefined) {
    return undefined;
  }

  return contextExists(launchState.contexts, boundContextId) ? boundContextId : undefined;
}

function initialRovingContextId(launchState: LaunchState): string | undefined {
  return initialSelectedContextId(launchState) ?? launchState.contexts[0]?.id;
}

function nextSelectedContextId(contexts: ContextState[], contextId: string): string | undefined {
  return contextExists(contexts, contextId) ? contextId : undefined;
}

function nextKeyboardContextId(
  contexts: ContextState[],
  currentContextId: string | undefined,
  direction: ContextNavigationDirection,
): string | undefined {
  if (contexts.length === 0) {
    return undefined;
  }

  const currentIndex = contexts.findIndex((context) => context.id === currentContextId);
  if (currentIndex === -1) {
    return contexts[0]?.id;
  }

  const nextIndex =
    direction === "next"
      ? Math.min(currentIndex + 1, contexts.length - 1)
      : Math.max(currentIndex - 1, 0);

  return contexts[nextIndex]?.id;
}

function contextExists(contexts: ContextState[], contextId: string): boolean {
  return contexts.some((context) => context.id === contextId);
}

export {
  initialRovingContextId,
  initialSelectedContextId,
  nextKeyboardContextId,
  nextSelectedContextId,
};
export type { ContextNavigationDirection };
