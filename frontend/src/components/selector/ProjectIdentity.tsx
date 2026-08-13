import type { ProjectState } from "../../lib/devctx-api";

interface ProjectIdentityProps {
  project: ProjectState;
}

function ProjectIdentity({ project }: ProjectIdentityProps) {
  return (
    <section aria-labelledby="project-identity-heading" className="min-w-0 border-b border-border pb-6">
      <div className="min-w-0 space-y-2">
        <h3
          id="project-identity-heading"
          className="truncate text-xl font-semibold"
          title={project.name}
        >
          {project.name}
        </h3>
        <p
          className="truncate font-mono text-sm text-muted-foreground"
          title={project.path}
          aria-label={`Project path: ${project.path}`}
        >
          {project.path}
        </p>
      </div>
    </section>
  );
}

export { ProjectIdentity };
