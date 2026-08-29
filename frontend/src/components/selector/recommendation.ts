import type { ContextState, LaunchState } from "../../lib/devctx-api";

type ContextRecommendationCategory = "remembered" | "verified" | "conflict";

interface ContextRecommendation {
	category: ContextRecommendationCategory;
	label: "Remembered" | "Verified" | "Conflict";
	detail: string;
}

function contextRecommendation(
	launchState: LaunchState,
	context: ContextState,
): ContextRecommendation | undefined {
	if (hasContextConflict(launchState, context.id)) {
		return {
			category: "conflict",
			label: "Conflict",
			detail: "This context conflicts with the project's remembered context.",
		};
	}

	if (
		launchState.binding.bound &&
		launchState.binding.contextId === context.id &&
		launchState.resolutionSource === "project_binding"
	) {
		return {
			category: "remembered",
			label: "Remembered",
			detail: "Remembered for this project.",
		};
	}

	if (context.confidence?.status === "ready") {
		return {
			category: "verified",
			label: "Verified",
			detail: "Dev Context verified the required launch checks.",
		};
	}

	return undefined;
}

function hasContextConflict(
	launchState: LaunchState,
	contextId: string,
): boolean {
	return launchState.warnings.some(
		(warning) =>
			warning.code === "context_mismatch" &&
			(warning.boundContextId === contextId ||
				warning.requestedContextId === contextId),
	);
}

export type { ContextRecommendation, ContextRecommendationCategory };
export { contextRecommendation };
