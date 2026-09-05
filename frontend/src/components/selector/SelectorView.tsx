import { type KeyboardEvent, useEffect, useRef, useState } from "react";

import type {
	ApiResult,
	BindProjectRequest,
	ContextState,
	CreateContextResult,
	DisplayError,
	LaunchProjectRequest,
	LaunchProjectResult,
	LaunchState,
	PreflightLaunchProjectRequest,
	PreflightLaunchProjectResult,
	ProjectBindingState,
	ProviderCredentialSession,
	UnbindProjectRequest,
} from "../../lib/devctx-api";
import { contextPositionFromShortcut } from "../command-palette/shortcut";
import { RunningEnvironmentConflictDialog } from "../running/RunningEnvironmentConflictDialog";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { AccountIdentityMismatchDialog } from "./AccountIdentityMismatchDialog";
import { hasAccountIdentityMismatch } from "./account-identity-mismatch";
import { bindingReplacementForLaunch } from "./binding-replacement";
import { ContextChoiceList } from "./ContextChoiceList";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
import { DanglingBindingDialog } from "./DanglingBindingDialog";
import { cancelSelector } from "./cancel-action";
import { missingDefaultContextIds } from "./default-context-actions";
import {
	FirstRunWelcome,
	shouldRenderFirstRunWelcome,
} from "./FirstRunWelcome";
import { GuiErrorNotice } from "./GuiErrorNotice";
import { LaunchFailureView } from "./LaunchFailureView";
import { LaunchVerificationProgress } from "./LaunchVerificationProgress";
import {
	createLaunchRequestGuard,
	launchSelectedContext,
} from "./launch-action";
import {
	defaultLaunchSuccessCloseBehavior,
	type LaunchSuccessCloseBehavior,
	shouldCloseSelectorAfterLaunch,
} from "./launch-success-close-behavior";
import {
	type LauncherSelection,
	type LauncherState,
	launcherSelection,
	launcherStateIsPending,
	selectingLauncherState,
} from "./launcher-state";
import { ProjectIdentity } from "./ProjectIdentity";
import {
	ProviderCredentialClassification,
	type ProviderSessionAssignments,
} from "./ProviderCredentialClassification.js";
import {
	canRememberProject,
	RememberProjectControl,
} from "./RememberProjectControl";
import { ReplaceBindingDialog } from "./ReplaceBindingDialog";
import { SelectorActions } from "./SelectorActions";
import { SelectorConfidenceSummary } from "./SelectorConfidenceSummary";
import { SelectorLayout } from "./SelectorLayout";
import { SingleContextLaunchView } from "./SingleContextLaunchView";
import {
	type ContextNavigationDirection,
	initialRovingContextId,
	initialSelectedContextId,
	nextKeyboardContextId,
	nextSelectedContextId,
} from "./selection-state";
import {
	canLaunchSelectedContextFromKeyboard,
	escapeKeyboardAction,
} from "./selector-keyboard";

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
	onCreatePersonalContext?: (
		importProviderIds: string[],
	) => Promise<ApiResult<CreateContextResult>>;
	onCreateCompanyContext?: (
		importProviderIds: string[],
	) => Promise<ApiResult<CreateContextResult>>;
	onRunDiagnostics?: () => void;
	onCodingToolLaunched?: (result: LaunchProjectResult) => void;
	showLaunchVerification?: boolean;
	showOnboardingReplay?: boolean;
	onDismissOnboardingReplay?: () => void;
}

function SelectorView({
	launchState,
	onBindProject,
	onUnbindProject,
	onPreflightLaunchProject,
	onLaunchProject,
	onCancel,
	launchSuccessCloseBehavior = defaultLaunchSuccessCloseBehavior,
	onCreatePersonalContext,
	onCreateCompanyContext,
	onRunDiagnostics,
	onCodingToolLaunched,
	showLaunchVerification = true,
	showOnboardingReplay = false,
	onDismissOnboardingReplay,
}: SelectorViewProps) {
	const [launcherState, setLauncherState] = useState<LauncherState>(() =>
		initialLauncherState(launchState),
	);
	const [onboardingPendingContextId, setOnboardingPendingContextId] = useState<
		string | undefined
	>(undefined);
	const [onboardingError, setOnboardingError] = useState<
		DisplayError | undefined
	>(undefined);
	const [providerSessionAssignments, setProviderSessionAssignments] =
		useState<ProviderSessionAssignments>({});
	const [showContextChoices, setShowContextChoices] = useState(false);
	const [contextSearch, setContextSearch] = useState("");
	const contextButtonRefs = useRef(new Map<string, HTMLButtonElement>());
	const launchGuard = useRef(createLaunchRequestGuard());
	const selection = launcherSelection(launcherState);
	const selectedContextId = selection?.selectedContextId;
	const rovingContextId = selection?.rovingContextId;
	const rememberProject = selection?.rememberProject ?? false;
	const mismatchDialogOpen =
		launcherState.status === "context_mismatch" ||
		launcherState.status === "identity_mismatch";
	const danglingBindingDialogOpen = launcherState.status === "dangling_binding";
	const selectedContext = launchState.contexts.find(
		(context) => context.id === selectedContextId,
	);
	const launchPending = launcherStateIsPending(launcherState);
	const cancellationPending =
		launchPending || onboardingPendingContextId !== undefined;
	const launchBlocked = selectedContextConfidenceBlocked(selectedContext);
	const singleHealthyContext = singleHealthyLaunchContext(launchState);
	const showSingleContextConfirmation =
		singleHealthyContext !== undefined && !showContextChoices;
	const keyboardLaunchAvailable = canLaunchSelectedContextFromKeyboard({
		selectedContextId,
		launchPending: cancellationPending || launchBlocked,
		mismatchDialogOpen: mismatchDialogOpen || danglingBindingDialogOpen,
		dialogOpen:
			mismatchDialogOpen ||
			danglingBindingDialogOpen ||
			launcherState.status === "existing_workspace",
	});

	useEffect(() => {
		setLauncherState(initialLauncherState(launchState));
		setOnboardingPendingContextId(undefined);
		setOnboardingError(undefined);
		setProviderSessionAssignments({});
		setShowContextChoices(false);
		setContextSearch("");
		launchGuard.current = createLaunchRequestGuard();
	}, [launchState]);

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
		if (launcherState.status !== "selecting") {
			return;
		}
		const nextContextId = nextSelectedContextId(
			launchState.contexts,
			contextId,
		);
		setLauncherState(
			selectingLauncherState({
				...launcherState.selection,
				selectedContextId: nextContextId,
				rovingContextId: nextContextId,
			}),
		);
	}

	function handleContextNavigation(
		contextId: string,
		direction: ContextNavigationDirection,
	) {
		if (launcherState.status !== "selecting") {
			return;
		}

		const nextContextId = nextKeyboardContextId(
			launchState.contexts,
			contextId,
			direction,
		);
		if (nextContextId === undefined) {
			return;
		}

		setLauncherState(
			selectingLauncherState({
				...launcherState.selection,
				selectedContextId: nextContextId,
				rovingContextId: nextContextId,
			}),
		);
		contextButtonRefs.current.get(nextContextId)?.focus();
	}

	function handleSelectorKeyDown(event: KeyboardEvent<HTMLDivElement>) {
		const contextPosition = contextPositionFromShortcut(event);
		if (
			contextPosition !== undefined &&
			!cancellationPending &&
			!(mismatchDialogOpen || launcherState.status === "existing_workspace")
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
				mismatchDialogOpen || launcherState.status === "existing_workspace",
		});
		if (action === "none") {
			return;
		}

		event.preventDefault();
		event.stopPropagation();

		if (action === "close-dialog" && selection !== undefined) {
			setLauncherState(selectingLauncherState(selection));
			return;
		}

		void cancelSelector({
			closeSelector: onCancel,
			canCancel: !cancellationPending,
		});
	}

	async function handleLaunch({
		confirmContextMismatch = false,
		contextId = selectedContextId,
		allowExistingEnvironmentLaunch = false,
		confirmIdentityMismatch = false,
	}: LaunchAttemptOptions = {}) {
		if (launcherState.status === "dangling_binding") {
			return;
		}
		const currentSelection = launcherSelection(launcherState);
		if (currentSelection === undefined) {
			return;
		}
		const contextToLaunch = launchState.contexts.find(
			(context) => context.id === contextId,
		);
		if (selectedContextConfidenceBlocked(contextToLaunch)) {
			return;
		}
		if (
			!confirmIdentityMismatch &&
			hasAccountIdentityMismatch(contextToLaunch)
		) {
			if (contextId !== undefined) {
				setLauncherState({
					status: "identity_mismatch",
					selection: currentSelection,
					contextId,
				});
			}
			return;
		}

		await launchGuard.current.run(async () => {
			setLauncherState({
				status: "preflighting",
				selection: currentSelection,
			});

			try {
				const result = await launchSelectedContext({
					projectPath: launchState.project.path,
					selectedContextId: contextId,
					bindingContextId:
						canRememberProject(launchState.binding) && rememberProject
							? contextId
							: undefined,
					confirmContextMismatch,
					allowExistingEnvironmentLaunch,
					onPreflightComplete: (preflight) => {
						setLauncherState({
							status: "launching",
							selection: currentSelection,
							steps: preflight.verificationSteps,
						});
					},
					bindProject: onBindProject,
					preflightLaunchProject: onPreflightLaunchProject,
					launchProject: onLaunchProject,
				});

				if (result && "runningEnvironmentConflict" in result) {
					setLauncherState({
						status: "existing_workspace",
						selection: currentSelection,
						conflict: result.runningEnvironmentConflict,
					});
				} else if (result?.ok) {
					if ("project" in result.data && "context" in result.data) {
						onCodingToolLaunched?.(result.data);
						const replacement = bindingReplacementForLaunch(
							launchState.binding,
							result.data.context.id,
						);
						if (replacement !== undefined) {
							setLauncherState({
								status: "binding_replacement",
								selection: currentSelection,
								...replacement,
								pending: false,
							});
							return;
						}
					}
					await finishSuccessfulLaunch(currentSelection);
				} else if (result && !result.ok) {
					if (
						result.error.code === "context_mismatch_requires_confirmation" &&
						result.error.contextMismatch
					) {
						setLauncherState({
							status: "context_mismatch",
							selection: currentSelection,
							error: result.error,
						});
					} else {
						setLauncherState({
							status: "failure",
							selection: currentSelection,
							error: result.error,
						});
					}
				} else {
					setLauncherState(selectingLauncherState(currentSelection));
				}
			} catch (error) {
				setLauncherState({
					status: "failure",
					selection: currentSelection,
					error: unexpectedLaunchError(error),
				});
			}
		});
	}

	async function finishSuccessfulLaunch(selection: LauncherSelection) {
		if (shouldCloseSelectorAfterLaunch(launchSuccessCloseBehavior)) {
			await cancelSelector({ closeSelector: onCancel });
		}
		setLauncherState(selectingLauncherState(selection));
	}

	async function handleBindingReplacement() {
		if (launcherState.status !== "binding_replacement" || launcherState.pending) {
			return;
		}

		setLauncherState({ ...launcherState, pending: true, error: undefined });
		try {
			const result = await onBindProject({
				projectPath: launchState.project.path,
				contextId: launcherState.replacementContextId,
			});
			if (!result.ok) {
				setLauncherState({ ...launcherState, pending: false, error: result.error });
				return;
			}
			await finishSuccessfulLaunch(launcherState.selection);
		} catch (error) {
			setLauncherState({
				...launcherState,
				pending: false,
				error: unexpectedBindingError(error),
			});
		}
	}

	async function handleDanglingBindingRemoval() {
		if (launcherState.status !== "dangling_binding" || launcherState.pending) {
			return;
		}

		setLauncherState({ ...launcherState, pending: true, error: undefined });
		try {
			const result = await onUnbindProject({
				projectPath: launchState.project.path,
			});
			if (!result.ok) {
				setLauncherState({ ...launcherState, pending: false, error: result.error });
				return;
			}
			setLauncherState(selectingLauncherState(launcherState.selection));
		} catch (error) {
			setLauncherState({
				...launcherState,
				pending: false,
				error: unexpectedBindingRemovalError(error),
			});
		}
	}

	function handleRememberProjectChange(rememberProject: boolean) {
		if (
			launcherState.status !== "selecting" ||
			!canRememberProject(launchState.binding)
		) {
			return;
		}
		setLauncherState(
			selectingLauncherState({
				...launcherState.selection,
				rememberProject,
			}),
		);
	}

	function handleClassifyProviderSession(
		providerId: string,
		contextId: "personal" | "company",
	) {
		setProviderSessionAssignments((current) => ({
			...current,
			[providerId]: contextId,
		}));
	}

	function importProviderIdsForContext(
		contextId: "personal" | "company",
	): string[] {
		return launchState.providerCredentialSessions
			.filter(
				(session) =>
					providerSessionAssignments[session.providerId] === contextId,
			)
			.map((session) => session.providerId);
	}

	async function handleCreateContext(
		contextId: "personal" | "company",
		createContext: (
			importProviderIds: string[],
		) => Promise<ApiResult<CreateContextResult>>,
	) {
		setOnboardingPendingContextId(contextId);
		setOnboardingError(undefined);

		try {
			const result = await createContext(
				importProviderIdsForContext(contextId),
			);
			if (!result.ok) {
				setOnboardingError(result.error);
			}
		} finally {
			setOnboardingPendingContextId(undefined);
		}
	}

	return (
		<section
			className="space-y-8"
			aria-label="Project launch options"
			onKeyDown={handleSelectorKeyDown}
		>
			{shouldRenderFirstRunWelcome(launchState) || showOnboardingReplay ? (
				<>
					<ProjectIdentity project={launchState.project} />
					<FirstRunWelcome
						launchState={launchState}
						providerCredentialSessions={launchState.providerCredentialSessions}
						providerSessionAssignments={providerSessionAssignments}
						pendingContextId={onboardingPendingContextId}
						error={onboardingError}
						onClassifyProviderSession={handleClassifyProviderSession}
						onCreatePersonal={
							onCreatePersonalContext
								? () =>
										void handleCreateContext(
											"personal",
											onCreatePersonalContext,
										)
								: undefined
						}
						onCreateCompany={
							onCreateCompanyContext
								? () =>
										void handleCreateContext("company", onCreateCompanyContext)
								: undefined
						}
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
								providerCredentialSessions={
									launchState.providerCredentialSessions
								}
								providerSessionAssignments={providerSessionAssignments}
								pendingContextId={onboardingPendingContextId}
								error={onboardingError}
								onClassifyProviderSession={handleClassifyProviderSession}
								onCreatePersonal={
									onCreatePersonalContext
										? () =>
												void handleCreateContext(
													"personal",
													onCreatePersonalContext,
												)
										: undefined
								}
								onCreateCompany={
									onCreateCompanyContext
										? () =>
												void handleCreateContext(
													"company",
													onCreateCompanyContext,
												)
										: undefined
								}
							/>
						</>
					}
					confidenceSummary={
						showSingleContextConfirmation ? null : (
							<SelectorConfidenceSummary
								context={selectedContext}
								project={launchState.project}
							/>
						)
					}
					rememberControl={
						<RememberProjectControl
							binding={launchState.binding}
							contexts={launchState.contexts}
							rememberProject={rememberProject}
							selectedContextId={selectedContextId}
							disabled={launchPending}
							onRememberProjectChange={handleRememberProjectChange}
						/>
					}
					launchActions={
						<>
							{launchPending && showLaunchVerification ? (
								<div className="mb-3">
									<LaunchVerificationProgress
										projectName={launchState.project.name}
										contextName={selectedContext?.name ?? "selected context"}
										steps={
											launcherState.status === "launching"
												? launcherState.steps
												: undefined
										}
									/>
								</div>
							) : null}

							{launcherState.status === "failure" ? (
								<LaunchFailureView
									error={launcherState.error}
									onRetry={() => void handleLaunch()}
									onRunDiagnostics={onRunDiagnostics}
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
									onKeepCurrent={() =>
										void finishSuccessfulLaunch(launcherState.selection)
									}
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
										void handleLaunch({ allowExistingEnvironmentLaunch: true })
									}
								/>
							) : null}

							<SelectorActions
								launchDisabled={
									selectedContextId === undefined ||
									launchBlocked ||
									danglingBindingDialogOpen
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

interface LaunchAttemptOptions {
	confirmContextMismatch?: boolean;
	contextId?: string;
	allowExistingEnvironmentLaunch?: boolean;
	confirmIdentityMismatch?: boolean;
}

function selectedContextConfidenceBlocked(
	context: ContextState | undefined,
): boolean {
	return context?.confidence?.status === "blocked";
}

function initialLauncherSelection(launchState: LaunchState): LauncherSelection {
	return {
		selectedContextId: initialSelectedContextId(launchState),
		rovingContextId: initialRovingContextId(launchState),
		rememberProject: false,
	};
}

function initialLauncherState(launchState: LaunchState): LauncherState {
	const selection = initialLauncherSelection(launchState);
	if (launchState.binding.dangling) {
		return { status: "dangling_binding", selection, pending: false };
	}
	return selectingLauncherState(selection);
}

function unexpectedLaunchError(error: unknown): DisplayError {
	const message = error instanceof Error ? error.message : "Launch failed.";
	return {
		code: "unexpected_error",
		message,
		recovery: "Retry the launch. If it keeps failing, review diagnostics.",
	};
}

function unexpectedBindingError(error: unknown): DisplayError {
	const message =
		error instanceof Error ? error.message : "Could not remember this context.";
	return {
		code: "unexpected_error",
		message,
		recovery: "Try again or keep the current remembered context.",
	};
}

function unexpectedBindingRemovalError(error: unknown): DisplayError {
	const message =
		error instanceof Error
			? error.message
			: "Could not remove the remembered context.";
	return {
		code: "unexpected_error",
		message,
		recovery: "Try again or choose a context for this launch without removing it.",
	};
}

function MissingDefaultContextActions({
	launchState,
	providerCredentialSessions,
	providerSessionAssignments,
	pendingContextId,
	error,
	onClassifyProviderSession,
	onCreatePersonal,
	onCreateCompany,
}: {
	launchState: LaunchState;
	providerCredentialSessions: ProviderCredentialSession[];
	providerSessionAssignments: ProviderSessionAssignments;
	pendingContextId?: string;
	error?: DisplayError;
	onClassifyProviderSession: (
		providerId: string,
		contextId: "personal" | "company",
	) => void;
	onCreatePersonal?: () => void;
	onCreateCompany?: () => void;
}) {
	const missingDefaults = missingDefaultContextIds(launchState.contexts);
	const missingPersonal = missingDefaults.includes("personal");
	const missingCompany = missingDefaults.includes("company");
	const pending = pendingContextId !== undefined;
	const classificationComplete = providerCredentialSessions.every(
		(session) => providerSessionAssignments[session.providerId] !== undefined,
	);

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
					<div className="flex flex-wrap gap-2">
						{missingPersonal ? (
							<Button
								type="button"
								variant="outline"
								size="sm"
								disabled={
									pending ||
									!classificationComplete ||
									onCreatePersonal === undefined
								}
								onClick={onCreatePersonal}
							>
								{pendingContextId === "personal"
									? "Creating..."
									: "Add Personal"}
							</Button>
						) : null}
						{missingCompany ? (
							<Button
								type="button"
								variant="outline"
								size="sm"
								disabled={
									pending ||
									!classificationComplete ||
									onCreateCompany === undefined
								}
								onClick={onCreateCompany}
							>
								{pendingContextId === "company" ? "Creating..." : "Add Company"}
							</Button>
						) : null}
					</div>
				</div>
				<ProviderCredentialClassification
					sessions={providerCredentialSessions}
					assignments={providerSessionAssignments}
					disabled={pending}
					onClassify={onClassifyProviderSession}
				/>
				{pendingContextId ? (
					<p className="mt-3 text-sm text-muted-foreground" role="status">
						Creating {pendingContextId === "personal" ? "Personal" : "Company"}{" "}
						context...
					</p>
				) : null}
				{error ? <GuiErrorNotice error={error} /> : null}
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
