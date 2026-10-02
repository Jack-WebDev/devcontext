import { useCallback, useRef, useState } from "react";

import type {
	ApiResult,
	ContextState,
	CreateContextRequest,
	CreateContextResult,
	DisplayError,
	LaunchState,
	ProjectState,
} from "../../lib/devctx-api";

type ContextCreationAction = (
	request: CreateContextRequest,
) => Promise<ApiResult<CreateContextResult>>;

const contextCreateSteps = ["identity", "projects", "tools", "review"] as const;

const initialContextCreationSteps: ContextCreationStep[] = [
	{ id: "create", label: "Create context", status: "pending" },
	{
		id: "initialize",
		label: "Initialize isolated tool storage",
		status: "pending",
	},
	{ id: "bind", label: "Save project associations", status: "pending" },
	{ id: "verify", label: "Verify context readiness", status: "pending" },
];

type ContextCreateStep = (typeof contextCreateSteps)[number];
type ContextCreateFlowStatus = ContextCreateStep | "creating" | "success";
type ContextCreationStepStatus =
	| "pending"
	| "running"
	| "complete"
	| "skipped"
	| "failed";

interface ContextCreateFlowState {
	status: ContextCreateFlowStatus;
	draft: CreateContextRequest;
	projects: ProjectState[];
}

interface ContextCreationStep {
	id: "create" | "bind" | "initialize" | "verify";
	label: string;
	status: ContextCreationStepStatus;
	detail?: string;
}

interface ContextCreationJourneyDependencies {
	createContext: ContextCreationAction;
	bindProject?: (request: {
		projectPath: string;
		contextId: string;
	}) => Promise<ApiResult<unknown>>;
	verifyContext?: (context: ContextState) => Promise<ApiResult<ContextState>>;
	initialProjects?: ProjectState[];
}

interface ContextCreationExecutionDependencies
	extends Omit<ContextCreationJourneyDependencies, "initialProjects"> {
	request: CreateContextRequest;
	projects: ProjectState[];
	createdContext?: ContextState;
	boundProjectPaths: string[];
	onStep: (
		id: ContextCreationStep["id"],
		status: ContextCreationStepStatus,
		detail?: string,
	) => void;
}

type ContextCreationExecutionResult =
	| {
			ok: true;
			context: ContextState;
			boundProjectPaths: string[];
	  }
	| {
			ok: false;
			error: DisplayError;
			context?: ContextState;
			boundProjectPaths: string[];
	  };

interface ContextCreationJourneyState {
	flow: ContextCreateFlowState;
	steps: ContextCreationStep[];
	created?: ContextState;
	error?: DisplayError;
	pending: boolean;
	updateDraft: (updates: Partial<CreateContextRequest>) => void;
	setDefaultEnabledDevelopmentToolIds: (ids: string[]) => void;
	updateProjects: (projects: ProjectState[]) => void;
	next: () => void;
	previous: () => void;
	edit: (section: Exclude<ContextCreateStep, "review">) => void;
	submit: () => Promise<void>;
	returnToReview: () => void;
	createAnother: () => void;
	recheck: () => Promise<void>;
}

function initialContextCreateFlow(
	draft: CreateContextRequest = {},
): ContextCreateFlowState {
	return { status: "identity", draft, projects: [] };
}

function updateContextCreateProjects(
	state: ContextCreateFlowState,
	projects: ProjectState[],
): ContextCreateFlowState {
	return { ...state, projects };
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

function editContextCreateSection(
	state: ContextCreateFlowState,
	section: Exclude<ContextCreateStep, "review">,
): ContextCreateFlowState {
	return state.status === "review" ? { ...state, status: section } : state;
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
	return contextCreateSteps.some((step) => step === status);
}

function updateContextCreationStep(
	steps: ContextCreationStep[],
	id: ContextCreationStep["id"],
	status: ContextCreationStepStatus,
	detail?: string,
): ContextCreationStep[] {
	return steps.map((step) =>
		step.id === id ? { ...step, status, detail } : step,
	);
}

async function executeContextCreation(
	dependencies: ContextCreationExecutionDependencies,
): Promise<ContextCreationExecutionResult> {
	let context = dependencies.createdContext;
	const boundProjectPaths = new Set(dependencies.boundProjectPaths);

	if (context === undefined) {
		dependencies.onStep("create", "running");
		const created = await dependencies.createContext(dependencies.request);
		if (!created.ok) {
			dependencies.onStep("create", "failed");
			return { ok: false, error: created.error, boundProjectPaths: [] };
		}
		context = created.data.context;
		dependencies.onStep("create", "complete");
		dependencies.onStep("initialize", "complete");
	} else {
		dependencies.onStep("create", "complete");
		dependencies.onStep("initialize", "complete");
	}

	const pendingProjects = dependencies.projects.filter(
		(project) => !boundProjectPaths.has(project.path),
	);
	if (
		dependencies.projects.length === 0 ||
		dependencies.bindProject === undefined
	) {
		dependencies.onStep("bind", "skipped", "No project associations selected.");
	} else if (pendingProjects.length === 0) {
		dependencies.onStep("bind", "complete", "Project associations are saved.");
	} else {
		dependencies.onStep("bind", "running");
		for (const [index, project] of pendingProjects.entries()) {
			const bound = await dependencies.bindProject({
				projectPath: project.path,
				contextId: context.id,
			});
			if (!bound.ok) {
				dependencies.onStep(
					"bind",
					"failed",
					`${index} of ${pendingProjects.length} project associations saved.`,
				);
				return {
					ok: false,
					error: bound.error,
					context,
					boundProjectPaths: [...boundProjectPaths],
				};
			}
			boundProjectPaths.add(project.path);
			dependencies.onStep(
				"bind",
				"running",
				`${index + 1} of ${pendingProjects.length} project associations saved.`,
			);
		}
		dependencies.onStep("bind", "complete");
	}

	dependencies.onStep("verify", "running");
	if (dependencies.verifyContext === undefined) {
		dependencies.onStep("verify", "complete");
		return { ok: true, context, boundProjectPaths: [...boundProjectPaths] };
	}
	const verified = await dependencies.verifyContext(context);
	if (!verified.ok) {
		dependencies.onStep("verify", "failed");
		return {
			ok: false,
			error: verified.error,
			context,
			boundProjectPaths: [...boundProjectPaths],
		};
	}
	dependencies.onStep("verify", "complete");
	return {
		ok: true,
		context: verified.data,
		boundProjectPaths: [...boundProjectPaths],
	};
}

interface ContextCreationState {
	pending: boolean;
	pendingRequest?: CreateContextRequest;
	error?: DisplayError;
	create: (
		request: CreateContextRequest,
	) => Promise<ApiResult<CreateContextResult> | undefined>;
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

function useContextCreationJourney(
	dependencies: ContextCreationJourneyDependencies,
): ContextCreationJourneyState {
	const [flow, setFlow] = useState<ContextCreateFlowState>(() => ({
		...initialContextCreateFlow(),
		projects: dependencies.initialProjects ?? [],
	}));
	const [steps, setSteps] = useState(initialContextCreationSteps);
	const [created, setCreated] = useState<ContextState>();
	const [error, setError] = useState<DisplayError>();
	const [boundProjectPaths, setBoundProjectPaths] = useState<string[]>([]);
	const requestInFlight = useRef(false);

	function updateDraft(updates: Partial<CreateContextRequest>) {
		setFlow((current) => updateContextCreateDraft(current, updates));
	}

	const setDefaultEnabledDevelopmentToolIds = useCallback((ids: string[]) => {
		setFlow((current) =>
			current.draft.enabledDevelopmentToolIds === undefined
				? updateContextCreateDraft(current, {
						enabledDevelopmentToolIds: ids,
					})
				: current,
		);
	}, []);

	function updateProjects(projects: ProjectState[]) {
		setFlow((current) => updateContextCreateProjects(current, projects));
	}

	function next() {
		setFlow(nextContextCreateStep);
	}

	function previous() {
		setFlow(previousContextCreateStep);
	}

	function edit(section: Exclude<ContextCreateStep, "review">) {
		setFlow((current) => editContextCreateSection(current, section));
	}

	async function submit() {
		if (
			requestInFlight.current ||
			(flow.status !== "review" && flow.status !== "creating")
		) {
			return;
		}

		requestInFlight.current = true;
		setFlow(beginContextCreation);
		setError(undefined);
		setSteps(initialContextCreationSteps);
		try {
			const result = await executeContextCreation({
				...dependencies,
				request: flow.draft,
				projects: flow.projects,
				createdContext: created,
				boundProjectPaths,
				onStep: (id, status, detail) =>
					setSteps((current) =>
						updateContextCreationStep(current, id, status, detail),
					),
			});
			setBoundProjectPaths(result.boundProjectPaths);
			if (result.context !== undefined) {
				setCreated(result.context);
			}
			if (!result.ok) {
				setError(result.error);
				return;
			}
			setFlow(completeContextCreation);
		} finally {
			requestInFlight.current = false;
		}
	}

	function returnToReview() {
		if (created === undefined) {
			setFlow(returnToContextCreateReview);
		}
	}

	function createAnother() {
		requestInFlight.current = false;
		setFlow(initialContextCreateFlow());
		setSteps(initialContextCreationSteps);
		setCreated(undefined);
		setError(undefined);
		setBoundProjectPaths([]);
	}

	async function recheck() {
		if (created === undefined || dependencies.verifyContext === undefined) {
			return;
		}
		const verified = await dependencies.verifyContext(created);
		if (verified.ok) {
			setCreated(verified.data);
		}
	}

	return {
		flow,
		steps,
		created,
		error,
		pending: requestInFlight.current,
		updateDraft,
		setDefaultEnabledDevelopmentToolIds,
		updateProjects,
		next,
		previous,
		edit,
		submit,
		returnToReview,
		createAnother,
		recheck,
	};
}

export type {
	ContextCreateFlowState,
	ContextCreateFlowStatus,
	ContextCreateStep,
	ContextCreationAction,
	ContextCreationExecutionDependencies,
	ContextCreationExecutionResult,
	ContextCreationJourneyDependencies,
	ContextCreationJourneyState,
	ContextCreationState,
	ContextCreationStep,
	ContextCreationStepStatus,
	CreateContextAndRefreshDependencies,
	CreateContextAndRefreshResult,
};
export {
	beginContextCreation,
	completeContextCreation,
	createContextAndRefresh,
	editContextCreateSection,
	executeContextCreation,
	initialContextCreateFlow,
	nextContextCreateStep,
	previousContextCreateStep,
	returnToContextCreateReview,
	updateContextCreateDraft,
	updateContextCreateProjects,
	useContextCreation,
	useContextCreationJourney,
};
