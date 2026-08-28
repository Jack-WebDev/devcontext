import type { ContextState } from "../../lib/devctx-api";

// hasAccountIdentityMismatch returns true only for the backend-owned,
// meaningful identity warning introduced by Phase 168.
export function hasAccountIdentityMismatch(context: ContextState | undefined): boolean {
  return context?.confidence?.checks.some(
    (check) => check.component === "identity" && check.severity === "needs_attention",
  ) ?? false;
}
