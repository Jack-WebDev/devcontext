import type {
	DisplayError,
	PreflightLaunchProjectResult,
	RunningEnvironmentConflict,
} from "../../lib/devctx-api";

interface LauncherSelection {
	selectedContextId?: string;
	rovingContextId?: string;
	rememberProject: boolean;
}

interface LaunchAttempt {
	confirmContextMismatch: boolean;
	contextId?: string;
	allowExistingEnvironmentLaunch: boolean;
	confirmIdentityMismatch: boolean;
}

type LauncherState =
	| { status: "resolving" }
	| { status: "selecting"; selection: LauncherSelection }
	| {
			status: "identity_mismatch";
			selection: LauncherSelection;
			contextId: string;
	  }
	| {
			status: "context_mismatch";
			selection: LauncherSelection;
			error: DisplayError;
	  }
	| { status: "preflighting"; selection: LauncherSelection }
	| {
			status: "preflight_review";
			selection: LauncherSelection;
			preflight: PreflightLaunchProjectResult;
			attempt: LaunchAttempt;
	  }
	| {
			status: "launching";
			selection: LauncherSelection;
			groups?: PreflightLaunchProjectResult["groups"];
			steps?: PreflightLaunchProjectResult["verificationSteps"];
	  }
	| {
			status: "existing_workspace";
			selection: LauncherSelection;
			conflict: RunningEnvironmentConflict;
	  }
	| {
			status: "binding_replacement";
			selection: LauncherSelection;
			boundContextId: string;
			replacementContextId: string;
			pending: boolean;
			error?: DisplayError;
	  }
	| {
			status: "dangling_binding";
			selection: LauncherSelection;
			pending: boolean;
			error?: DisplayError;
	  }
	| {
			status: "failure";
			selection: LauncherSelection;
			error: DisplayError;
	  };

function selectingLauncherState(
	selection: LauncherSelection,
): LauncherState {
	return { status: "selecting", selection };
}

function launcherSelection(state: LauncherState): LauncherSelection | undefined {
	return "selection" in state ? state.selection : undefined;
}

function launcherStateIsPending(state: LauncherState): boolean {
	return (
		state.status === "preflighting" ||
		state.status === "launching" ||
		(state.status === "binding_replacement" && state.pending) ||
		(state.status === "dangling_binding" && state.pending)
	);
}

export type { LaunchAttempt, LauncherSelection, LauncherState };
export { launcherSelection, launcherStateIsPending, selectingLauncherState };
