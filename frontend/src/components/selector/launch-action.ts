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
  bindProject: (request: BindProjectRequest) => Promise<ApiResult<ProjectBindingState>>;
  launchProject: (request: LaunchProjectRequest) => Promise<ApiResult<LaunchProjectResult>>;
}

async function launchSelectedContext(dependencies: LaunchSelectorDependencies): Promise<ApiResult<LaunchProjectResult> | undefined> {
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
  });
}

export { launchSelectedContext };
export type { LaunchSelectorDependencies };
