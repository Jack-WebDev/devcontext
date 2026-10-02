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
	ContextCreateSuccessScreen,
	ContextCreationProgress,
} from "./ContextCreateCompletion";
import { ContextCreateDevelopmentToolsScreen } from "./ContextCreateDevelopmentToolsScreen";
import { ContextCreateIdentityScreen } from "./ContextCreateIdentityScreen";
import { ContextCreateProjectsScreen } from "./ContextCreateProjectsScreen";
import { ContextCreateReviewScreen } from "./ContextCreateReviewScreen";
import { useContextCreationJourney } from "./context-creation.js";

export { ContextDetailsDrawer } from "./ContextDetailsDrawer";

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
	const creation = useContextCreationJourney({
		createContext: create,
		bindProject,
		verifyContext,
		initialProjects,
	});
	const { setDefaultEnabledDevelopmentToolIds } = creation;
	const [catalogIntegrations, setCatalogIntegrations] =
		useState<DevelopmentToolIntegration[]>();
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
			setDefaultEnabledDevelopmentToolIds(
				result.data.developmentTools
					.filter((integration) => integration.enabled)
					.map((integration) => integration.id),
			);
		});
		return () => {
			active = false;
		};
	}, [loadCreationOptions, setDefaultEnabledDevelopmentToolIds]);

	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent className="max-w-152 overflow-hidden">
				<SheetHeader className="border-b border-border/60 bg-muted/20">
					<SheetTitle>New context</SheetTitle>
					<SheetDescription>
						Create an isolated development identity.
					</SheetDescription>
				</SheetHeader>
				<div className="flex-1 overflow-y-auto px-8 py-7">
					{creation.flow.status === "identity" ? (
						<ContextCreateIdentityScreen
							draft={creation.flow.draft}
							onDraftChange={creation.updateDraft}
							onContinue={creation.next}
						/>
					) : null}
					{creation.flow.status === "projects" ? (
						<ContextCreateProjectsScreen
							projects={creation.flow.projects}
							onProjectsChange={creation.updateProjects}
							onBack={creation.previous}
							onContinue={creation.next}
						/>
					) : null}
					{creation.flow.status === "tools" ? (
						<ContextCreateDevelopmentToolsScreen
							integrations={integrations}
							loading={catalogLoading}
							error={catalogError}
							enabledIntegrationIds={
								creation.flow.draft.enabledDevelopmentToolIds
							}
							onEnabledIntegrationIdsChange={(ids) =>
								creation.updateDraft({ enabledDevelopmentToolIds: ids })
							}
							onBack={creation.previous}
							onContinue={creation.next}
						/>
					) : null}
					{creation.flow.status === "review" ? (
						<ContextCreateReviewScreen
							draft={creation.flow.draft}
							projects={creation.flow.projects}
							integrations={integrations}
							onEdit={creation.edit}
							onCreate={() => void creation.submit()}
						/>
					) : null}
					{creation.flow.status === "creating" ? (
						<ContextCreationProgress
							steps={creation.steps}
							error={creation.error?.message}
							onRetry={() => void creation.submit()}
							onBack={
								creation.created === undefined
									? creation.returnToReview
									: undefined
							}
						/>
					) : null}
					{creation.flow.status === "success" && creation.created ? (
						<ContextCreateSuccessScreen
							context={creation.created}
							projectName={projectName}
							onOpenProject={onOpenProject}
							onRecheckContext={() => void creation.recheck()}
							onViewContext={() => {
								if (creation.created !== undefined) {
									onViewContext?.(creation.created.id);
								}
							}}
							onCreateAnother={creation.createAnother}
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
