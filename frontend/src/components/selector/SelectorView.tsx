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
	RunningEnvironmentConflict,
} from "../../lib/devctx-api";
import {
	contextPositionFromShortcut,
	keyboardShortcuts,
} from "../command-palette/shortcut";
import { RunningEnvironmentConflictDialog } from "../running/RunningEnvironmentConflictDialog";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { AccountIdentityMismatchDialog } from "./AccountIdentityMismatchDialog";
import { hasAccountIdentityMismatch } from "./account-identity-mismatch";
import { ContextCard } from "./ContextCard";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
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
import { ProjectIdentity } from "./ProjectIdentity";
import {
	ProviderCredentialClassification,
	type ProviderSessionAssignments,
} from "./ProviderCredentialClassification.js";
import { RememberProjectControl } from "./RememberProjectControl";
import { recommendationReason } from "./recommendation";
import { SelectorActions } from "./SelectorActions";
import { SelectorConfidenceSummary } from "./SelectorConfidenceSummary";
import { SelectorLayout } from "./SelectorLayout";
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
	const [selectedContextId, setSelectedContextId] = useState<
		string | undefined
	>(() => initialSelectedContextId(launchState));
	const [rovingContextId, setRovingContextId] = useState<string | undefined>(
		() => initialRovingContextId(launchState),
	);
	const [rememberProject, setRememberProject] = useState(false);
	const [launchLifecycle, setLaunchLifecycle] = useState<LaunchLifecycleState>({
		status: "idle",
	});
	const [launchError, setLaunchError] = useState<DisplayError | undefined>(
		undefined,
	);
	const [mismatchError, setMismatchError] = useState<DisplayError | undefined>(
		undefined,
	);
	const [identityMismatchContextID, setIdentityMismatchContextID] = useState<
		string | undefined
	>(undefined);
	const [runningEnvironmentConflict, setRunningEnvironmentConflict] = useState<
		RunningEnvironmentConflict | undefined
	>(undefined);
	const [onboardingPendingContextId, setOnboardingPendingContextId] = useState<
		string | undefined
	>(undefined);
	const [onboardingError, setOnboardingError] = useState<
		DisplayError | undefined
	>(undefined);
	const [providerSessionAssignments, setProviderSessionAssignments] =
		useState<ProviderSessionAssignments>({});
	const contextButtonRefs = useRef(new Map<string, HTMLButtonElement>());
	const launchGuard = useRef(createLaunchRequestGuard());
	const mismatchDialogOpen =
		mismatchError?.contextMismatch !== undefined ||
		identityMismatchContextID !== undefined;
	const selectedContext = launchState.contexts.find(
		(context) => context.id === selectedContextId,
	);
	const launchPending = launchLifecycle.status !== "idle";
	const launchBlocked = selectedContextConfidenceBlocked(selectedContext);
	const keyboardLaunchAvailable = canLaunchSelectedContextFromKeyboard({
		selectedContextId,
		launchPending: launchPending || launchBlocked,
		mismatchDialogOpen,
	});

	useEffect(() => {
		setSelectedContextId(initialSelectedContextId(launchState));
		setRovingContextId(initialRovingContextId(launchState));
		setRememberProject(false);
		setLaunchLifecycle({ status: "idle" });
		setLaunchError(undefined);
		setMismatchError(undefined);
		setIdentityMismatchContextID(undefined);
		setOnboardingPendingContextId(undefined);
		setOnboardingError(undefined);
		setProviderSessionAssignments({});
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
		const nextContextId = nextSelectedContextId(
			launchState.contexts,
			contextId,
		);
		setSelectedContextId(nextContextId);
		setRovingContextId(nextContextId);
		setIdentityMismatchContextID(undefined);
	}

	function handleContextNavigation(
		contextId: string,
		direction: ContextNavigationDirection,
	) {
		if (launchPending) {
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

		setSelectedContextId(nextContextId);
		setRovingContextId(nextContextId);
		contextButtonRefs.current.get(nextContextId)?.focus();
	}

	function handleSelectorKeyDown(event: KeyboardEvent<HTMLDivElement>) {
		const contextPosition = contextPositionFromShortcut(event);
		if (
			contextPosition !== undefined &&
			!launchPending &&
			!mismatchDialogOpen
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
			launchPending,
			mismatchDialogOpen,
		});
		if (action === "none") {
			return;
		}

		event.preventDefault();
		event.stopPropagation();

		if (action === "close-dialog") {
			setMismatchError(undefined);
			setIdentityMismatchContextID(undefined);
			return;
		}

		void cancelSelector({ closeSelector: onCancel });
	}

	async function handleLaunch({
		confirmContextMismatch = false,
		contextId = selectedContextId,
		allowExistingEnvironmentLaunch = false,
		confirmIdentityMismatch = false,
	}: LaunchAttemptOptions = {}) {
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
			setIdentityMismatchContextID(contextId);
			return;
		}

		await launchGuard.current.run(async () => {
			setLaunchLifecycle({ status: "preflighting" });
			setLaunchError(undefined);
			if (confirmContextMismatch) {
				setMismatchError(undefined);
			}
			if (confirmIdentityMismatch) {
				setIdentityMismatchContextID(undefined);
			}

			try {
				const result = await launchSelectedContext({
					projectPath: launchState.project.path,
					selectedContextId: contextId,
					rememberProject,
					confirmContextMismatch,
					allowExistingEnvironmentLaunch,
					onPreflightComplete: (preflight) => {
						setLaunchLifecycle({
							status: "launching",
							steps: preflight.verificationSteps,
						});
					},
					bindProject: onBindProject,
					preflightLaunchProject: onPreflightLaunchProject,
					launchProject: onLaunchProject,
				});

				if (result && "runningEnvironmentConflict" in result) {
					setRunningEnvironmentConflict(result.runningEnvironmentConflict);
				} else if (result?.ok) {
					if ("project" in result.data && "context" in result.data) {
						onCodingToolLaunched?.(result.data);
					}
					if (shouldCloseSelectorAfterLaunch(launchSuccessCloseBehavior)) {
						await cancelSelector({ closeSelector: onCancel });
					}
				} else if (result && !result.ok) {
					if (
						result.error.code === "context_mismatch_requires_confirmation" &&
						result.error.contextMismatch
					) {
						setMismatchError(result.error);
					} else {
						setLaunchError(result.error);
					}
				}
			} finally {
				setLaunchLifecycle({ status: "idle" });
			}
		});
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
		<div className="space-y-8" onKeyDown={handleSelectorKeyDown}>
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
							{launchState.contexts.length === 0 ? (
								<SelectorEmptyContextState />
							) : (
								<div className="grid gap-4 sm:grid-cols-2">
									{launchState.contexts.map((context) => (
										<ContextCard
											key={context.id}
											context={context}
											selected={selectedContextId === context.id}
											recommendation={recommendationForContext(
												launchState,
												context.id,
											)}
											disabled={launchPending}
											tabIndex={rovingContextId === context.id ? 0 : -1}
											buttonRef={setContextButtonRef(context.id)}
											onSelect={handleSelectContext}
											onNavigate={handleContextNavigation}
											onLaunchSelected={
												keyboardLaunchAvailable
													? () => void handleLaunch()
													: undefined
											}
											onProviderSetup={(contextId) => {
												handleSelectContext(contextId);
												void handleLaunch({ contextId });
											}}
										/>
									))}
								</div>
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
						<SelectorConfidenceSummary
							context={selectedContext}
							project={launchState.project}
						/>
					}
					rememberControl={
						<RememberProjectControl
							binding={launchState.binding}
							contexts={launchState.contexts}
							rememberProject={rememberProject}
							selectedContextId={selectedContextId}
							disabled={launchPending}
							onRememberProjectChange={setRememberProject}
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
											launchLifecycle.status === "launching"
												? launchLifecycle.steps
												: undefined
										}
									/>
								</div>
							) : null}

							{launchError ? (
								<LaunchFailureView
									error={launchError}
									onRetry={() => void handleLaunch()}
									onRunDiagnostics={onRunDiagnostics}
									onCancel={() =>
										void cancelSelector({ closeSelector: onCancel })
									}
								/>
							) : null}

							{mismatchError?.contextMismatch ? (
								<ContextMismatchDialog
									mismatch={mismatchError.contextMismatch}
									contexts={launchState.contexts}
									launchPending={launchPending}
									onCancel={() => setMismatchError(undefined)}
									onOpenAnyway={() => {
										const mismatch = mismatchError.contextMismatch;
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

							{identityMismatchContextID !== undefined ? (
								<AccountIdentityMismatchDialog
									contextName={
										launchState.contexts.find(
											(context) => context.id === identityMismatchContextID,
										)?.name ?? "selected context"
									}
									launchPending={launchPending}
									onCancel={() => setIdentityMismatchContextID(undefined)}
									onReviewConfiguration={() => {
										setIdentityMismatchContextID(undefined);
										onRunDiagnostics?.();
									}}
									onLaunchAnyway={() =>
										void handleLaunch({
											contextId: identityMismatchContextID,
											confirmIdentityMismatch: true,
										})
									}
								/>
							) : null}

							{runningEnvironmentConflict ? (
								<RunningEnvironmentConflictDialog
									conflict={runningEnvironmentConflict}
									launchPending={launchPending}
									onCancel={() => setRunningEnvironmentConflict(undefined)}
									onLaunchAnother={() =>
										void handleLaunch({ allowExistingEnvironmentLaunch: true })
									}
								/>
							) : null}

							<SelectorActions
								launchDisabled={
									selectedContextId === undefined || launchBlocked
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
		</div>
	);
}

type LaunchLifecycleState =
	| { status: "idle" }
	| { status: "preflighting" }
	| {
			status: "launching";
			steps?: PreflightLaunchProjectResult["verificationSteps"];
	  };

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

function SelectorEmptyContextState() {
	return (
		<Card
			as="section"
			size="sm"
			className="border border-border bg-muted/30 p-5"
			aria-labelledby="empty-context-title"
		>
			<h3 id="empty-context-title" className="font-medium">
				Create a development context
			</h3>
			<p className="mt-2 text-sm text-muted-foreground">
				Contexts keep provider accounts and coding-tool storage separate for the
				projects you open.
			</p>
			<ul className="mt-4 space-y-2 text-sm text-muted-foreground">
				<li>
					<span className="font-medium text-foreground">Personal:</span>{" "}
					personal projects and accounts.
				</li>
				<li>
					<span className="font-medium text-foreground">Company:</span> work or
					client projects and accounts.
				</li>
				<li>
					<span className="font-medium text-foreground">Custom:</span> create a
					tailored context from Contexts when custom creation is available.
				</li>
			</ul>
		</Card>
	);
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

function recommendationForContext(
	launchState: LaunchState,
	contextId: string,
): string | undefined {
	if (launchState.selectedContextId !== contextId) {
		return undefined;
	}

	return recommendationReason(launchState.resolutionSource);
}

export { SelectorView };
