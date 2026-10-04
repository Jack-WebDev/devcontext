import { type KeyboardEvent, useEffect, useRef, useState } from "react";

import type {
	ApiResult,
	BindProjectRequest,
	ContextState,
	LaunchProjectRequest,
	LaunchProjectResult,
	LaunchState,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	ProjectBindingState,
	UnbindProjectRequest,
} from "../../lib/devctx-api";
import { contextPositionFromShortcut } from "../command-palette/shortcut";
import type { ProjectLaunchPending } from "../project-launch/project-launch-journey.js";
import { RunningEnvironmentConflictDialog } from "../running/RunningEnvironmentConflictDialog";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { AccountIdentityMismatchDialog } from "./AccountIdentityMismatchDialog";
import { ContextChoiceList } from "./ContextChoiceList";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
import { cancelSelector } from "./cancel-action";
import { DanglingBindingDialog } from "./DanglingBindingDialog";
import { missingDefaultContextIds } from "./default-context-actions";
import {
	FirstRunWelcome,
	shouldRenderFirstRunWelcome,
} from "./FirstRunWelcome";
import { LaunchFailureView } from "./LaunchFailureView";
import { LaunchProgressView } from "./LaunchProgressView";
import {
	defaultLaunchSuccessCloseBehavior,
	type LaunchSuccessCloseBehavior,
	shouldCloseSelectorAfterLaunch,
} from "./launch-success-close-behavior";
import {
	type LauncherState,
	launcherSelection,
	launcherStateIsPending,
	selectingLauncherState,
} from "./launcher-state";
import { PreflightReviewView } from "./PreflightReviewView";
import { ProjectIdentity } from "./ProjectIdentity";
import {
	canRememberProject,
	RememberProjectControl,
} from "./RememberProjectControl";
import { ReplaceBindingDialog } from "./ReplaceBindingDialog";
import { SelectorActions } from "./SelectorActions";
import { SelectorConfidenceSummary } from "./SelectorConfidenceSummary";
import { SelectorLayout } from "./SelectorLayout";
import { SingleContextLaunchView } from "./SingleContextLaunchView";
import type { ContextNavigationDirection } from "./selection-state";
import {
	canLaunchSelectedContextFromKeyboard,
	escapeKeyboardAction,
} from "./selector-keyboard";
import {
	createSelectorLaunchModule,
	initialSelectorLaunchState,
	type LaunchAttemptOptions,
	type SelectorLaunchModule,
} from "./selector-launch.js";
import { WelcomeView } from "./WelcomeView";

interface SelectorViewProps {
	launchState: LaunchState;
	onBindProject: (
		request: BindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	onUnbindProject: (
		request: UnbindProjectRequest,
	) => Promise<ApiResult<ProjectBindingState>>;
	onPreflightLaunchProject: (
		request: PreflightLaunchProjectRequest,
	) => Promise<ApiResult<PreflightLaunchProjectResult>>;
	onLaunchProject: (
		request: LaunchProjectRequest,
	) => Promise<ApiResult<LaunchProjectResult>>;
	onCancel: () => Promise<void> | void;
	launchSuccessCloseBehavior?: LaunchSuccessCloseBehavior;
	onRunDiagnostics?: () => void;
	onCodingToolLaunched?: (result: LaunchProjectResult) => void;
	showLaunchVerification?: boolean;
	projectMemoryEnabled?: boolean;
	requireContextMismatchConfirmation?: boolean;
	showOnboardingReplay?: boolean;
	onDismissOnboardingReplay?: () => void;
	onStartContextCreation: () => void;
	onRetryDetection?: () => void;
}

function SelectorView({
	launchState,
	onBindProject,
	onUnbindProject,
	onPreflightLaunchProject,
	onLaunchProject,
	onCancel,
	launchSuccessCloseBehavior = defaultLaunchSuccessCloseBehavior,
	onRunDiagnostics,
	onCodingToolLaunched,
	showLaunchVerification = true,
	projectMemoryEnabled = true,
	requireContextMismatchConfirmation = true,
	showOnboardingReplay = false,
	onDismissOnboardingReplay,
	onStartContextCreation,
	onRetryDetection,
}: SelectorViewProps) {
	const [launcherState, setLauncherState] = useState<LauncherState>(() =>
		initialSelectorLaunchState(launchState),
	);
	const [showContextChoices, setShowContextChoices] = useState(false);
	const [contextSearch, setContextSearch] = useState("");
	const contextButtonRefs = useRef(new Map<string, HTMLButtonElement>());
	const launchModuleRef = useRef<SelectorLaunchModule | undefined>(undefined);
	if (launchModuleRef.current === undefined) {
		launchModuleRef.current = createSelectorLaunchModule({
			launchState,
			adapter: {
				bindProject: onBindProject,
				unbindProject: onUnbindProject,
				preflightLaunchProject: onPreflightLaunchProject,
				launchProject: onLaunchProject,
				closeSelector: onCancel,
				onCodingToolLaunched,
			},
			options: {
				launchSuccessClosesSelector: shouldCloseSelectorAfterLaunch(
					launchSuccessCloseBehavior,
				),
				projectMemoryEnabled,
				requireContextMismatchConfirmation,
			},
			onStateChange: setLauncherState,
		});
	}
	const launchModule = launchModuleRef.current;
	const selection = launcherSelection(launcherState);
	const selectedContextId = selection?.selectedContextId;
	const rovingContextId = selection?.rovingContextId;
	const rememberProject = selection?.rememberProject ?? false;
	const mismatchDialogOpen =
		launcherState.status === "context_mismatch" ||
		launcherState.status === "identity_mismatch";
	const danglingBindingDialogOpen = launcherState.status === "dangling_binding";
	const preflightReviewOpen = launcherState.status === "preflight_review";
	const selectedContext = launchState.contexts.find(
		(context) => context.id === selectedContextId,
	);
	const launchPending = launcherStateIsPending(launcherState);
	const cancellationPending = launchPending;
	const singleHealthyContext = singleHealthyLaunchContext(launchState);
	const showSingleContextConfirmation =
		singleHealthyContext !== undefined && !showContextChoices;
	const showWelcome =
		shouldRenderFirstRunWelcome(launchState) && launchState.firstRun;
	const launchInProgress =
		launcherState.status === "preflighting" ||
		launcherState.status === "launching";
	const keyboardLaunchAvailable = canLaunchSelectedContextFromKeyboard({
		selectedContextId,
		launchPending: cancellationPending,
		mismatchDialogOpen:
			mismatchDialogOpen || danglingBindingDialogOpen || preflightReviewOpen,
		dialogOpen:
			mismatchDialogOpen ||
			danglingBindingDialogOpen ||
			preflightReviewOpen ||
			launcherState.status === "existing_workspace",
	});

	useEffect(() => {
		launchModule.updateInputs(
			launchState,
			{
				bindProject: onBindProject,
				unbindProject: onUnbindProject,
				preflightLaunchProject: onPreflightLaunchProject,
				launchProject: onLaunchProject,
				closeSelector: onCancel,
				onCodingToolLaunched,
			},
			{
				launchSuccessClosesSelector: shouldCloseSelectorAfterLaunch(
					launchSuccessCloseBehavior,
				),
				projectMemoryEnabled,
				requireContextMismatchConfirmation,
			},
		);
	}, [
		launchModule,
		launchState,
		onBindProject,
		onCancel,
		onCodingToolLaunched,
		onLaunchProject,
		onPreflightLaunchProject,
		onUnbindProject,
		launchSuccessCloseBehavior,
		projectMemoryEnabled,
		requireContextMismatchConfirmation,
	]);

	useEffect(() => {
		launchModule.reset(launchState);
		setShowContextChoices(false);
		setContextSearch("");
	}, [launchModule, launchState]);

	function setContextButtonRef(contextId: string) {
		return (button: HTMLButtonElement | null) => {
			if (button) {
				contextButtonRefs.current.set(contextId, button);
			} else {
				contextButtonRefs.current.delete(contextId);
			}
		};
	}

	function handleSelectContext(contextId: string) {
		launchModule.selectContext(contextId);
	}

	function handleContextNavigation(
		contextId: string,
		direction: ContextNavigationDirection,
	) {
		const nextContextId = launchModule.navigateContext(contextId, direction);
		if (nextContextId === undefined) {
			return;
		}

		contextButtonRefs.current.get(nextContextId)?.focus();
	}

	function handleSelectorKeyDown(event: KeyboardEvent<HTMLDivElement>) {
		const contextPosition = contextPositionFromShortcut(event);
		if (
			contextPosition !== undefined &&
			!cancellationPending &&
			!(
				mismatchDialogOpen ||
				preflightReviewOpen ||
				launcherState.status === "existing_workspace"
			)
		) {
			const context = launchState.contexts[contextPosition];
			if (context !== undefined) {
				event.preventDefault();
				handleSelectContext(context.id);
				contextButtonRefs.current.get(context.id)?.focus();
			}
			return;
		}

		if (event.key !== "Escape") {
			return;
		}

		const action = escapeKeyboardAction({
			selectedContextId,
			launchPending: cancellationPending,
			mismatchDialogOpen,
			dialogOpen:
				mismatchDialogOpen ||
				preflightReviewOpen ||
				launcherState.status === "existing_workspace",
		});
		if (action === "none") {
			return;
		}

		event.preventDefault();
		event.stopPropagation();

		if (action === "close-dialog" && selection !== undefined) {
			launchModule.dismissToSelection();
			return;
		}

		void cancelSelector({
			closeSelector: onCancel,
			canCancel: !cancellationPending,
		});
	}

	function handleLaunch(options?: LaunchAttemptOptions) {
		return launchModule.launch(options);
	}

	function continueLaunch(
		pending: ProjectLaunchPending,
		decision: "continue-after-review" | "launch-another",
	) {
		return launchModule.continueLaunch(pending, decision);
	}

	function handleBindingReplacement() {
		return launchModule.replaceBinding();
	}

	function handleDanglingBindingRemoval() {
		return launchModule.removeDanglingBinding();
	}

	function handleRememberProjectChange(rememberProject: boolean) {
		if (canRememberProject(launchState.binding)) {
			launchModule.setRememberProject(rememberProject);
		}
	}

	if (launchInProgress) {
		return (
			<LaunchProgressView
				projectName={launchState.project.name}
				contextName={selectedContext?.name ?? "selected context"}
				showVerification={showLaunchVerification}
				groups={
					launcherState.status === "launching"
						? launcherState.groups
						: undefined
				}
				steps={
					launcherState.status === "launching" ? launcherState.steps : undefined
				}
			/>
		);
	}

	return (
		<section
			className="space-y-8"
			aria-label="Project launch options"
			onKeyDown={handleSelectorKeyDown}
		>
			{showWelcome ? (
				<WelcomeView onCreateFirstContext={onStartContextCreation} />
			) : showOnboardingReplay ? (
				<>
					<ProjectIdentity project={launchState.project} />
					<FirstRunWelcome
						launchState={launchState}
						replay={showOnboardingReplay && !launchState.firstRun}
						onContinue={
							showOnboardingReplay && !launchState.firstRun
								? onDismissOnboardingReplay
								: undefined
						}
					/>
				</>
			) : (
				<SelectorLayout
					projectIdentity={<ProjectIdentity project={launchState.project} />}
					contextCards={
						<>
							{showSingleContextConfirmation ? (
								<SingleContextLaunchView
									context={singleHealthyContext}
									projectName={launchState.project.name}
									onChooseAnother={() => setShowContextChoices(true)}
								/>
							) : (
								<ContextChoiceList
									launchState={launchState}
									selectedContextId={selectedContextId}
									rovingContextId={rovingContextId}
									launchPending={launchPending}
									keyboardLaunchAvailable={keyboardLaunchAvailable}
									search={contextSearch}
									onSearchChange={setContextSearch}
									buttonRef={setContextButtonRef}
									onSelect={handleSelectContext}
									onNavigate={handleContextNavigation}
									onLaunch={() => void handleLaunch()}
									onProviderSetup={(contextId) => {
										handleSelectContext(contextId);
										void handleLaunch({ contextId });
									}}
								/>
							)}

							<MissingDefaultContextActions
								launchState={launchState}
								onStartContextCreation={onStartContextCreation}
							/>
						</>
					}
					confidenceSummary={
						showSingleContextConfirmation ? null : (
							<SelectorConfidenceSummary
								context={selectedContext}
								project={launchState.project}
								onRetryDetection={onRetryDetection}
							/>
						)
					}
					rememberControl={
						<RememberProjectControl
							binding={launchState.binding}
							contexts={launchState.contexts}
							rememberProject={rememberProject}
							projectMemoryEnabled={projectMemoryEnabled}
							selectedContextId={selectedContextId}
							disabled={launchPending}
							onRememberProjectChange={handleRememberProjectChange}
						/>
					}
					launchActions={
						<>
							{launcherState.status === "preflight_review" ? (
								<PreflightReviewView
									projectName={launcherState.pending.preflight.project.name}
									contextName={launcherState.pending.preflight.context.name}
									preflight={launcherState.pending.preflight}
									onFixFirst={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onLaunchWithoutIt={
										launcherState.pending.preflight.groups.some(
											(group) => group.blocking,
										)
											? undefined
											: () =>
													void continueLaunch(
														launcherState.pending,
														"continue-after-review",
													)
									}
								/>
							) : null}

							{launcherState.status === "failure" ? (
								<LaunchFailureView
									error={launcherState.error}
									onRetry={() => void handleLaunch()}
									onRunDiagnostics={onRunDiagnostics}
									onChooseAnotherContext={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onCancel={() =>
										void cancelSelector({ closeSelector: onCancel })
									}
								/>
							) : null}

							{launcherState.status === "binding_replacement" ? (
								<ReplaceBindingDialog
									boundContextName={
										launchState.contexts.find(
											(context) => context.id === launcherState.boundContextId,
										)?.name ?? launcherState.boundContextId
									}
									replacementContextName={
										launchState.contexts.find(
											(context) =>
												context.id === launcherState.replacementContextId,
										)?.name ?? launcherState.replacementContextId
									}
									pending={launcherState.pending}
									error={launcherState.error}
									onKeepCurrent={() => void launchModule.keepCurrentBinding()}
									onReplace={() => void handleBindingReplacement()}
								/>
							) : null}

							{launcherState.status === "dangling_binding" ? (
								<DanglingBindingDialog
									missingContextId={launchState.binding.missingContextId}
									pending={launcherState.pending}
									error={launcherState.error}
									onChooseContext={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onRemoveBinding={() => void handleDanglingBindingRemoval()}
									onCancel={() =>
										void cancelSelector({ closeSelector: onCancel })
									}
								/>
							) : null}

							{launcherState.status === "context_mismatch" &&
							launcherState.error.contextMismatch ? (
								<ContextMismatchDialog
									mismatch={launcherState.error.contextMismatch}
									contexts={launchState.contexts}
									launchPending={launchPending}
									onCancel={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onUseRememberedContext={() => {
										const mismatch = launcherState.error.contextMismatch;
										if (mismatch !== undefined) {
											void handleLaunch({ contextId: mismatch.boundContextId });
										}
									}}
									onOpenAnyway={() => {
										const mismatch = launcherState.error.contextMismatch;
										if (mismatch !== undefined) {
											void handleLaunch({
												contextId: mismatch.requestedContextId,
												confirmContextMismatch: true,
												confirmIdentityMismatch: true,
											});
										}
									}}
								/>
							) : null}

							{launcherState.status === "identity_mismatch" ? (
								<AccountIdentityMismatchDialog
									contextName={
										launchState.contexts.find(
											(context) => context.id === launcherState.contextId,
										)?.name ?? "selected context"
									}
									launchPending={launchPending}
									onCancel={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onReviewConfiguration={() => {
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										);
										onRunDiagnostics?.();
									}}
									onLaunchAnyway={() =>
										void handleLaunch({
											contextId: launcherState.contextId,
											confirmIdentityMismatch: true,
										})
									}
								/>
							) : null}

							{launcherState.status === "existing_workspace" ? (
								<RunningEnvironmentConflictDialog
									conflict={launcherState.conflict}
									launchPending={launchPending}
									onCancel={() =>
										setLauncherState(
											selectingLauncherState(launcherState.selection),
										)
									}
									onLaunchAnother={() =>
										void continueLaunch(launcherState.pending, "launch-another")
									}
								/>
							) : null}

							<SelectorActions
								launchDisabled={
									selectedContextId === undefined ||
									danglingBindingDialogOpen ||
									preflightReviewOpen
								}
								launchPending={launchPending}
								projectName={launchState.project.name}
								contextName={selectedContext?.name}
								confidence={selectedContext?.confidence}
								onLaunch={() => void handleLaunch()}
								onCancel={() =>
									void cancelSelector({ closeSelector: onCancel })
								}
							/>
						</>
					}
				/>
			)}
		</section>
	);
}

function MissingDefaultContextActions({
	launchState,
	onStartContextCreation,
}: {
	launchState: LaunchState;
	onStartContextCreation: () => void;
}) {
	const missingDefaults = missingDefaultContextIds(launchState.contexts);
	const missingPersonal = missingDefaults.includes("personal");
	const missingCompany = missingDefaults.includes("company");

	if (!missingPersonal && !missingCompany) {
		return null;
	}

	return (
		<Card
			as="section"
			size="sm"
			className="border border-border bg-muted/30 py-0"
			aria-label="Add default contexts"
		>
			<CardContent className="p-4">
				<div className="flex flex-wrap items-center justify-between gap-3">
					<div className="min-w-0">
						<h3 className="text-sm font-semibold">
							Add another default context
						</h3>
						<p className="mt-1 text-sm text-muted-foreground">
							Create the missing Personal or Company context when this machine
							needs both identities.
						</p>
					</div>
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={onStartContextCreation}
					>
						Create a context
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function singleHealthyLaunchContext(
	launchState: LaunchState,
): ContextState | undefined {
	const [context] = launchState.contexts;
	if (
		launchState.contexts.length !== 1 ||
		context?.tool.status !== "ready" ||
		context.confidence?.status !== "ready"
	) {
		return undefined;
	}
	return context;
}

export { SelectorView };
