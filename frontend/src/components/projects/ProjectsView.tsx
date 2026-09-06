import { useState } from "react";
import type { ProjectListItem } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ProjectsViewProps {
	projects: ProjectListItem[];
	launchingProjectPath?: string;
	errorProjectPath?: string;
	launchError?: string;
	onLaunch?: (project: ProjectListItem) => void;
	onOpenFolder?: (project: ProjectListItem) => void;
	onOpenDetail?: (project: ProjectListItem) => void;
	onForget?: (project: ProjectListItem) => void;
}

type ProjectAssignmentFilter = "all" | "assigned" | "unassigned";

function ProjectsView({
	projects,
	launchingProjectPath,
	errorProjectPath,
	launchError,
	onLaunch,
	onOpenFolder,
	onOpenDetail,
	onForget,
}: ProjectsViewProps) {
	return (
		<section aria-labelledby="projects-heading" className="space-y-6">
			<div>
				<p className="text-sm text-muted-foreground">Known projects</p>
				<h2 id="projects-heading" className="text-2xl font-semibold">
					Projects
				</h2>
				<p className="mt-1 text-sm text-muted-foreground">
					Review the context Dev Context will use before launching a project.
				</p>
			</div>

			<FilteredProjects
				projects={projects}
				launchingProjectPath={launchingProjectPath}
				errorProjectPath={errorProjectPath}
				launchError={launchError}
				onLaunch={onLaunch}
				onOpenFolder={onOpenFolder}
				onOpenDetail={onOpenDetail}
				onForget={onForget}
			/>
		</section>
	);
}

function FilteredProjects(props: ProjectsViewProps) {
	const [assignmentFilter, setAssignmentFilter] =
		useState<ProjectAssignmentFilter>("all");
	const visibleProjects = filterProjects(props.projects, assignmentFilter);
	const emptyMessage =
		props.projects.length === 0
			? "No projects have been launched yet. Projects appear here after a successful launch."
			: assignmentFilter === "assigned"
				? "No assigned projects yet."
				: "No unassigned projects yet.";

	return (
		<>
			{props.projects.length > 0 ? (
				<fieldset className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm">
					<legend className="font-medium">Project assignment</legend>
					{([
						["all", "All projects"],
						["assigned", "Assigned"],
						["unassigned", "Unassigned"],
					] as const).map(([value, label]) => (
						<label key={value} className="flex items-center gap-2">
							<input
								type="radio"
								name="project-assignment-filter"
								value={value}
								checked={assignmentFilter === value}
								onChange={() => setAssignmentFilter(value)}
							/>
							{label}
						</label>
					))}
				</fieldset>
			) : null}

			{visibleProjects.length === 0 ? (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="p-5 text-sm text-muted-foreground">
						{emptyMessage}
					</CardContent>
				</Card>
			) : (
				<div className="space-y-4">
					{visibleProjects.map((project, index) => (
						<ProjectCard
							key={project.project.path}
							id={`project-${index}-heading`}
							project={project}
							launching={props.launchingProjectPath === project.project.path}
							launchError={
								props.errorProjectPath === project.project.path
									? props.launchError
									: undefined
							}
							onLaunch={props.onLaunch}
							onOpenFolder={props.onOpenFolder}
							onOpenDetail={props.onOpenDetail}
							onForget={props.onForget}
						/>
					))}
				</div>
			)}
		</>
	);
}

function ProjectCard({
	id,
	project,
	launching,
	launchError,
	onLaunch,
	onOpenFolder,
	onOpenDetail,
	onForget,
}: {
	id: string;
	project: ProjectListItem;
	launching: boolean;
	launchError?: string;
	onLaunch?: (project: ProjectListItem) => void;
	onOpenFolder?: (project: ProjectListItem) => void;
	onOpenDetail?: (project: ProjectListItem) => void;
	onForget?: (project: ProjectListItem) => void;
}) {
	const isAssigned = project.contextId !== undefined;
	const contextName = project.contextName ?? project.contextId ?? "Unassigned";
	const canLaunch = project.contextId !== undefined && onLaunch !== undefined;

	return (
		<Card
			as="article"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby={id}
		>
			<CardContent className="space-y-4 p-5">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<h3
							id={id}
							className="truncate text-lg font-semibold"
							title={project.project.name}
						>
							{project.project.name}
						</h3>
					</div>
					<span
						className={
							isAssigned
								? "inline-flex shrink-0 border border-border bg-muted/30 px-2 py-1 text-xs font-medium text-foreground"
								: "inline-flex shrink-0 border border-border bg-background px-2 py-1 text-xs font-medium text-muted-foreground"
						}
					>
						{isAssigned ? "Assigned" : "Unassigned"}
					</span>
				</div>

				<dl className="grid gap-4 border-y border-border py-4 text-sm sm:grid-cols-2">
					<ProjectDetail
						label="Normal context"
						value={contextName}
						prominent
					/>
					<ProjectDetail
						label="Binding state"
						value={
							isAssigned
								? "This project opens with its normal context."
								: "Choose a context before launching."
						}
					/>
					<ProjectDetail label="Path" value={project.project.path} mono />
					<ProjectDetail
						label="Last opened"
						value={formatProjectTime(project.lastLaunchedAt)}
					/>
					<ProjectDetail
						label="Workspace"
						value={project.running ? "Active" : "Not active"}
					/>
				</dl>

				{launchError ? (
					<p className="text-sm text-destructive" role="alert">
						{launchError}
					</p>
				) : null}

				<div className="flex flex-wrap gap-3">
					<Button
						type="button"
						variant="outline"
						size="sm"
						disabled={onOpenDetail === undefined}
						onClick={() => onOpenDetail?.(project)}
					>
						View details
					</Button>
					<Button
						type="button"
						size="sm"
						disabled={!canLaunch || launching}
						onClick={() => onLaunch?.(project)}
					>
						{launching ? `Launching ${contextName}...` : `Launch ${contextName}`}
					</Button>
					<Button
						type="button"
						variant="outline"
						size="sm"
						disabled={onOpenFolder === undefined}
						onClick={() => onOpenFolder?.(project)}
					>
						Open folder
					</Button>
					{onForget ? (
						<Button
							type="button"
							variant="destructive"
							size="sm"
							onClick={() => onForget(project)}
						>
							Forget project
						</Button>
					) : null}
				</div>
			</CardContent>
		</Card>
	);
}

function filterProjects(
	projects: ProjectListItem[],
	filter: ProjectAssignmentFilter,
): ProjectListItem[] {
	if (filter === "all") return projects;
	const assigned = filter === "assigned";
	return projects.filter((project) => (project.contextId !== undefined) === assigned);
}

function ProjectDetail({
	label,
	value,
	prominent = false,
	mono = false,
}: {
	label: string;
	value: string;
	prominent?: boolean;
	mono?: boolean;
}) {
	return (
		<div className="min-w-0">
			<dt className="text-muted-foreground">{label}</dt>
			<dd
				className={`mt-1 truncate ${
					prominent ? "text-base font-semibold" : "font-medium"
				} ${mono ? "font-mono text-xs" : ""}`}
				title={value}
			>
				{value}
			</dd>
		</div>
	);
}

function formatProjectTime(value: string | undefined): string {
	if (value === undefined) {
		return "Never launched";
	}
	const time = new Date(value);
	return Number.isNaN(time.getTime()) ? "Unavailable" : time.toLocaleString();
}

export type { ProjectAssignmentFilter, ProjectsViewProps };
export { filterProjects, formatProjectTime, ProjectsView };
