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
import type {
	ProjectLaunchDecision,
	ProjectLaunchPending,
} from "../project-launch/project-launch-journey.js";
import {
	continueProjectLaunchJourney,
	runProjectLaunchJourney,
} from "../project-launch/project-launch-journey.js";

interface LaunchSelectorDependencies {
	projectPath: string;
	selectedContextId?: string;
	bindingContextId?: string;
	confirmContextMismatch?: boolean;
	onLaunchStarting?: (preflight: PreflightLaunchProjectResult) => void;
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
	| {
			runningEnvironmentConflict: RunningEnvironmentConflict;
			pending: ProjectLaunchPending;
	  }
	| { preflightReview: ProjectLaunchPending }
	| undefined;

interface LaunchSelectorContinuationDependencies {
	projectPath: string;
	bindingContextId?: string;
	onLaunchStarting?: (preflight: PreflightLaunchProjectResult) => void;
	bindProject: (
		request: BindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	pending: ProjectLaunchPending;
	decision: ProjectLaunchDecision;
}

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
	};

	const result = await runProjectLaunchJourney({
		request: launchRequest,
		prepareLaunch: prepareLaunchForSelection(dependencies, bindingContextId),
		onLaunchStarting: dependencies.onLaunchStarting,
		preflightLaunchProject: dependencies.preflightLaunchProject,
		launchProject: dependencies.launchProject,
	});
	return launchSelectorResult(result);
}

async function continueLaunchingSelectedContext(
	dependencies: LaunchSelectorContinuationDependencies,
): Promise<LaunchSelectorResult> {
	const result = await continueProjectLaunchJourney({
		pending: dependencies.pending,
		decision: dependencies.decision,
		prepareLaunch: prepareLaunchForSelection(
			dependencies,
			dependencies.bindingContextId,
		),
		onLaunchStarting: dependencies.onLaunchStarting,
		launchProject: dependencies.launchProject,
	});
	return launchSelectorResult(result);
}

function prepareLaunchForSelection(
	dependencies: Pick<
		LaunchSelectorDependencies,
		"bindingContextId" | "bindProject" | "projectPath"
	>,
	bindingContextId: string | undefined,
) {
	if (bindingContextId === undefined) {
		return undefined;
	}

	return async () => {
		const binding = await dependencies.bindProject({
			projectPath: dependencies.projectPath,
			contextId: bindingContextId,
		});
		return binding.ok
			? { ok: true as const }
			: { ok: false as const, error: binding.error };
	};
}

function launchSelectorResult(
	result: Awaited<ReturnType<typeof runProjectLaunchJourney>>,
): Exclude<LaunchSelectorResult, undefined> {
	if (result.kind === "launched") {
		return { ok: true, data: result.result };
	}
	if (result.kind === "failed") {
		return { ok: false, error: result.error };
	}
	if (result.kind === "preflight-review") {
		return { preflightReview: result.pending };
	}
	return {
		runningEnvironmentConflict: result.conflict,
		pending: result.pending,
	};
}

export type {
	LaunchRequestGuard,
	LaunchSelectorContinuationDependencies,
	LaunchSelectorDependencies,
	LaunchSelectorResult,
};
export {
	continueLaunchingSelectedContext,
	createLaunchRequestGuard,
	launchSelectedContext,
};
