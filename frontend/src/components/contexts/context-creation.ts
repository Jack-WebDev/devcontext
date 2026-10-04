import { useCallback, useRef, useState } from "react";

import type {
	ApiResult,
	ContextState,
	CreateContextRequest,
	CreateContextResult,
	DisplayError,
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
	{ id: "refresh", label: "Refresh context state", status: "pending" },
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
	id: "create" | "bind" | "initialize" | "refresh";
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
	refreshContext?: (context: ContextState) => Promise<ApiResult<ContextState>>;
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

	dependencies.onStep("refresh", "running");
	if (dependencies.refreshContext === undefined) {
		dependencies.onStep("refresh", "complete");
		return { ok: true, context, boundProjectPaths: [...boundProjectPaths] };
	}
	const refreshed = await dependencies.refreshContext(context);
	if (!refreshed.ok) {
		dependencies.onStep("refresh", "failed");
		return {
			ok: false,
			error: refreshed.error,
			context,
			boundProjectPaths: [...boundProjectPaths],
		};
	}
	dependencies.onStep("refresh", "complete");
	return {
		ok: true,
		context: refreshed.data,
		boundProjectPaths: [...boundProjectPaths],
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
		if (created === undefined || dependencies.refreshContext === undefined) {
			return;
		}
		const refreshed = await dependencies.refreshContext(created);
		if (refreshed.ok) {
			setCreated(refreshed.data);
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
	ContextCreationStep,
	ContextCreationStepStatus,
};
export {
	beginContextCreation,
	completeContextCreation,
	editContextCreateSection,
	executeContextCreation,
	initialContextCreateFlow,
	nextContextCreateStep,
	previousContextCreateStep,
	returnToContextCreateReview,
	updateContextCreateDraft,
	updateContextCreateProjects,
	useContextCreationJourney,
};
