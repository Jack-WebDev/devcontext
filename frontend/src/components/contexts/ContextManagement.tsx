import { useState } from "react";
import type {
	ApiResult,
	ContextListItem,
	CreateContextRequest,
	CreateContextResult,
	DevelopmentToolIntegration,
} from "../../lib/devctx-api";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "../ui/sheet.js";
import { useContextCreation } from "./context-creation";
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

export function CreateContextDialog({
	contexts,
	onClose,
	create,
}: {
	contexts: ContextListItem[];
	onClose: () => void;
	create: (request: CreateContextRequest) => Promise<ApiResult<CreateContextResult>>;
}) {
	const [flow, setFlow] = useState(initialContextCreateFlow);
	const contextCreation = useContextCreation(create);
	const integrations = uniqueIntegrations(contexts);

	async function submit() {
		const result = await contextCreation.create(flow.draft);
		if (result?.ok) onClose();
	}

	return (
		<Sheet open onOpenChange={(open) => !open && onClose()}>
			<SheetContent>
				<SheetHeader>
					<SheetTitle>New context</SheetTitle>
					<SheetDescription>Create an isolated development identity.</SheetDescription>
				</SheetHeader>
				<div className="max-h-[calc(100vh-10rem)] overflow-y-auto px-8 pb-8">
					{flow.status === "identity" ? (
						<ContextCreateIdentityScreen
							draft={flow.draft}
							onDraftChange={(draft) => setFlow((current) => updateContextCreateDraft(current, draft))}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "projects" ? (
						<ContextCreateProjectsScreen
							projects={flow.projects}
							onProjectsChange={(projects) => setFlow((current) => updateContextCreateProjects(current, projects))}
							onBack={() => setFlow(previousContextCreateStep)}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "tools" ? (
						<ContextCreateDevelopmentToolsScreen
							integrations={integrations}
							enabledIntegrationIds={flow.draft.enabledDevelopmentToolIds}
							onEnabledIntegrationIdsChange={(ids) => setFlow((current) => updateContextCreateDraft(current, { enabledDevelopmentToolIds: ids }))}
							onBack={() => setFlow(previousContextCreateStep)}
							onContinue={() => setFlow(nextContextCreateStep)}
						/>
					) : null}
					{flow.status === "review" ? (
						<ContextCreateReviewScreen
							draft={flow.draft}
							projects={flow.projects}
							integrations={integrations}
							pending={contextCreation.pending}
							error={contextCreation.error?.message}
							onEdit={(section) => setFlow((current) => editContextCreateSection(current, section))}
							onCreate={() => void submit()}
						/>
					) : null}
				</div>
			</SheetContent>
		</Sheet>
	);
}

function uniqueIntegrations(contexts: ContextListItem[]): DevelopmentToolIntegration[] {
	const integrations = contexts.flatMap((item) => item.context.developmentTools ?? []);
	return integrations.filter((integration, index) => integrations.findIndex((candidate) => candidate.id === integration.id) === index);
}
