import { useEffect, useState } from "react";
import type {
	ApiResult,
	ContextListItem,
	ContextState,
	CreateContextRequest,
	CreateContextResult,
	DevelopmentToolIntegration,
	ProjectState,
} from "../../lib/devctx-api";
import {
	Sheet,
	SheetContent,
	SheetDescription,
	SheetHeader,
	SheetTitle,
} from "../ui/sheet.js";
import {
	ContextCreationProgress,
	ContextCreateSuccessScreen,
	type CreationStep,
} from "./ContextCreateCompletion";
import {
	editContextCreateSection,
	initialContextCreateFlow,
	nextContextCreateStep,
	previousContextCreateStep,
	updateContextCreateDraft,
	updateContextCreateProjects,
} from "./context-create-flow";
import { ContextCreateDevelopmentToolsScreen } from "./ContextCreateDevelopmentToolsScreen";
import { ContextCreateIdentityScreen } from "./ContextCreateIdentityScreen";
import { ContextCreateProjectsScreen } from "./ContextCreateProjectsScreen";
import { ContextCreateReviewScreen } from "./ContextCreateReviewScreen";

export { ContextDetailsDrawer } from "./ContextDetailsDrawer";

const initialSteps: CreationStep[] = [
	{ id: "create", label: "Create context", status: "pending" },
	{
		id: "initialize",
		label: "Initialize isolated tool storage",
		status: "pending",
	},
	{ id: "bind", label: "Save project associations", status: "pending" },
	{ id: "verify", label: "Verify context readiness", status: "pending" },
];

export function CreateContextDialog({
	contexts,
	onClose,
	create,
	loadCreationOptions,
	bindProject,
	verifyContext,
	initialProjects = [],
	projectName,
	onOpenProject,
	onViewContext,
}: {
	contexts: ContextListItem[];
	onClose: () => void;
	create: (
		request: CreateContextRequest,
	) => Promise<ApiResult<CreateContextResult>>;
	loadCreationOptions?: () => Promise<
		ApiResult<{ developmentTools: DevelopmentToolIntegration[] }>
	>;
	bindProject?: (request: {
		projectPath: string;
		contextId: string;
	}) => Promise<ApiResult<unknown>>;
	verifyContext?: (context: ContextState) => Promise<ApiResult<ContextState>>;
	initialProjects?: ProjectState[];
	projectName?: string;
	onOpenProject?: (context: ContextState) => void;
	onViewContext?: (contextId: string) => void;
}) {
	const [flow, setFlow] = useState(() => ({
		...initialContextCreateFlow(),
		projects: initialProjects,
	}));
	const [steps, setSteps] = useState(initialSteps);
	const [error, setError] = useState<string>();
	const [created, setCreated] = useState<ContextState>();
	const [catalogIntegrations, setCatalogIntegrations] = useState<
		DevelopmentToolIntegration[]
	>();
	const [catalogLoading, setCatalogLoading] = useState(
		loadCreationOptions !== undefined,
	);
	const [catalogError, setCatalogError] = useState<string>();
	const integrations = catalogIntegrations ?? uniqueIntegrations(contexts);

	useEffect(() => {
		let active = true;
		if (!loadCreationOptions) {
			return () => {
				active = false;
			};
		}
		void loadCreationOptions().then((result) => {
			if (!active) return;
			setCatalogLoading(false);
			if (!result.ok) {
				setCatalogError(result.error.message);
				return;
			}
			setCatalogIntegrations(result.data.developmentTools);
			setFlow((current) => {
				if (current.draft.enabledDevelopmentToolIds !== undefined) {
					return current;
				}
				return updateContextCreateDraft(current, {
					enabledDevelopmentToolIds: result.data.developmentTools
						.filter((integration) => integration.enabled)
						.map((integration) => integration.id),
				});
			});
		});
		return () => {
			active = false;
		};
	}, [loadCreationOptions]);

	function updateStep(
		id: CreationStep["id"],
		status: CreationStep["status"],
		detail?: string,
	) {
		setSteps((current) =>
			current.map((step) =>
				step.id === id ? { ...step, status, detail } : step,
			),
		);
	}

	async function submit() {
		setFlow((current) => ({ ...current, status: "creating" }));
		setError(undefined);
		setSteps(initialSteps);
		let context = created;
		if (!context) {
			updateStep("create", "running");
			const createdResult = await create(flow.draft);
			if (!createdResult.ok) {
				updateStep("create", "failed");
				setError(createdResult.error.message);
				return;
			}
			context = createdResult.data.context;
			setCreated(context);
			updateStep("create", "complete");
			// Storage initialization happens inside the successful create use case.
			updateStep("initialize", "complete");
		} else {
			updateStep("create", "complete");
			updateStep("initialize", "complete");
		}
		if (flow.projects.length === 0 || !bindProject) {
			updateStep("bind", "skipped", "No project associations selected.");
		} else {
			updateStep("bind", "running");
			for (const project of flow.projects) {
				const bound = await bindProject({
					projectPath: project.path,
					contextId: context.id,
				});
				if (!bound.ok) {
					updateStep("bind", "failed");
					setError(bound.error.message);
					return;
				}
			}
			updateStep("bind", "complete");
		}
		updateStep("verify", "running");
		const verified = verifyContext
			? await verifyContext(context)
			: { ok: true as const, data: undefined };
		if (!verified.ok) {
			updateStep("verify", "failed");
			setError(verified.error.message);
			return;
		}
		updateStep("verify", "complete");
		setFlow((current) => ({ ...current, status: "success" }));
	}

	function createAnother() {
		setCreated(undefined);
		setError(undefined);
		setSteps(initialSteps);
		setFlow(initialContextCreateFlow());
	}
	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent>
				<SheetHeader>
					<SheetTitle>New context</SheetTitle>
					<SheetDescription>
						Create an isolated development identity.
					</SheetDescription>
				</SheetHeader>
				<div className="max-h-[calc(100vh-10rem)] overflow-y-auto px-8 pb-8">
					{flow.status === "identity" ? (
						<ContextCreateIdentityScreen
							draft={flow.draft}
							onDraftChange={(draft) =>
								setFlow((current) => updateContextCreateDraft(current, draft))
							}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "projects" ? (
						<ContextCreateProjectsScreen
							projects={flow.projects}
							onProjectsChange={(projects) =>
								setFlow((current) =>
									updateContextCreateProjects(current, projects),
								)
							}
							onBack={() => setFlow(previousContextCreateStep)}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "tools" ? (
						<ContextCreateDevelopmentToolsScreen
							integrations={integrations}
							loading={catalogLoading}
							error={catalogError}
							enabledIntegrationIds={flow.draft.enabledDevelopmentToolIds}
							onEnabledIntegrationIdsChange={(ids) =>
								setFlow((current) =>
									updateContextCreateDraft(current, {
										enabledDevelopmentToolIds: ids,
									}),
								)
							}
							onBack={() => setFlow(previousContextCreateStep)}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "review" ? (
						<ContextCreateReviewScreen
							draft={flow.draft}
							projects={flow.projects}
							integrations={integrations}
							onEdit={(section) =>
								setFlow((current) => editContextCreateSection(current, section))
							}
							onCreate={() => void submit()}
						/>
					) : null}
					{flow.status === "creating" ? (
						<ContextCreationProgress
							steps={steps}
							error={error}
							onRetry={() => void submit()}
							onBack={
								created
									? undefined
									: () =>
											setFlow((current) => ({ ...current, status: "review" }))
							}
						/>
					) : null}
					{flow.status === "success" && created ? (
						<ContextCreateSuccessScreen
							context={created}
							projectName={projectName}
							onOpenProject={onOpenProject}
							onRecheckContext={(context) => {
								void verifyContext?.(context).then((result) => {
									if (result?.ok) setCreated(result.data);
								});
							}}
							onViewContext={() => onViewContext?.(created.id)}
							onCreateAnother={createAnother}
						/>
					) : null}
				</div>
			</SheetContent>
		</Sheet>
	);
}

function uniqueIntegrations(
	contexts: ContextListItem[],
): DevelopmentToolIntegration[] {
	const integrations = contexts.flatMap(
		(item) => item.context.developmentTools ?? [],
	);
	return integrations.filter(
		(integration, index) =>
			integrations.findIndex((candidate) => candidate.id === integration.id) ===
			index,
	);
}
