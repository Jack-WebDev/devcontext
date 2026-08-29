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
import {
	type LoadState,
	loadStateFromResult,
} from "../app/load-state.js";

interface LauncherFlowProps {
	projectPath: string;
}

// LauncherFlow is intentionally separate from the management shell. Later
// launcher phases add resolution and selection states inside this focused
// surface without bringing management navigation into a project launch.
function LauncherFlow({ projectPath }: LauncherFlowProps) {
	const [launchState, setLaunchState] = useState<LoadState<LaunchState>>({
		status: "loading",
	});

	useEffect(() => {
		let active = true;
		void devContextApi.getLaunchState({ projectPath }).then((result) => {
			if (active) {
				setLaunchState(loadStateFromResult(result));
			}
		});
		return () => {
			active = false;
		};
	}, [projectPath]);

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
			getLaunchState: () => devContextApi.getLaunchState({ projectPath }),
		});
		if (!result.ok) {
			return { ok: false, error: result.error };
		}

		setLaunchState({ status: "loaded", data: result.launchState });
		return { ok: true, data: result.created };
	}

	return (
		<LauncherSurface projectPath={projectPath}>
			{launchState.status === "loading" ? (
				<p className="text-sm text-muted-foreground" role="status">
					Loading launch options...
				</p>
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
