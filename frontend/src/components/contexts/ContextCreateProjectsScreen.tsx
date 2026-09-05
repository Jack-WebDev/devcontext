import { useEffect, useState, type DragEvent } from "react";

import {
	devContextApi,
	type ApiResult,
	type ProjectState,
	type RecentProjectsState,
} from "../../lib/devctx-api.js";
import { Button } from "../ui/button.js";

interface ContextCreateProjectsScreenProps {
	projects: ProjectState[];
	onProjectsChange: (projects: ProjectState[]) => void;
	onBack?: () => void;
	onContinue: () => void;
	chooseProjectDirectory?: () => Promise<ApiResult<string | undefined>>;
	validateProjectDirectory?: (request: {
		projectPath: string;
	}) => Promise<ApiResult<ProjectState>>;
	getRecentProjects?: () => Promise<ApiResult<RecentProjectsState>>;
	initialRecentProjects?: ProjectState[];
}

interface ProjectDropData {
	files: Iterable<unknown>;
	getData(format: string): string;
}

function addProjectToDraft(
	projects: ProjectState[],
	project: ProjectState,
): ProjectState[] {
	return projects.some((candidate) => candidate.path === project.path)
		? projects
		: [...projects, project];
}

function recentProjectChoices(
	recentProjects: ProjectState[],
	selectedProjects: ProjectState[],
): ProjectState[] {
	return recentProjects.filter(
		(project, index) =>
			!selectedProjects.some((selected) => selected.path === project.path) &&
			recentProjects.findIndex(
				(candidate) => candidate.path === project.path,
			) === index,
	);
}

function projectPathsFromDrop(data: ProjectDropData): string[] {
	const paths = Array.from(
		data.files,
		(file) => (file as { path?: unknown }).path,
	).filter(
		(path): path is string => typeof path === "string" && path.length > 0,
	);
	if (paths.length > 0) return paths;

	return data
		.getData("text/uri-list")
		.split(/\r?\n/)
		.filter((value) => value && !value.startsWith("#"))
		.flatMap((value) => {
			try {
				const url = new URL(value);
				if (url.protocol !== "file:" || url.hostname) return [];
				const path = decodeURIComponent(url.pathname);
				return [/^\/[a-zA-Z]:\//.test(path) ? path.slice(1) : path];
			} catch {
				return [];
			}
		});
}

function ContextCreateProjectsScreen({
	projects,
	onProjectsChange,
	onBack,
	onContinue,
	chooseProjectDirectory = devContextApi.chooseProjectDirectory,
	validateProjectDirectory = devContextApi.validateProjectDirectory,
	getRecentProjects = devContextApi.getRecentProjects,
	initialRecentProjects = [],
}: ContextCreateProjectsScreenProps) {
	const [addingProject, setAddingProject] = useState(false);
	const [error, setError] = useState<string>();
	const [recentProjects, setRecentProjects] = useState(initialRecentProjects);
	const [dropActive, setDropActive] = useState(false);

	useEffect(() => {
		let active = true;
		void getRecentProjects().then((result) => {
			if (active && result.ok) {
				setRecentProjects(result.data.projects.map((recent) => recent.project));
			}
		});
		return () => {
			active = false;
		};
	}, [getRecentProjects]);

	async function addProjectPaths(paths: string[]) {
		if (paths.length === 0) {
			setError("Drop a local project folder to add it.");
			return;
		}

		setAddingProject(true);
		setError(undefined);
		try {
			let nextProjects = projects;
			for (const path of paths) {
				const validated = await validateProjectDirectory({ projectPath: path });
				if (!validated.ok) {
					setError(validated.error.message);
					continue;
				}
				nextProjects = addProjectToDraft(nextProjects, validated.data);
			}
			if (nextProjects !== projects) onProjectsChange(nextProjects);
		} finally {
			setAddingProject(false);
		}
	}

	async function addProject() {
		setAddingProject(true);
		setError(undefined);
		try {
			const selected = await chooseProjectDirectory();
			if (!selected.ok || selected.data === undefined) {
				if (!selected.ok) setError(selected.error.message);
				return;
			}
			await addProjectPaths([selected.data]);
		} finally {
			setAddingProject(false);
		}
	}

	function handleDrop(event: DragEvent<HTMLDivElement>) {
		event.preventDefault();
		setDropActive(false);
		void addProjectPaths(projectPathsFromDrop(event.dataTransfer));
	}

	const availableRecentProjects = recentProjectChoices(
		recentProjects,
		projects,
	);

	return (
		<section
			aria-labelledby="context-projects-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Create a context
				</p>
				<h2 id="context-projects-title" className="text-2xl font-semibold">
					Which projects normally belong to this context?
				</h2>
				<p className="text-sm text-muted-foreground">
					Add folders now to prepare project associations. Nothing is changed
					until you create the context.
				</p>
			</div>

			<div
				className={`space-y-3 rounded-xl ${dropActive ? "ring-2 ring-ring ring-offset-2" : ""}`}
				onDragEnter={(event) => {
					event.preventDefault();
					setDropActive(true);
				}}
				onDragOver={(event) => event.preventDefault()}
				onDragLeave={() => setDropActive(false)}
				onDrop={handleDrop}
			>
				{projects.length === 0 ? (
					<p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
						No projects selected yet.
					</p>
				) : (
					<ul className="space-y-2" aria-label="Pending projects">
						{projects.map((project) => (
							<li
								key={project.path}
								className="flex items-center justify-between gap-3 rounded-lg border bg-card p-3"
							>
								<span className="min-w-0">
									<span className="block truncate text-sm font-medium">
										{project.name}
									</span>
									<span className="block truncate text-xs text-muted-foreground">
										{project.path}
									</span>
								</span>
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={() =>
										onProjectsChange(
											projects.filter(
												(candidate) => candidate.path !== project.path,
											),
										)
									}
								>
									Remove
								</Button>
							</li>
						))}
					</ul>
				)}
				<Button
					type="button"
					variant="outline"
					disabled={addingProject}
					onClick={() => void addProject()}
				>
					{addingProject ? "Adding project..." : "Choose folder"}
				</Button>
				<p className="text-xs text-muted-foreground">
					Or drop local project folders here.
				</p>
				{availableRecentProjects.length > 0 ? (
					<div className="space-y-2">
						<h3 className="text-sm font-medium">Recent projects</h3>
						<ul className="space-y-2" aria-label="Recent projects">
							{availableRecentProjects.map((project) => (
								<li key={project.path}>
									<Button
										type="button"
										variant="outline"
										className="h-auto w-full justify-start px-3 py-2 text-left"
										disabled={addingProject}
										onClick={() => void addProjectPaths([project.path])}
									>
										<span className="min-w-0">
											<span className="block truncate">{project.name}</span>
											<span className="block truncate text-xs font-normal text-muted-foreground">
												{project.path}
											</span>
										</span>
									</Button>
								</li>
							))}
						</ul>
					</div>
				) : null}
				{error ? (
					<p role="alert" className="text-sm text-destructive">
						{error}
					</p>
				) : null}
			</div>

			<div className="flex items-center justify-between gap-3">
				{onBack ? (
					<Button type="button" variant="outline" onClick={onBack}>
						Back
					</Button>
				) : (
					<span />
				)}
				<Button type="button" onClick={onContinue}>
					Continue
				</Button>
			</div>
		</section>
	);
}

export type { ContextCreateProjectsScreenProps };
export {
	ContextCreateProjectsScreen,
	addProjectToDraft,
	projectPathsFromDrop,
	recentProjectChoices,
};
