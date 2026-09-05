import { useRef, useState } from "react";

import type {
	ApiResult,
	CreateContextRequest,
	CreateContextResult,
	DisplayError,
	LaunchState,
} from "../../lib/devctx-api";

type ContextCreationAction = (
	request: CreateContextRequest,
) => Promise<ApiResult<CreateContextResult>>;

interface ContextCreationState {
	pending: boolean;
	pendingRequest?: CreateContextRequest;
	error?: DisplayError;
	create: (request: CreateContextRequest) => Promise<ApiResult<CreateContextResult> | undefined>;
	reset: () => void;
}

interface CreateContextAndRefreshDependencies {
	request: CreateContextRequest;
	createContext: ContextCreationAction;
	getLaunchState: () => Promise<ApiResult<LaunchState>>;
}

type CreateContextAndRefreshResult =
	| { ok: true; created: CreateContextResult; launchState: LaunchState }
	| { ok: false; error: DisplayError };

// Creates any context request, then reloads the launch state that owns the
// current user journey. The refresh prevents callers from duplicating the
// create-and-reconcile sequence.
async function createContextAndRefresh(
	dependencies: CreateContextAndRefreshDependencies,
): Promise<CreateContextAndRefreshResult> {
	const created = await dependencies.createContext(dependencies.request);
	if (!created.ok) {
		return { ok: false, error: created.error };
	}

	const refreshed = await dependencies.getLaunchState();
	if (!refreshed.ok) {
		return { ok: false, error: refreshed.error };
	}

	return { ok: true, created: created.data, launchState: refreshed.data };
}

// Keeps mutation state independent from the form that supplies a creation
// request, so the existing dialog and the later multi-step flow can share it.
function useContextCreation(
	createContext?: ContextCreationAction,
): ContextCreationState {
	const [pendingRequest, setPendingRequest] = useState<
		CreateContextRequest | undefined
	>(undefined);
	const [error, setError] = useState<DisplayError>();
	const requestInFlight = useRef(false);

	async function create(
		request: CreateContextRequest,
	): Promise<ApiResult<CreateContextResult> | undefined> {
		if (createContext === undefined || requestInFlight.current) {
			return undefined;
		}

		requestInFlight.current = true;
		setPendingRequest(request);
		setError(undefined);
		try {
			const result = await createContext(request);
			if (!result.ok) {
				setError(result.error);
			}
			return result;
		} finally {
			requestInFlight.current = false;
			setPendingRequest(undefined);
		}
	}

	function reset() {
		requestInFlight.current = false;
		setPendingRequest(undefined);
		setError(undefined);
	}

	return {
		pending: pendingRequest !== undefined,
		pendingRequest,
		error,
		create,
		reset,
	};
}

export type {
	ContextCreationAction,
	ContextCreationState,
	CreateContextAndRefreshDependencies,
	CreateContextAndRefreshResult,
};
export { createContextAndRefresh, useContextCreation };
