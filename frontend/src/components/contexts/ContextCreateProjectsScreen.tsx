import { useState } from "react";

import {
	devContextApi,
	type ApiResult,
	type ProjectState,
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
}

function addProjectToDraft(
	projects: ProjectState[],
	project: ProjectState,
): ProjectState[] {
	return projects.some((candidate) => candidate.path === project.path)
		? projects
		: [...projects, project];
}

function ContextCreateProjectsScreen({
	projects,
	onProjectsChange,
	onBack,
	onContinue,
	chooseProjectDirectory = devContextApi.chooseProjectDirectory,
	validateProjectDirectory = devContextApi.validateProjectDirectory,
}: ContextCreateProjectsScreenProps) {
	const [picking, setPicking] = useState(false);
	const [error, setError] = useState<string>();

	async function addProject() {
		setPicking(true);
		setError(undefined);
		try {
			const selected = await chooseProjectDirectory();
			if (!selected.ok || selected.data === undefined) {
				if (!selected.ok) setError(selected.error.message);
				return;
			}

			const validated = await validateProjectDirectory({
				projectPath: selected.data,
			});
			if (!validated.ok) {
				setError(validated.error.message);
				return;
			}
			onProjectsChange(addProjectToDraft(projects, validated.data));
		} finally {
			setPicking(false);
		}
	}

	return (
		<section aria-labelledby="context-projects-title" className="mx-auto max-w-xl space-y-6">
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">Create a context</p>
				<h2 id="context-projects-title" className="text-2xl font-semibold">
					Which projects normally belong to this context?
				</h2>
				<p className="text-sm text-muted-foreground">
					Add folders now to prepare project associations. Nothing is changed until you create the context.
				</p>
			</div>

			<div className="space-y-3">
				{projects.length === 0 ? (
					<p className="rounded-lg border border-dashed p-4 text-sm text-muted-foreground">
						No projects selected yet.
					</p>
				) : (
					<ul className="space-y-2" aria-label="Pending projects">
						{projects.map((project) => (
							<li key={project.path} className="flex items-center justify-between gap-3 rounded-lg border bg-card p-3">
								<span className="min-w-0">
									<span className="block truncate text-sm font-medium">{project.name}</span>
									<span className="block truncate text-xs text-muted-foreground">{project.path}</span>
								</span>
								<Button type="button" variant="ghost" size="sm" onClick={() => onProjectsChange(projects.filter((candidate) => candidate.path !== project.path))}>
									Remove
								</Button>
							</li>
						))}
					</ul>
				)}
				<Button type="button" variant="outline" disabled={picking} onClick={() => void addProject()}>
					{picking ? "Opening folder picker..." : "Choose folder"}
				</Button>
				{error ? <p role="alert" className="text-sm text-destructive">{error}</p> : null}
			</div>

			<div className="flex items-center justify-between gap-3">
				{onBack ? <Button type="button" variant="outline" onClick={onBack}>Back</Button> : <span />}
				<Button type="button" onClick={onContinue}>Continue</Button>
			</div>
		</section>
	);
}

export type { ContextCreateProjectsScreenProps };
export { ContextCreateProjectsScreen, addProjectToDraft };
