import { useEffect, useState } from "react";

import type {
	ApiResult,
	CreateContextResult,
	LaunchState,
} from "../../lib/devctx-api";
import { devContextApi } from "../../lib/devctx-api.js";
import { devContextWindow } from "../../lib/devctx-window.js";
import { notifyCodingToolLaunched } from "../notifications/notifications.js";
import { GuiErrorNotice } from "../selector/GuiErrorNotice.js";
import { createOnboardingContextAndRefresh } from "../selector/onboarding-action.js";
import { SelectorView } from "../selector/SelectorView.js";
import { LauncherSurface } from "./LauncherSurface.js";
import { ProjectNotFoundView } from "./ProjectNotFoundView.js";
import { ProjectResolvingView } from "./ProjectResolvingView.js";
import {
	type LoadState,
	loadStateFromResult,
} from "../app/load-state.js";

interface LauncherFlowProps {
	projectPath: string;
}

type ProjectLaunchState = LoadState<LaunchState> & { projectPath: string };

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
	const [choosingFolder, setChoosingFolder] = useState(false);
	const resolving =
		launchState.projectPath !== activeProjectPath ||
		launchState.status === "loading";

	useEffect(() => {
		setHostProjectPath(projectPath);
		setRequestedProjectPath(projectPath);
	}, [projectPath]);

	useEffect(() => {
		let active = true;
		setLaunchState({ projectPath: activeProjectPath, status: "loading" });
		void devContextApi.getLaunchState({ projectPath: activeProjectPath }).then((result) => {
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
	}, [activeProjectPath]);

	async function createContext(
		contextId: string,
		importProviderIds: string[],
	): Promise<ApiResult<CreateContextResult>> {
		const result = await createOnboardingContextAndRefresh({
			contextId,
			importProviderIds,
			createContext: (requestedContextId, requestedProviderIDs) =>
				devContextApi.createContext({
					contextId: requestedContextId,
					importProviderIds: requestedProviderIDs,
				}),
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
			) : (
				<SelectorView
					launchState={launchState.data}
					onBindProject={devContextApi.bindProject}
					onPreflightLaunchProject={devContextApi.preflightLaunchProject}
					onLaunchProject={devContextApi.launchProject}
					onCancel={devContextWindow.closeSelector}
					onCreatePersonalContext={(providerIds) =>
						createContext("personal", providerIds)
					}
					onCreateCompanyContext={(providerIds) =>
						createContext("company", providerIds)
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
		</LauncherSurface>
	);
}

export { LauncherFlow };
