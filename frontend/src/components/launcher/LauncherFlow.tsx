import { useEffect, useState } from "react";

import type {
	ApiResult,
	CreateContextRequest,
	CreateContextResult,
	LaunchState,
	SettingsState,
} from "../../lib/devctx-api";
import { devContextApi } from "../../lib/devctx-api.js";
import { devContextWindow } from "../../lib/devctx-window.js";
import { notifyCodingToolLaunched } from "../notifications/notifications.js";
import { GuiErrorNotice } from "../selector/GuiErrorNotice.js";
import { createContextAndRefresh } from "../contexts/context-creation.js";
import { CreateContextDialog } from "../contexts/ContextManagement.js";
import { SelectorView } from "../selector/SelectorView.js";
import { LauncherSurface } from "./LauncherSurface.js";
import { ProjectNotFoundView } from "./ProjectNotFoundView.js";
import { ProjectResolvingView } from "./ProjectResolvingView.js";
import { type LoadState, loadStateFromResult } from "../app/load-state.js";

interface LauncherFlowProps {
	projectPath: string;
}

type ProjectLaunchState = LoadState<LaunchState> & { projectPath: string };
type LauncherSettingsState = LoadState<SettingsState>;

// LauncherFlow is intentionally separate from the management shell. Later
// launcher phases add resolution and selection states inside this focused
// surface without bringing management navigation into a project launch.
function LauncherFlow({ projectPath }: LauncherFlowProps) {
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
	const resolving =
		launchState.projectPath !== activeProjectPath ||
		launchState.status === "loading" ||
		settings.status === "loading";

	useEffect(() => {
		setHostProjectPath(projectPath);
		setRequestedProjectPath(projectPath);
	}, [projectPath]);

	useEffect(() => {
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

	async function createContext(
		request: CreateContextRequest,
	): Promise<ApiResult<CreateContextResult>> {
		const result = await createContextAndRefresh({
			request,
			createContext: devContextApi.createContext,
			getLaunchState: () =>
				devContextApi.getLaunchState({ projectPath: activeProjectPath }),
		});
		if (!result.ok) {
			return { ok: false, error: result.error };
		}

		setLaunchState({
			projectPath: activeProjectPath,
			status: "loaded",
			data: result.launchState,
		});
		return { ok: true, data: result.created };
	}

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
					onCancel={() => void devContextWindow.closeSelector()}
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
					onCancel={devContextWindow.closeSelector}
					onCreateContext={createContext}
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
					onClose={() => setCreatingFirstContext(false)}
					create={devContextApi.createContext}
					loadCreationOptions={devContextApi.getContextTemplates}
					bindProject={devContextApi.bindProject}
					verifyContext={async (context) => {
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
					onOpenProject={async (context) => {
						const launched = await devContextApi.launchProject({
							projectPath: activeProjectPath,
							contextId: context.id,
						});
						if (!launched.ok) return;
						notifyCodingToolLaunched({
							projectName: launched.data.project.name,
							contextName: launched.data.context.name,
							toolName: launched.data.context.tool.name,
						});
						setCreatingFirstContext(false);
					}}
				/>
			) : null}
		</LauncherSurface>
	);
}

export { LauncherFlow };
