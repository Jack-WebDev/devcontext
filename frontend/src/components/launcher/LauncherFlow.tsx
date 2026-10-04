import { useEffect, useState } from "react";

import type {
	DisplayError,
	LaunchState,
	RunningEnvironmentConflict,
	SettingsState,
} from "../../lib/devctx-api";
import { devContextApi } from "../../lib/devctx-api.js";
import { devContextWindow } from "../../lib/devctx-window.js";
import { type LoadState, loadStateFromResult } from "../app/load-state.js";
import { CreateContextDialog } from "../contexts/ContextManagement.js";
import { notifyCodingToolLaunched } from "../notifications/notifications.js";
import {
	continueProjectLaunchJourney,
	type ProjectLaunchAdapter,
	type ProjectLaunchJourneyResult,
	type ProjectLaunchPending,
	runProjectLaunchJourney,
} from "../project-launch/project-launch-journey.js";
import { RunningEnvironmentConflictDialog } from "../running/RunningEnvironmentConflictDialog.js";
import { GuiErrorNotice } from "../selector/GuiErrorNotice.js";
import { PreflightReviewDialog } from "../selector/PreflightReviewDialog.js";
import { SelectorView } from "../selector/SelectorView.js";
import { LauncherSurface } from "./LauncherSurface.js";
import { ProjectNotFoundView } from "./ProjectNotFoundView.js";
import { ProjectResolvingView } from "./ProjectResolvingView.js";

interface LauncherFlowProps {
	projectPath: string;
	onCancel?: () => void;
	onRunDiagnostics?: () => void;
}

type ProjectLaunchState = LoadState<LaunchState> & { projectPath: string };
type LauncherSettingsState = LoadState<SettingsState>;

interface PendingRunningEnvironmentLaunch {
	conflict: RunningEnvironmentConflict;
	pending: ProjectLaunchPending;
}

const desktopProjectLaunchAdapter: ProjectLaunchAdapter = {
	preflightLaunchProject: devContextApi.preflightLaunchProject,
	launchProject: devContextApi.launchProject,
};

// LauncherFlow is intentionally separate from the management shell. Later
// launcher phases add resolution and selection states inside this focused
// surface without bringing management navigation into a project launch.
function LauncherFlow({
	projectPath,
	onCancel,
	onRunDiagnostics,
}: LauncherFlowProps) {
	const cancel = onCancel ?? (() => void devContextWindow.closeSelector());
	const [requestedProjectPath, setRequestedProjectPath] = useState(projectPath);
	const [hostProjectPath, setHostProjectPath] = useState(projectPath);
	const projectPathChangedByHost = hostProjectPath !== projectPath;
	const activeProjectPath = projectPathChangedByHost
		? projectPath
		: requestedProjectPath;
	const [launchState, setLaunchState] = useState<ProjectLaunchState>({
		projectPath: activeProjectPath,
		status: "loading",
	});
	const [settings, setSettings] = useState<LauncherSettingsState>({
		status: "loading",
	});
	const [choosingFolder, setChoosingFolder] = useState(false);
	const [detectionRetry, setDetectionRetry] = useState(0);
	const [creatingFirstContext, setCreatingFirstContext] = useState(false);
	const [createdContextLaunchPending, setCreatedContextLaunchPending] =
		useState(false);
	const [createdContextLaunchError, setCreatedContextLaunchError] = useState<
		DisplayError | undefined
	>();
	const [pendingPreflightReview, setPendingPreflightReview] = useState<
		ProjectLaunchPending | undefined
	>();
	const [pendingRunningEnvironmentLaunch, setPendingRunningEnvironmentLaunch] =
		useState<PendingRunningEnvironmentLaunch | undefined>();
	const resolving =
		launchState.projectPath !== activeProjectPath ||
		launchState.status === "loading" ||
		settings.status === "loading";

	useEffect(() => {
		setHostProjectPath(projectPath);
		setRequestedProjectPath(projectPath);
	}, [projectPath]);

	useEffect(() => {
		// This counter intentionally re-runs detection without changing the request.
		void detectionRetry;
		let active = true;
		setLaunchState({ projectPath: activeProjectPath, status: "loading" });
		void devContextApi
			.getLaunchState({ projectPath: activeProjectPath })
			.then((result) => {
				if (active) {
					setLaunchState({
						projectPath: activeProjectPath,
						...loadStateFromResult(result),
					});
				}
			});
		return () => {
			active = false;
		};
	}, [activeProjectPath, detectionRetry]);

	useEffect(() => {
		let active = true;
		void devContextApi.getSettings().then((result) => {
			if (active) {
				setSettings(loadStateFromResult(result));
			}
		});
		return () => {
			active = false;
		};
	}, []);

	async function chooseProjectFolder() {
		setChoosingFolder(true);
		try {
			const result = await devContextApi.chooseProjectDirectory();
			if (result.ok && result.data !== undefined) {
				setRequestedProjectPath(result.data);
			}
		} finally {
			setChoosingFolder(false);
		}
	}

	async function handleCreatedContextLaunchOutcome(
		result: ProjectLaunchJourneyResult,
	) {
		if (result.kind === "preflight-review") {
			setPendingPreflightReview(result.pending);
			return;
		}
		if (result.kind === "running-environment-conflict") {
			setPendingRunningEnvironmentLaunch({
				conflict: result.conflict,
				pending: result.pending,
			});
			return;
		}
		if (result.kind === "failed") {
			setCreatedContextLaunchError(result.error);
			return;
		}

		notifyCodingToolLaunched({
			projectName: result.result.project.name,
			contextName: result.result.context.name,
			toolName: result.result.context.tool.name,
		});
		setCreatingFirstContext(false);
	}

	async function launchCreatedContext(contextId: string) {
		if (createdContextLaunchPending) {
			return;
		}
		setCreatedContextLaunchPending(true);
		setCreatedContextLaunchError(undefined);
		try {
			await handleCreatedContextLaunchOutcome(
				await runProjectLaunchJourney({
					request: { projectPath: activeProjectPath, contextId },
					...desktopProjectLaunchAdapter,
				}),
			);
		} finally {
			setCreatedContextLaunchPending(false);
		}
	}

	async function continueCreatedContextLaunch(
		pending: ProjectLaunchPending,
		decision: "continue-after-review" | "launch-another",
	) {
		if (createdContextLaunchPending) {
			return;
		}
		setCreatedContextLaunchPending(true);
		setCreatedContextLaunchError(undefined);
		try {
			await handleCreatedContextLaunchOutcome(
				await continueProjectLaunchJourney({
					pending,
					decision,
					...desktopProjectLaunchAdapter,
				}),
			);
		} finally {
			setCreatedContextLaunchPending(false);
		}
	}

	const projectPathError =
		!resolving &&
		launchState.status === "error" &&
		launchState.error.projectPathIssue !== undefined;

	return (
		<LauncherSurface projectPath={activeProjectPath}>
			{resolving ? (
				<ProjectResolvingView />
			) : projectPathError ? (
				<ProjectNotFoundView
					choosingFolder={choosingFolder}
					onChooseFolder={() => void chooseProjectFolder()}
					onCancel={cancel}
				/>
			) : launchState.status === "error" ? (
				<GuiErrorNotice error={launchState.error} />
			) : settings.status === "error" ? (
				<GuiErrorNotice error={settings.error} />
			) : (
				<SelectorView
					launchState={launchState.data}
					onBindProject={devContextApi.bindProject}
					onUnbindProject={devContextApi.unbindProject}
					onPreflightLaunchProject={devContextApi.preflightLaunchProject}
					onLaunchProject={devContextApi.launchProject}
					onCancel={cancel}
					onRunDiagnostics={onRunDiagnostics}
					onStartContextCreation={() => setCreatingFirstContext(true)}
					onRetryDetection={() => setDetectionRetry((attempt) => attempt + 1)}
					launchSuccessCloseBehavior={
						settings.data.closeAfterLaunch ? "close_selector" : "keep_open"
					}
					showLaunchVerification={settings.data.launchVerification}
					projectMemoryEnabled={settings.data.rememberProjects}
					requireContextMismatchConfirmation={
						settings.data.warnOnContextMismatch
					}
					onCodingToolLaunched={(result) =>
						notifyCodingToolLaunched({
							projectName: result.project.name,
							contextName: result.context.name,
							toolName: result.context.tool.name,
						})
					}
				/>
			)}
			{creatingFirstContext && launchState.status === "loaded" ? (
				<CreateContextDialog
					contexts={[]}
					initialProjects={[launchState.data.project]}
					projectName={launchState.data.project.name}
					launchPending={createdContextLaunchPending}
					launchError={createdContextLaunchError}
					onClose={() => setCreatingFirstContext(false)}
					create={devContextApi.createContext}
					loadCreationOptions={devContextApi.getContextTemplates}
					bindProject={devContextApi.bindProject}
					refreshContext={async (context) => {
						const result = await devContextApi.getLaunchState({
							projectPath: activeProjectPath,
						});
						if (result.ok) {
							setLaunchState({
								projectPath: activeProjectPath,
								status: "loaded",
								data: result.data,
							});
							const verifiedContext = result.data.contexts.find(
								(candidate) => candidate.id === context.id,
							);
							if (!verifiedContext) {
								return {
									ok: false,
									error: {
										code: "internal_error",
										message: "The new context could not be verified.",
										recovery: "Try creating the context again.",
									},
								};
							}
							return { ok: true, data: verifiedContext };
						}
						return result;
					}}
					onOpenProject={(context) => void launchCreatedContext(context.id)}
				/>
			) : null}
			{pendingPreflightReview ? (
				<PreflightReviewDialog
					preflight={pendingPreflightReview.preflight}
					pending={createdContextLaunchPending}
					error={createdContextLaunchError}
					onCancel={() => {
						if (!createdContextLaunchPending) {
							setPendingPreflightReview(undefined);
							setCreatedContextLaunchError(undefined);
						}
					}}
					onContinue={() =>
						void continueCreatedContextLaunch(
							pendingPreflightReview,
							"continue-after-review",
						)
					}
				/>
			) : null}
			{pendingRunningEnvironmentLaunch ? (
				<RunningEnvironmentConflictDialog
					conflict={pendingRunningEnvironmentLaunch.conflict}
					launchPending={createdContextLaunchPending}
					error={createdContextLaunchError}
					onCancel={() => {
						if (!createdContextLaunchPending) {
							setPendingRunningEnvironmentLaunch(undefined);
							setCreatedContextLaunchError(undefined);
						}
					}}
					onLaunchAnother={() =>
						void continueCreatedContextLaunch(
							pendingRunningEnvironmentLaunch.pending,
							"launch-another",
						)
					}
				/>
			) : null}
		</LauncherSurface>
	);
}

export { LauncherFlow };
