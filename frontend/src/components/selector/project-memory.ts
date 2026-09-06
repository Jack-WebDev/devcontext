import type { ProjectBindingState } from "../../lib/devctx-api";

function projectMemoryBindingContextId({
	projectMemoryEnabled,
	binding,
	rememberProject,
	selectedContextId,
}: {
	projectMemoryEnabled: boolean;
	binding: ProjectBindingState;
	rememberProject: boolean;
	selectedContextId?: string;
}): string | undefined {
	if (
		!projectMemoryEnabled ||
		!rememberProject ||
		selectedContextId === undefined ||
		binding.bound ||
		binding.dangling
	) {
		return undefined;
	}

	return selectedContextId;
}

export { projectMemoryBindingContextId };
