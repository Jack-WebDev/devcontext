import type {
	ApiResult,
	BindProjectRequest,
	LaunchProjectRequest,
	LaunchProjectResult,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	ProjectBindingState,
} from "../../lib/devctx-api";
import type {
	ProjectLaunchDecision,
	ProjectLaunchJourneyResult,
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
): Promise<ProjectLaunchJourneyResult | undefined> {
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

	return runProjectLaunchJourney({
		request: launchRequest,
		prepareLaunch: prepareLaunchForSelection(dependencies, bindingContextId),
		onLaunchStarting: dependencies.onLaunchStarting,
		preflightLaunchProject: dependencies.preflightLaunchProject,
		launchProject: dependencies.launchProject,
	});
}

async function continueLaunchingSelectedContext(
	dependencies: LaunchSelectorContinuationDependencies,
): Promise<Exclude<ProjectLaunchJourneyResult, { kind: "preflight-review" }>> {
	return continueProjectLaunchJourney({
		pending: dependencies.pending,
		decision: dependencies.decision,
		prepareLaunch: prepareLaunchForSelection(
			dependencies,
			dependencies.bindingContextId,
		),
		onLaunchStarting: dependencies.onLaunchStarting,
		launchProject: dependencies.launchProject,
	});
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

export type {
	LaunchRequestGuard,
	LaunchSelectorContinuationDependencies,
	LaunchSelectorDependencies,
};
export {
	continueLaunchingSelectedContext,
	createLaunchRequestGuard,
	launchSelectedContext,
};
