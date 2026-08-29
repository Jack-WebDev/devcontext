import type { DisplayError } from "../../lib/devctx-api";

export type LoadState<T> =
	| { status: "loading" }
	| { status: "loaded"; data: T }
	| { status: "error"; error: DisplayError };

export function loadStateFromResult<T>(
	result: { ok: true; data: T } | { ok: false; error: DisplayError },
): LoadState<T> {
	return result.ok
		? { status: "loaded", data: result.data }
		: { status: "error", error: result.error };
}
