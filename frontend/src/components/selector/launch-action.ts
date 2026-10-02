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
import { runProjectLaunchJourney } from "../project-launch/project-launch-journey.js";

interface LaunchSelectorDependencies {
	projectPath: string;
	selectedContextId?: string;
	bindingContextId?: string;
	confirmContextMismatch?: boolean;
	allowExistingEnvironmentLaunch?: boolean;
	confirmPreflightWarnings?: boolean;
	onPreflightComplete?: (
		result: PreflightLaunchProjectResult,
	) => boolean | undefined | Promise<boolean | undefined>;
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
	| { preflightReview: PreflightLaunchProjectResult }
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
	const bindingContextId = dependencies.bindingContextId;
	if (contextId === undefined) {
		return undefined;
	}
	if (bindingContextId !== undefined && bindingContextId !== contextId) {
		throw new Error("Binding context must match the selected launch context.");
	}

	const launchRequest = {
		projectPath: dependencies.projectPath,
		contextId,
		...(dependencies.confirmContextMismatch
			? { confirmContextMismatch: true }
			: {}),
		...(dependencies.confirmPreflightWarnings
			? { confirmPreflightWarnings: true }
			: {}),
	};

	const result = await runProjectLaunchJourney({
		request: launchRequest,
		allowExistingEnvironmentLaunch: dependencies.allowExistingEnvironmentLaunch,
		onPreflightComplete: dependencies.onPreflightComplete,
		prepareLaunch:
			bindingContextId === undefined
				? undefined
				: async () => {
						const binding = await dependencies.bindProject({
							projectPath: dependencies.projectPath,
							contextId: bindingContextId,
						});
						return binding.ok
							? { ok: true }
							: { ok: false, error: binding.error };
					},
		preflightLaunchProject: dependencies.preflightLaunchProject,
		launchProject: dependencies.launchProject,
	});

	if (result.kind === "launched") {
		return { ok: true, data: result.result };
	}
	if (result.kind === "failed") {
		return { ok: false, error: result.error };
	}
	if (result.kind === "preflight-review") {
		return { preflightReview: result.preflight };
	}
	return { runningEnvironmentConflict: result.conflict };
}

export type {
	LaunchRequestGuard,
	LaunchSelectorDependencies,
	LaunchSelectorResult,
};
export { createLaunchRequestGuard, launchSelectedContext };
