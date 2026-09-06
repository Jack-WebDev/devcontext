import type { ProjectListItem } from "../../lib/devctx-api";
import type { ReactNode } from "react";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { formatProjectTime } from "./ProjectsView.js";

interface ProjectDetailViewProps {
	project?: ProjectListItem;
	onBack: () => void;
	onLaunch?: (project: ProjectListItem) => void;
	onOpenFolder?: (project: ProjectListItem) => void;
	onChangeContext?: (project: ProjectListItem) => void;
	onRemoveBinding?: (project: ProjectListItem) => void;
	onForget?: (project: ProjectListItem) => void;
	onLocate?: (project: ProjectListItem) => void;
}

function ProjectDetailView({
	project,
	onBack,
	onLaunch,
	onOpenFolder,
	onChangeContext,
	onRemoveBinding,
	onForget,
	onLocate,
}: ProjectDetailViewProps) {
	if (project === undefined) {
		return (
			<section aria-labelledby="project-detail-heading" className="space-y-6">
				<div>
					<p className="text-sm text-muted-foreground">Projects</p>
					<h2 id="project-detail-heading" className="text-2xl font-semibold">
						Project not found
					</h2>
					<p className="mt-1 text-sm text-muted-foreground">
						This project is no longer in Dev Context's known projects.
					</p>
				</div>
				<Button type="button" variant="outline" onClick={onBack}>
					Back to projects
				</Button>
			</section>
		);
	}

	const isAssigned = project.contextId !== undefined;
	const contextName = project.contextName ?? project.contextId ?? "Unassigned";
	const canLaunch = isAssigned && onLaunch !== undefined;

	return (
		<section aria-labelledby="project-detail-heading" className="space-y-6">
			<nav aria-label="Breadcrumb" className="flex items-center gap-2 text-sm text-muted-foreground">
				<button
					type="button"
					className="rounded-sm hover:text-foreground focus-visible:outline-2 focus-visible:outline-ring focus-visible:outline-offset-2"
					onClick={onBack}
				>
					Projects
				</button>
				<span aria-hidden="true">/</span>
				<span aria-current="page" className="truncate text-foreground">
					{project.project.name}
				</span>
			</nav>

			<div className="flex flex-wrap items-start justify-between gap-4">
				<div className="min-w-0">
					<p className="text-sm text-muted-foreground">Project</p>
					<h2 id="project-detail-heading" className="truncate text-2xl font-semibold">
						{project.project.name}
					</h2>
					<p className="mt-1 truncate font-mono text-sm text-muted-foreground" title={project.project.path}>
						{project.project.path}
					</p>
				</div>
				<Button type="button" variant="outline" onClick={onBack}>
					Back to projects
				</Button>
			</div>

			<div className="grid gap-4 lg:grid-cols-2">
				<ProjectDetailSection title="Context">
					<ProjectDetailRow label="Normal context" value={contextName} />
					<ProjectDetailRow
						label="Binding"
						value={
							isAssigned
								? "Assigned — this is the context Dev Context normally uses."
								: "Unassigned — choose a context before launching."
						}
					/>
				</ProjectDetailSection>
				<ProjectDetailSection title="Launch behavior">
					<p className="text-sm text-muted-foreground">
						{isAssigned
							? `Launching this project uses ${contextName} unless you choose a different context for that launch.`
							: "This project has no normal context. Dev Context will ask you to choose one before launching."}
					</p>
				</ProjectDetailSection>
				<ProjectDetailSection title="Workspace">
					<p className="text-sm text-muted-foreground">
						{project.running
							? "This project has an active Dev Context workspace."
							: "This project has no active Dev Context workspace."}
					</p>
				</ProjectDetailSection>
				<ProjectDetailSection title="Activity">
					<ProjectDetailRow
						label="Last opened"
						value={formatProjectTime(project.lastLaunchedAt)}
					/>
				</ProjectDetailSection>
			</div>

			<div className="flex flex-wrap gap-3">
				<Button
					type="button"
					disabled={!canLaunch}
					onClick={() => onLaunch?.(project)}
				>
					{isAssigned ? `Launch ${contextName}` : "Choose a context to launch"}
				</Button>
				<Button
					type="button"
					variant="outline"
					disabled={onOpenFolder === undefined}
					onClick={() => onOpenFolder?.(project)}
				>
					Open folder
				</Button>
				{isAssigned ? (
					<>
						<Button
							type="button"
							variant="outline"
							disabled={onChangeContext === undefined}
							onClick={() => onChangeContext?.(project)}
						>
							Change context
						</Button>
						<Button
							type="button"
							variant="destructive"
							disabled={onRemoveBinding === undefined}
							onClick={() => onRemoveBinding?.(project)}
						>
							Remove binding
						</Button>
					</>
				) : null}
				<Button
					type="button"
					variant="destructive"
					disabled={onForget === undefined}
					onClick={() => onForget?.(project)}
				>
					Forget project
				</Button>
				<Button
					type="button"
					variant="outline"
					disabled={!isAssigned || onLocate === undefined}
					title={!isAssigned ? "Assign a context before relocating this project." : undefined}
					onClick={() => onLocate?.(project)}
				>
					Locate project
				</Button>
			</div>
		</section>
	);
}

function ProjectDetailSection({
	title,
	children,
}: {
	title: string;
	children: ReactNode;
}) {
	return (
		<Card as="section" hierarchy="secondary" className="py-0">
			<CardContent className="space-y-3 p-5">
				<h3 className="font-semibold">{title}</h3>
				{children}
			</CardContent>
		</Card>
	);
}

function ProjectDetailRow({ label, value }: { label: string; value: string }) {
	return (
		<div className="grid gap-1 text-sm">
			<p className="text-muted-foreground">{label}</p>
			<p className="font-medium">{value}</p>
		</div>
	);
}

export type { ProjectDetailViewProps };
export { ProjectDetailView };
