import type { ProjectState } from "../../lib/devctx-api";

interface ProjectIdentityProps {
  project: ProjectState;
}

function ProjectIdentity({ project }: ProjectIdentityProps) {
  return (
    <section
      aria-labelledby="project-identity-heading"
      className="min-w-0 border border-border bg-card px-4 py-3 shadow-sm sm:px-5"
      data-selector-project-identity
    >
      <div className="min-w-0 space-y-1.5">
        <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Current project</p>
        <h3
          id="project-identity-heading"
          className="truncate text-lg font-semibold"
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
      <dl className="mt-3 grid grid-cols-2 gap-x-6 border-t border-border pt-3 text-xs">
        <ProjectMetadataPlaceholder label="Git branch" />
        <ProjectMetadataPlaceholder label="Last opened" />
      </dl>
    </section>
  );
}

function ProjectMetadataPlaceholder({ label }: { label: string }) {
  return (
    <div className="min-w-0">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium text-muted-foreground">Unavailable</dd>
    </div>
  );
}

export { ProjectIdentity };
