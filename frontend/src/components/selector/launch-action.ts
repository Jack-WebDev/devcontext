import type {
	ApiResult,
	BindProjectRequest,
	LaunchProjectRequest,
	LaunchProjectResult,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	ProjectBindingState,
	RunningEnvironmentConflict,
} from "../../lib/devctx-api";

interface LaunchSelectorDependencies {
	projectPath: string;
	selectedContextId?: string;
	bindingContextId?: string;
	confirmContextMismatch?: boolean;
	allowExistingEnvironmentLaunch?: boolean;
	onPreflightComplete?: (result: PreflightLaunchProjectResult) => void;
	bindProject: (
		request: BindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	preflightLaunchProject: (
		request: PreflightLaunchProjectRequest,
	) => Promise<ApiResult<PreflightLaunchProjectResult>>;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
}

type LaunchSelectorResult =
	| ApiResult<LaunchProjectResult>
	| ApiResult<PreflightLaunchProjectResult>
	| ApiResult<ProjectBindingState>
	| { runningEnvironmentConflict: RunningEnvironmentConflict }
	| undefined;

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

async function launchSelectedContext(
	dependencies: LaunchSelectorDependencies,
): Promise<LaunchSelectorResult> {
	const contextId = dependencies.selectedContextId;
	if (contextId === undefined) {
		return undefined;
	}

	if (dependencies.bindingContextId !== undefined) {
		if (dependencies.bindingContextId !== contextId) {
			throw new Error("Binding context must match the selected launch context.");
		}
		const binding = await dependencies.bindProject({
			projectPath: dependencies.projectPath,
			contextId: dependencies.bindingContextId,
		});
		if (!binding.ok) {
			return binding;
		}
	}

	const launchRequest = {
		projectPath: dependencies.projectPath,
		contextId,
		...(dependencies.confirmContextMismatch
			? { confirmContextMismatch: true }
			: {}),
	};

	const preflight = await dependencies.preflightLaunchProject(launchRequest);
	if (!preflight.ok) {
		return preflight;
	}

	if (
		preflight.data.runningEnvironmentConflict &&
		!dependencies.allowExistingEnvironmentLaunch
	) {
		return {
			runningEnvironmentConflict: preflight.data.runningEnvironmentConflict,
		};
	}

	dependencies.onPreflightComplete?.(preflight.data);

	return dependencies.launchProject({
		...launchRequest,
	});
}

export type {
	LaunchRequestGuard,
	LaunchSelectorDependencies,
	LaunchSelectorResult,
};
export { createLaunchRequestGuard, launchSelectedContext };
