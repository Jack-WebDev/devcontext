import type { ApiResult, CreateContextResult, DisplayError, LaunchState } from "../../lib/devctx-api";

interface CreateOnboardingContextDependencies {
  contextId: string;
  importProviderIds?: string[];
  createContext: (contextId: string, importProviderIds: string[]) => Promise<ApiResult<CreateContextResult>>;
  getLaunchState: () => Promise<ApiResult<LaunchState>>;
}

type CreateOnboardingContextResult =
  | { ok: true; created: CreateContextResult; launchState: LaunchState }
  | { ok: false; error: DisplayError };

async function createOnboardingContextAndRefresh(
  dependencies: CreateOnboardingContextDependencies,
): Promise<CreateOnboardingContextResult> {
  const created = await dependencies.createContext(dependencies.contextId, dependencies.importProviderIds ?? []);
  if (!created.ok) {
    return { ok: false, error: created.error };
  }

  const refreshed = await dependencies.getLaunchState();
  if (!refreshed.ok) {
    return { ok: false, error: refreshed.error };
  }

  return { ok: true, created: created.data, launchState: refreshed.data };
}

export { createOnboardingContextAndRefresh };
export type { CreateOnboardingContextDependencies, CreateOnboardingContextResult };
