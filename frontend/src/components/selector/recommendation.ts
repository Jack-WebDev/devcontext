import type { ContextState, LaunchState } from "../../lib/devctx-api";

type ContextRecommendationCategory = "remembered" | "verified" | "conflict";

interface ContextRecommendation {
	category: ContextRecommendationCategory;
	label: "Remembered" | "Verified" | "Conflict";
	detail: string;
	reasons: string[];
}

function contextRecommendation(
	launchState: LaunchState,
	context: ContextState,
): ContextRecommendation | undefined {
	const conflictWarning = contextConflictWarning(launchState, context.id);
	if (conflictWarning) {
		return {
			category: "conflict",
			label: "Conflict",
			detail: "This context conflicts with the project's remembered context.",
			reasons: [conflictWarning.message],
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
			reasons: [`${launchState.project.name} is bound to ${context.name}.`],
		};
	}

	if (context.confidence?.status === "ready") {
		return {
			category: "verified",
			label: "Verified",
			detail: "Dev Context verified the required launch checks.",
			reasons: verificationReasons(context),
		};
	}

	return undefined;
}

function verificationReasons(context: ContextState): string[] {
	const readyChecks = context.confidence?.checks.filter(
		(check) => check.severity === "ready",
	);
	if (readyChecks === undefined || readyChecks.length === 0) {
		return ["Required launch checks are ready."];
	}
	return readyChecks.map((check) => `${check.label}: ${check.message}`);
}

function contextConflictWarning(
	launchState: LaunchState,
	contextId: string,
): LaunchState["warnings"][number] | undefined {
	return launchState.warnings.find(
		(warning) =>
			warning.code === "context_mismatch" &&
			(warning.boundContextId === contextId ||
				warning.requestedContextId === contextId),
	);
}

export type { ContextRecommendation, ContextRecommendationCategory };
export { contextRecommendation };
