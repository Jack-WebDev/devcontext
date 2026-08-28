import type { ProjectListItem } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { ProjectSafetyLabel } from "./ProjectSafetyLabel.js";

interface ProjectsViewProps {
  projects: ProjectListItem[];
  launchingProjectPath?: string;
  errorProjectPath?: string;
  launchError?: string;
  onLaunch?: (project: ProjectListItem) => void;
  onChangeContext?: (project: ProjectListItem) => void;
  onOpenFolder?: (project: ProjectListItem) => void;
  onForget?: (project: ProjectListItem) => void;
}

function ProjectsView({
  projects,
  launchingProjectPath,
  errorProjectPath,
  launchError,
  onLaunch,
  onChangeContext,
  onOpenFolder,
  onForget,
}: ProjectsViewProps) {
  return (
    <section aria-labelledby="projects-heading" className="space-y-6">
      <div>
        <p className="text-sm text-muted-foreground">Known projects</p>
        <h2 id="projects-heading" className="text-2xl font-semibold">Projects</h2>
        <p className="mt-1 text-sm text-muted-foreground">Review the context Dev Context will use before launching a project.</p>
      </div>

      {projects.length === 0 ? (
        <Card as="section" hierarchy="secondary" className="py-0">
          <CardContent className="p-5 text-sm text-muted-foreground">
            No projects have been launched yet. Projects appear here after a successful launch.
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-4" aria-label="Known projects">
          {projects.map((project) => (
            <ProjectCard
              key={project.project.path}
              project={project}
              launching={launchingProjectPath === project.project.path}
              launchError={errorProjectPath === project.project.path ? launchError : undefined}
              onLaunch={onLaunch}
              onChangeContext={onChangeContext}
              onOpenFolder={onOpenFolder}
              onForget={onForget}
            />
          ))}
        </div>
      )}
    </section>
  );
}

function ProjectCard({
  project,
  launching,
  launchError,
  onLaunch,
  onChangeContext,
  onOpenFolder,
  onForget,
}: {
  project: ProjectListItem;
  launching: boolean;
  launchError?: string;
  onLaunch?: (project: ProjectListItem) => void;
  onChangeContext?: (project: ProjectListItem) => void;
  onOpenFolder?: (project: ProjectListItem) => void;
  onForget?: (project: ProjectListItem) => void;
}) {
  const contextName = project.contextName ?? project.contextId ?? "No context selected";
  const canLaunch = project.contextId !== undefined && onLaunch !== undefined;

  return (
    <Card as="article" hierarchy="secondary" className="py-0" aria-labelledby={`project-${project.project.path}-heading`}>
      <CardContent className="space-y-4 p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <h3 id={`project-${project.project.path}-heading`} className="truncate text-lg font-semibold" title={project.project.name}>
              {project.project.name}
            </h3>
            <p className="mt-1 truncate font-mono text-sm text-muted-foreground" title={project.project.path}>{project.project.path}</p>
          </div>
          <div className="flex shrink-0 flex-wrap justify-end gap-2"><ProjectSafetyLabel contextName={project.contextName ?? project.contextId} /><span className="text-sm font-medium text-muted-foreground">{project.running ? "Running" : "Not running"}</span></div>
        </div>

        <dl className="grid gap-3 border-t border-border pt-4 text-sm sm:grid-cols-2">
          <ProjectDetail label="Remembered context" value={contextName} />
          <ProjectDetail label="Last launched" value={formatProjectTime(project.lastLaunchedAt)} />
        </dl>

        {launchError ? <p className="text-sm text-destructive" role="alert">{launchError}</p> : null}

        <div className="flex flex-wrap gap-3 border-t border-border pt-4" aria-label={`Actions for ${project.project.name}`}>
          <Button type="button" size="sm" disabled={!canLaunch || launching} onClick={() => onLaunch?.(project)}>
            {launching ? `Launching ${contextName}...` : `Launch ${contextName}`}
          </Button>
          <Button type="button" variant="outline" size="sm" disabled={onChangeContext === undefined} onClick={() => onChangeContext?.(project)}>
            Change context
          </Button>
          <Button type="button" variant="outline" size="sm" disabled={onOpenFolder === undefined} onClick={() => onOpenFolder?.(project)}>
            Open folder
          </Button>
          <Button
            type="button"
            variant="destructive"
            size="sm"
            disabled={onForget === undefined}
            title={onForget === undefined ? "Forgetting projects will be available when project removal is supported." : undefined}
            onClick={() => onForget?.(project)}
          >
            Forget project
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function ProjectDetail({label, value}: {label: string; value: string}) {
  return (
    <div className="min-w-0">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="mt-1 truncate font-medium" title={value}>{value}</dd>
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

export { ProjectsView, formatProjectTime };
export type { ProjectsViewProps };
