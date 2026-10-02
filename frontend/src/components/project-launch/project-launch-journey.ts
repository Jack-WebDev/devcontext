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
	allowExistingEnvironmentLaunch?: boolean;
	onPreflightComplete?: (
		preflight: PreflightLaunchProjectResult,
	) => boolean | undefined | Promise<boolean | undefined>;
	prepareLaunch?: () => Promise<ProjectLaunchPreparationResult>;
}

type ProjectLaunchPreparationResult =
	| { ok: true }
	| { ok: false; error: DisplayError };

interface ProjectLaunchContinuationDependencies {
	request: LaunchProjectRequest;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	preflight?: PreflightLaunchProjectResult;
	allowExistingEnvironmentLaunch?: boolean;
}

type ProjectLaunchJourneyResult =
	| { kind: "launched"; result: LaunchProjectResult }
	| { kind: "failed"; error: DisplayError }
	| { kind: "preflight-review"; preflight: PreflightLaunchProjectResult }
	| {
			kind: "running-environment-conflict";
			conflict: RunningEnvironmentConflict;
	  };

function requiresPreflightReview(
	preflight: PreflightLaunchProjectResult,
): boolean {
	return preflight.groups.some((group) => group.status !== "ready");
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

	const shouldLaunch = await dependencies.onPreflightComplete?.(preflight.data);
	if (shouldLaunch === false) {
		return { kind: "preflight-review", preflight: preflight.data };
	}

	return continueProjectLaunchJourney({
		request: dependencies.request,
		preflight: preflight.data,
		allowExistingEnvironmentLaunch: dependencies.allowExistingEnvironmentLaunch,
		launchProject: async (request) => {
			const prepared = await dependencies.prepareLaunch?.();
			if (prepared !== undefined && !prepared.ok) {
				return prepared;
			}
			return dependencies.launchProject(request);
		},
	});
}

async function continueProjectLaunchJourney(
	dependencies: ProjectLaunchContinuationDependencies,
): Promise<Exclude<ProjectLaunchJourneyResult, { kind: "preflight-review" }>> {
	const conflict = dependencies.preflight?.runningEnvironmentConflict;
	if (conflict !== undefined && !dependencies.allowExistingEnvironmentLaunch) {
		return { kind: "running-environment-conflict", conflict };
	}

	const launch = await dependencies.launchProject(dependencies.request);
	if (!launch.ok) {
		return { kind: "failed", error: launch.error };
	}
	return { kind: "launched", result: launch.data };
}

export type {
	ProjectLaunchContinuationDependencies,
	ProjectLaunchJourneyDependencies,
	ProjectLaunchJourneyResult,
	ProjectLaunchPreparationResult,
};
export {
	continueProjectLaunchJourney,
	requiresPreflightReview,
	runProjectLaunchJourney,
};
