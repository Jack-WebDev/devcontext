import type {
	CreateContextRequest,
	DevelopmentToolIntegration,
	ProjectState,
} from "../../lib/devctx-api.js";
import type { ReactNode } from "react";
import { Button } from "../ui/button.js";

type ReviewSection = "identity" | "projects" | "tools";

interface ContextCreateReviewScreenProps {
	draft: CreateContextRequest;
	projects: ProjectState[];
	integrations: DevelopmentToolIntegration[];
	pending?: boolean;
	error?: string;
	onEdit: (section: ReviewSection) => void;
	onCreate: () => void;
}

function selectedIntegrations(
	draft: CreateContextRequest,
	integrations: DevelopmentToolIntegration[],
): DevelopmentToolIntegration[] {
	const selected = new Set(draft.enabledDevelopmentToolIds ?? []);
	return integrations.filter((integration) => selected.has(integration.id));
}

function ContextCreateReviewScreen({
	draft,
	projects,
	integrations,
	pending = false,
	error,
	onEdit,
	onCreate,
}: ContextCreateReviewScreenProps) {
	const tools = selectedIntegrations(draft, integrations);
	return (
		<section aria-labelledby="context-review-title" className="mx-auto max-w-xl space-y-6">
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">Create a context</p>
				<h2 id="context-review-title" className="text-2xl font-semibold">Review your context</h2>
				<p className="text-sm text-muted-foreground">
					Confirm what will stay together before creating this isolated development identity.
				</p>
			</div>

			<ReviewSection title="Identity" onEdit={() => onEdit("identity")}>
				<p className="font-medium">{draft.name?.trim() || "Untitled context"}</p>
				{draft.purpose ? <p className="text-sm text-muted-foreground">{draft.purpose}</p> : null}
				{draft.description ? <p className="text-sm text-muted-foreground">{draft.description}</p> : null}
				<p className="text-sm text-muted-foreground">Appearance: {draft.icon || "Default icon"}, {draft.accent || "default accent"}</p>
			</ReviewSection>

			<ReviewSection title="Projects" onEdit={() => onEdit("projects")}>
				{projects.length === 0 ? <p className="text-sm text-muted-foreground">No projects linked yet. You can add them later.</p> : (
					<ul className="space-y-1 text-sm">{projects.map((project) => <li key={project.path}>{project.name} <span className="text-muted-foreground">({project.path})</span></li>)}</ul>
				)}
			</ReviewSection>

			<ReviewSection title="Development tools" onEdit={() => onEdit("tools")}>
				{tools.length === 0 ? <p className="text-sm text-muted-foreground">No development tools selected. You can configure them later.</p> : (
					<ul className="space-y-1 text-sm">{tools.map((tool) => <li key={tool.id}>{tool.name} <span className="text-muted-foreground">({tool.category})</span></li>)}</ul>
				)}
			</ReviewSection>

			<ReviewSection title="Isolation and launch behavior">
				<p className="text-sm text-muted-foreground">This context keeps its selected tools and their accounts separate. When you choose it for a project, Dev Context opens that project with this context’s launch tool.</p>
			</ReviewSection>

			{error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}
			<div className="flex justify-end"><Button type="button" disabled={pending} onClick={onCreate}>{pending ? "Creating..." : "Create context"}</Button></div>
		</section>
	);
}

function ReviewSection({ title, children, onEdit }: { title: string; children: ReactNode; onEdit?: () => void }) {
	return <section className="rounded-xl border bg-card p-4" aria-label={title}>
		<div className="flex items-center justify-between gap-3"><h3 className="font-medium">{title}</h3>{onEdit ? <Button type="button" variant="ghost" size="sm" onClick={onEdit}>Edit</Button> : null}</div>
		<div className="mt-3 space-y-1">{children}</div>
	</section>;
}

export type { ContextCreateReviewScreenProps, ReviewSection };
export { ContextCreateReviewScreen, selectedIntegrations };
