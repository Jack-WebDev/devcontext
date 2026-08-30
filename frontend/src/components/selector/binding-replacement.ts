import type { ProjectBindingState } from "../../lib/devctx-api";

interface BindingReplacement {
	boundContextId: string;
	replacementContextId: string;
}

function bindingReplacementForLaunch(
	binding: ProjectBindingState,
	launchedContextId: string | undefined,
): BindingReplacement | undefined {
	if (
		!binding.bound ||
		binding.dangling ||
		binding.contextId === undefined ||
		launchedContextId === undefined ||
		binding.contextId === launchedContextId
	) {
		return undefined;
	}

	return {
		boundContextId: binding.contextId,
		replacementContextId: launchedContextId,
	};
}

export type { BindingReplacement };
export { bindingReplacementForLaunch };
