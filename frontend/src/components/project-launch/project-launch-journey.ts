import type {
	ApiResult,
	DisplayError,
	LaunchProjectRequest,
	LaunchProjectResult,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	RunningEnvironmentConflict,
} from "../../lib/devctx-api";

interface ProjectLaunchJourneyDependencies {
	request: LaunchProjectRequest;
	preflightLaunchProject: (
		request: PreflightLaunchProjectRequest,
	) => Promise<ApiResult<PreflightLaunchProjectResult>>;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	prepareLaunch?: () => Promise<ProjectLaunchPreparationResult>;
	onLaunchStarting?: (preflight: PreflightLaunchProjectResult) => void;
}

type ProjectLaunchPreparationResult =
	| { ok: true }
	| { ok: false; error: DisplayError };

interface ProjectLaunchContinuationDependencies {
	pending: ProjectLaunchPending;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	decision: ProjectLaunchDecision;
	prepareLaunch?: () => Promise<ProjectLaunchPreparationResult>;
	onLaunchStarting?: (preflight: PreflightLaunchProjectResult) => void;
}

interface ProjectLaunchPending {
	request: LaunchProjectRequest;
	preflight: PreflightLaunchProjectResult;
}

type ProjectLaunchDecision = "continue-after-review" | "launch-another";

type ProjectLaunchJourneyResult =
	| { kind: "launched"; result: LaunchProjectResult }
	| { kind: "failed"; error: DisplayError }
	| { kind: "preflight-review"; pending: ProjectLaunchPending }
	| {
			kind: "running-environment-conflict";
			conflict: RunningEnvironmentConflict;
			pending: ProjectLaunchPending;
	  };

function requiresPreflightReview(
	preflight: PreflightLaunchProjectResult,
): boolean {
	return (preflight.groups ?? []).some((group) => group.status !== "ready");
}

async function runProjectLaunchJourney(
	dependencies: ProjectLaunchJourneyDependencies,
): Promise<ProjectLaunchJourneyResult> {
	const preflight = await dependencies.preflightLaunchProject(
		dependencies.request,
	);
	if (!preflight.ok) {
		return { kind: "failed", error: preflight.error };
	}

	const pending = { request: dependencies.request, preflight: preflight.data };
	if (requiresPreflightReview(preflight.data)) {
		return { kind: "preflight-review", pending };
	}

	return continueProjectLaunchJourney({
		pending,
		decision: "continue-after-review",
		launchProject: dependencies.launchProject,
		prepareLaunch: dependencies.prepareLaunch,
		onLaunchStarting: dependencies.onLaunchStarting,
	});
}

async function continueProjectLaunchJourney(
	dependencies: ProjectLaunchContinuationDependencies,
): Promise<Exclude<ProjectLaunchJourneyResult, { kind: "preflight-review" }>> {
	const { pending } = dependencies;
	const conflict = pending.preflight.runningEnvironmentConflict;
	if (conflict !== undefined && dependencies.decision !== "launch-another") {
		return { kind: "running-environment-conflict", conflict, pending };
	}

	dependencies.onLaunchStarting?.(pending.preflight);
	const prepared = await dependencies.prepareLaunch?.();
	if (prepared !== undefined && !prepared.ok) {
		return { kind: "failed", error: prepared.error };
	}
	const launch = await dependencies.launchProject({
		...pending.request,
		...(requiresPreflightReview(pending.preflight)
			? { confirmPreflightWarnings: true }
			: {}),
	});
	if (!launch.ok) {
		return { kind: "failed", error: launch.error };
	}
	return { kind: "launched", result: launch.data };
}

export type {
	ProjectLaunchContinuationDependencies,
	ProjectLaunchDecision,
	ProjectLaunchJourneyDependencies,
	ProjectLaunchJourneyResult,
	ProjectLaunchPending,
	ProjectLaunchPreparationResult,
};
export {
	continueProjectLaunchJourney,
	requiresPreflightReview,
	runProjectLaunchJourney,
};
