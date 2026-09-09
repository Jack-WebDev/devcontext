import { useState } from "react";
import type { ProjectListItem } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { EmptyState } from "../ui/collection-state.js";
import { PageHeader } from "../ui/page-header.js";

interface ProjectsViewProps {
	projects: ProjectListItem[];
	launchingProjectPath?: string;
	errorProjectPath?: string;
	launchError?: string;
	onLaunch?: (project: ProjectListItem) => void;
	onOpenFolder?: (project: ProjectListItem) => void;
	onOpenDetail?: (project: ProjectListItem) => void;
	onForget?: (project: ProjectListItem) => void;
	onStartLaunch?: () => void;
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
	onStartLaunch,
}: ProjectsViewProps) {
	return (
		<section
			aria-labelledby="projects-heading"
			className="page-content page-section-stack"
		>
			<PageHeader
				id="projects-heading"
				eyebrow="Known projects"
				title="Projects"
				description="Review the context Dev Context will use before launching a project."
			/>

			<FilteredProjects
				projects={projects}
				launchingProjectPath={launchingProjectPath}
				errorProjectPath={errorProjectPath}
				launchError={launchError}
				onLaunch={onLaunch}
				onOpenFolder={onOpenFolder}
				onOpenDetail={onOpenDetail}
				onForget={onForget}
				onStartLaunch={onStartLaunch}
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
				<fieldset className="flex flex-wrap items-center gap-1 rounded-xl bg-muted/45 p-1.5 text-sm">
					<legend className="sr-only">Project assignment</legend>
					{(
						[
							["all", "All projects"],
							["assigned", "Assigned"],
							["unassigned", "Unassigned"],
						] as const
					).map(([value, label]) => (
						<label
							key={value}
							className="has-[:checked]:bg-card has-[:checked]:shadow-sm flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-medium text-muted-foreground transition-colors has-[:checked]:text-foreground has-[:focus-visible]:ring-2 has-[:focus-visible]:ring-ring/25"
						>
							<input
								className="sr-only"
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
				<EmptyState
					title={
						props.projects.length === 0
							? "No projects yet"
							: "No matching projects"
					}
					description={emptyMessage}
					{...(props.projects.length === 0
						? { actionLabel: "Launch a project", onAction: props.onStartLaunch }
						: {})}
				/>
			) : (
				<div className="collection-surface divide-y divide-border/50">
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
			hierarchy="tertiary"
			className="collection-row rounded-none py-0"
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
								? "inline-flex shrink-0 rounded-md bg-[var(--green-soft)] px-2.5 py-1 text-xs font-semibold text-[var(--green-strong)]"
								: "inline-flex shrink-0 rounded-md bg-muted px-2.5 py-1 text-xs font-medium text-muted-foreground"
						}
					>
						{isAssigned ? "Assigned" : "Unassigned"}
					</span>
				</div>

				<dl className="grid gap-x-8 gap-y-4 rounded-xl bg-muted/30 p-4 text-sm sm:grid-cols-2 lg:grid-cols-3">
					<ProjectDetail label="Normal context" value={contextName} prominent />
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

				<div className="flex flex-wrap items-center gap-2">
					<Button
						type="button"
						variant="ghost"
						size="sm"
						disabled={onOpenDetail === undefined}
						onClick={() => onOpenDetail?.(project)}
					>
						View details
					</Button>
					<Button
						type="button"
						size="sm"
						className="ml-auto"
						disabled={!canLaunch || launching}
						onClick={() => onLaunch?.(project)}
					>
						{launching
							? `Launching ${contextName}...`
							: `Launch ${contextName}`}
					</Button>
					<Button
						type="button"
						variant="ghost"
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
	return projects.filter(
		(project) => (project.contextId !== undefined) === assigned,
	);
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
