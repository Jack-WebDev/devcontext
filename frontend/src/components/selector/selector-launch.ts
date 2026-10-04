import type {
	ApiResult,
	BindProjectRequest,
	DisplayError,
	LaunchProjectRequest,
	LaunchProjectResult,
	LaunchState,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	ProjectBindingState,
	UnbindProjectRequest,
} from "../../lib/devctx-api";
import type {
	ProjectLaunchJourneyResult,
	ProjectLaunchPending,
} from "../project-launch/project-launch-journey.js";
import { hasAccountIdentityMismatch } from "./account-identity-mismatch.js";
import { bindingReplacementForLaunch } from "./binding-replacement.js";
import {
	continueLaunchingSelectedContext,
	createLaunchRequestGuard,
	launchSelectedContext,
} from "./launch-action.js";
import {
	type LauncherSelection,
	type LauncherState,
	launcherSelection,
	selectingLauncherState,
} from "./launcher-state.js";
import { projectMemoryBindingContextId } from "./project-memory.js";
import {
	type ContextNavigationDirection,
	initialRovingContextId,
	initialSelectedContextId,
	nextKeyboardContextId,
	nextSelectedContextId,
} from "./selection-state.js";

interface SelectorLaunchAdapter {
	bindProject: (
		request: BindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	unbindProject: (
		request: UnbindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	preflightLaunchProject: (
		request: PreflightLaunchProjectRequest,
	) => Promise<ApiResult<PreflightLaunchProjectResult>>;
	launchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	closeSelector: () => Promise<void> | void;
	onCodingToolLaunched?: (result: LaunchProjectResult) => void;
}

interface SelectorLaunchOptions {
	launchSuccessClosesSelector: boolean;
	projectMemoryEnabled: boolean;
	requireContextMismatchConfirmation: boolean;
}

interface SelectorLaunchModuleDependencies {
	launchState: LaunchState;
	adapter: SelectorLaunchAdapter;
	options: SelectorLaunchOptions;
	onStateChange: (state: LauncherState) => void;
}

interface LaunchAttemptOptions {
	confirmContextMismatch?: boolean;
	confirmIdentityMismatch?: boolean;
	contextId?: string;
}

interface SelectorLaunchModule {
	getState(): LauncherState;
	updateInputs(
		launchState: LaunchState,
		adapter: SelectorLaunchAdapter,
		options: SelectorLaunchOptions,
	): void;
	reset(launchState: LaunchState): void;
	selectContext(contextId: string): void;
	navigateContext(
		contextId: string,
		direction: ContextNavigationDirection,
	): string | undefined;
	setRememberProject(rememberProject: boolean): void;
	dismissToSelection(): void;
	launch(options?: LaunchAttemptOptions): Promise<void>;
	continueLaunch(
		pending: ProjectLaunchPending,
		decision: "continue-after-review" | "launch-another",
	): Promise<void>;
	replaceBinding(): Promise<void>;
	keepCurrentBinding(): Promise<void>;
	removeDanglingBinding(): Promise<void>;
}

function createSelectorLaunchModule(
	dependencies: SelectorLaunchModuleDependencies,
): SelectorLaunchModule {
	let launchState = dependencies.launchState;
	let adapter = dependencies.adapter;
	let options = dependencies.options;
	let state = initialSelectorLaunchState(launchState);
	const requestGuard = createLaunchRequestGuard();

	function transition(next: LauncherState) {
		state = next;
		dependencies.onStateChange(next);
	}

	function currentSelection(): LauncherSelection | undefined {
		return launcherSelection(state);
	}

	async function finishSuccessfulLaunch(selection: LauncherSelection) {
		if (options.launchSuccessClosesSelector) {
			await adapter.closeSelector();
		}
		transition(selectingLauncherState(selection));
	}

	async function applyLaunchResult(
		result: ProjectLaunchJourneyResult,
		selection: LauncherSelection,
	) {
		if (result.kind === "preflight-review") {
			transition({
				status: "preflight_review",
				selection,
				pending: result.pending,
			});
			return;
		}
		if (result.kind === "running-environment-conflict") {
			transition({
				status: "existing_workspace",
				selection,
				conflict: result.conflict,
				pending: result.pending,
			});
			return;
		}
		if (result.kind === "failed") {
			if (
				result.error.code === "context_mismatch_requires_confirmation" &&
				result.error.contextMismatch
			) {
				transition({
					status: "context_mismatch",
					selection,
					error: result.error,
				});
				return;
			}
			transition({ status: "failure", selection, error: result.error });
			return;
		}

		adapter.onCodingToolLaunched?.(result.result);
		const replacement = bindingReplacementForLaunch(
			launchState.binding,
			result.result.context.id,
		);
		if (replacement !== undefined) {
			transition({
				status: "binding_replacement",
				selection,
				...replacement,
				pending: false,
			});
			return;
		}
		await finishSuccessfulLaunch(selection);
	}

	function preparationContextId(
		selection: LauncherSelection,
	): string | undefined {
		return projectMemoryBindingContextId({
			projectMemoryEnabled: options.projectMemoryEnabled,
			binding: launchState.binding,
			rememberProject: selection.rememberProject,
			selectedContextId: selection.selectedContextId,
		});
	}

	return {
		getState: () => state,
		updateInputs(nextLaunchState, nextAdapter, nextOptions) {
			launchState = nextLaunchState;
			adapter = nextAdapter;
			options = nextOptions;
		},
		reset(nextLaunchState) {
			launchState = nextLaunchState;
			transition(initialSelectorLaunchState(nextLaunchState));
		},
		selectContext(contextId) {
			if (state.status !== "selecting") return;
			const selectedContextId = nextSelectedContextId(
				launchState.contexts,
				contextId,
			);
			transition(
				selectingLauncherState({
					...state.selection,
					selectedContextId,
					rovingContextId: selectedContextId,
				}),
			);
		},
		navigateContext(contextId, direction) {
			if (state.status !== "selecting") return undefined;
			const nextContextId = nextKeyboardContextId(
				launchState.contexts,
				contextId,
				direction,
			);
			if (nextContextId === undefined) return undefined;
			transition(
				selectingLauncherState({
					...state.selection,
					selectedContextId: nextContextId,
					rovingContextId: nextContextId,
				}),
			);
			return nextContextId;
		},
		setRememberProject(rememberProject) {
			if (state.status !== "selecting") return;
			transition(
				selectingLauncherState({ ...state.selection, rememberProject }),
			);
		},
		dismissToSelection() {
			const selection = currentSelection();
			if (selection !== undefined)
				transition(selectingLauncherState(selection));
		},
		async launch(launchOptions = {}) {
			if (state.status === "dangling_binding") return;
			const selection = currentSelection();
			const contextId = launchOptions.contextId ?? selection?.selectedContextId;
			const context = launchState.contexts.find(
				(item) => item.id === contextId,
			);
			if (selection === undefined || context === undefined) {
				if (selection !== undefined)
					transition(selectingLauncherState(selection));
				return;
			}
			if (
				!launchOptions.confirmIdentityMismatch &&
				hasAccountIdentityMismatch(context)
			) {
				transition({
					status: "identity_mismatch",
					selection,
					contextId: context.id,
				});
				return;
			}

			await requestGuard.run(async () => {
				transition({ status: "preflighting", selection });
				try {
					const result = await launchSelectedContext({
						projectPath: launchState.project.path,
						selectedContextId: context.id,
						bindingContextId: preparationContextId(selection),
						confirmContextMismatch:
							launchOptions.confirmContextMismatch ||
							!options.requireContextMismatchConfirmation,
						onLaunchStarting: (preflight) =>
							transition({
								status: "launching",
								selection,
								groups: preflight.groups,
								steps: preflight.verificationSteps,
							}),
						bindProject: adapter.bindProject,
						preflightLaunchProject: adapter.preflightLaunchProject,
						launchProject: adapter.launchProject,
					});
					if (result !== undefined) await applyLaunchResult(result, selection);
				} catch (error) {
					transition({
						status: "failure",
						selection,
						error: unexpectedLaunchError(error),
					});
				}
			});
		},
		async continueLaunch(pending, decision) {
			const selection = currentSelection();
			if (selection === undefined) return;
			await requestGuard.run(async () => {
				try {
					const result = await continueLaunchingSelectedContext({
						pending,
						decision,
						projectPath: launchState.project.path,
						bindingContextId: preparationContextId(selection),
						onLaunchStarting: (preflight) =>
							transition({
								status: "launching",
								selection,
								groups: preflight.groups,
								steps: preflight.verificationSteps,
							}),
						bindProject: adapter.bindProject,
						launchProject: adapter.launchProject,
					});
					await applyLaunchResult(result, selection);
				} catch (error) {
					transition({
						status: "failure",
						selection,
						error: unexpectedLaunchError(error),
					});
				}
			});
		},
		async replaceBinding() {
			if (state.status !== "binding_replacement" || state.pending) return;
			const pendingState = { ...state, pending: true, error: undefined };
			transition(pendingState);
			try {
				const result = await adapter.bindProject({
					projectPath: launchState.project.path,
					contextId: state.replacementContextId,
				});
				if (!result.ok) {
					transition({ ...pendingState, pending: false, error: result.error });
					return;
				}
				await finishSuccessfulLaunch(state.selection);
			} catch (error) {
				transition({
					...pendingState,
					pending: false,
					error: unexpectedBindingError(error),
				});
			}
		},
		async keepCurrentBinding() {
			if (state.status !== "binding_replacement" || state.pending) return;
			await finishSuccessfulLaunch(state.selection);
		},
		async removeDanglingBinding() {
			if (state.status !== "dangling_binding" || state.pending) return;
			const pendingState = { ...state, pending: true, error: undefined };
			transition(pendingState);
			try {
				const result = await adapter.unbindProject({
					projectPath: launchState.project.path,
				});
				if (!result.ok) {
					transition({ ...pendingState, pending: false, error: result.error });
					return;
				}
				transition(selectingLauncherState(state.selection));
			} catch (error) {
				transition({
					...pendingState,
					pending: false,
					error: unexpectedBindingRemovalError(error),
				});
			}
		},
	};
}

function initialSelectorLaunchState(launchState: LaunchState): LauncherState {
	const selection: LauncherSelection = {
		selectedContextId: initialSelectedContextId(launchState),
		rovingContextId: initialRovingContextId(launchState),
		rememberProject: false,
	};
	return launchState.binding.dangling
		? { status: "dangling_binding", selection, pending: false }
		: selectingLauncherState(selection);
}

function unexpectedLaunchError(error: unknown): DisplayError {
	return {
		code: "unexpected_error",
		message: error instanceof Error ? error.message : "Launch failed.",
		recovery: "Retry the launch. If it keeps failing, review diagnostics.",
	};
}

function unexpectedBindingError(error: unknown): DisplayError {
	return {
		code: "unexpected_error",
		message:
			error instanceof Error
				? error.message
				: "Could not remember this context.",
		recovery: "Try again or keep the current remembered context.",
	};
}

function unexpectedBindingRemovalError(error: unknown): DisplayError {
	return {
		code: "unexpected_error",
		message:
			error instanceof Error
				? error.message
				: "Could not remove the remembered context.",
		recovery:
			"Try again or choose a context for this launch without removing it.",
	};
}

export type {
	LaunchAttemptOptions,
	SelectorLaunchAdapter,
	SelectorLaunchModule,
	SelectorLaunchModuleDependencies,
	SelectorLaunchOptions,
};
export { createSelectorLaunchModule, initialSelectorLaunchState };
