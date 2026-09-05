import type { CreateContextRequest } from "../../lib/devctx-api";

const contextCreateSteps = [
	"identity",
	"projects",
	"tools",
	"review",
] as const;

type ContextCreateStep = (typeof contextCreateSteps)[number];
type ContextCreateFlowStatus = ContextCreateStep | "creating" | "success";

interface ContextCreateFlowState {
	status: ContextCreateFlowStatus;
	draft: CreateContextRequest;
}

function initialContextCreateFlow(
	draft: CreateContextRequest = {},
): ContextCreateFlowState {
	return { status: "identity", draft };
}

function updateContextCreateDraft(
	state: ContextCreateFlowState,
	updates: Partial<CreateContextRequest>,
): ContextCreateFlowState {
	return { ...state, draft: { ...state.draft, ...updates } };
}

function nextContextCreateStep(
	state: ContextCreateFlowState,
): ContextCreateFlowState {
	if (!isContextCreateStep(state.status)) {
		return state;
	}

	const index = contextCreateSteps.indexOf(state.status);
	const next = contextCreateSteps[index + 1];
	return next === undefined ? state : { ...state, status: next };
}

function previousContextCreateStep(
	state: ContextCreateFlowState,
): ContextCreateFlowState {
	if (!isContextCreateStep(state.status)) {
		return state;
	}

	const index = contextCreateSteps.indexOf(state.status);
	const previous = contextCreateSteps[index - 1];
	return previous === undefined ? state : { ...state, status: previous };
}

function beginContextCreation(
	state: ContextCreateFlowState,
): ContextCreateFlowState {
	return state.status === "review" ? { ...state, status: "creating" } : state;
}

function completeContextCreation(
	state: ContextCreateFlowState,
): ContextCreateFlowState {
	return state.status === "creating" ? { ...state, status: "success" } : state;
}

function returnToContextCreateReview(
	state: ContextCreateFlowState,
): ContextCreateFlowState {
	return state.status === "creating" ? { ...state, status: "review" } : state;
}

function isContextCreateStep(
	status: ContextCreateFlowStatus,
): status is ContextCreateStep {
	return contextCreateSteps.includes(status as ContextCreateStep);
}

export type { ContextCreateFlowState, ContextCreateFlowStatus, ContextCreateStep };
export {
	beginContextCreation,
	completeContextCreation,
	initialContextCreateFlow,
	nextContextCreateStep,
	previousContextCreateStep,
	returnToContextCreateReview,
	updateContextCreateDraft,
};
