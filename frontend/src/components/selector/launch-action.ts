import type {
  ApiResult,
  BindProjectRequest,
  LaunchProjectRequest,
  LaunchProjectResult,
  ProjectBindingState,
} from "../../lib/devctx-api";

interface LaunchSelectorDependencies {
  projectPath: string;
  selectedContextId?: string;
  rememberProject: boolean;
  confirmContextMismatch?: boolean;
  bindProject: (request: BindProjectRequest) => Promise<ApiResult<ProjectBindingState>>;
  launchProject: (request: LaunchProjectRequest) => Promise<ApiResult<LaunchProjectResult>>;
}

type LaunchSelectorResult = ApiResult<LaunchProjectResult> | ApiResult<ProjectBindingState> | undefined;

interface LaunchRequestGuard {
  run<T>(operation: () => Promise<T>): Promise<T | undefined>;
}

function createLaunchRequestGuard(): LaunchRequestGuard {
  let inFlight = false;

  return {
    async run(operation) {
      if (inFlight) {
        return undefined;
      }

      inFlight = true;
      try {
        return await operation();
      } finally {
        inFlight = false;
      }
    },
  };
}

async function launchSelectedContext(dependencies: LaunchSelectorDependencies): Promise<LaunchSelectorResult> {
  const contextId = dependencies.selectedContextId;
  if (contextId === undefined) {
    return undefined;
  }

  if (dependencies.rememberProject) {
    const binding = await dependencies.bindProject({
      projectPath: dependencies.projectPath,
      contextId,
    });
    if (!binding.ok) {
      return binding;
    }
  }

  return dependencies.launchProject({
    projectPath: dependencies.projectPath,
    contextId,
    ...(dependencies.confirmContextMismatch ? { confirmContextMismatch: true } : {}),
  });
}

export { createLaunchRequestGuard, launchSelectedContext };
export type { LaunchRequestGuard, LaunchSelectorDependencies, LaunchSelectorResult };
