import type { DisplayError, HomeDashboardState, RecentProjectState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { StatusIndicator } from "../status/StatusIndicator.js";

interface HomeViewProps {
  dashboard: HomeDashboardState;
  launchPending: boolean;
  launchError?: DisplayError;
  onQuickLaunch: () => void;
  onReviewLaunchOptions: () => void;
  onRecentProjectSelect?: (project: RecentProjectState) => void;
}

function HomeView({
  dashboard,
  launchPending,
  launchError,
  onQuickLaunch,
  onReviewLaunchOptions,
  onRecentProjectSelect,
}: HomeViewProps) {
  return (
    <div className="space-y-6">
      <HomeProjectSection project={dashboard.project} onReviewLaunchOptions={onReviewLaunchOptions} />
      <HomeCurrentContextSection currentContext={dashboard.currentContext} />
      <HomeQuickLaunchSection
        dashboard={dashboard}
        pending={launchPending}
        error={launchError}
        onQuickLaunch={onQuickLaunch}
      />
      <HomeRecentProjectsSection projects={dashboard.recentProjects} onSelect={onRecentProjectSelect} />
    </div>
  );
}

function HomeRecentProjectsSection({
  projects,
  onSelect,
}: {
  projects: RecentProjectState[];
  onSelect?: (project: RecentProjectState) => void;
}) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="home-recent-projects-heading">
      <CardContent className="space-y-3 p-5">
        <div>
          <h2 id="home-recent-projects-heading" className="font-semibold">Recent projects</h2>
          <p className="mt-1 text-sm text-muted-foreground">Review a project before launching it.</p>
        </div>
        {projects.length === 0 ? (
          <p className="text-sm text-muted-foreground">No recent projects yet.</p>
        ) : (
          <ul className="divide-y divide-border border-y border-border" aria-label="Recent projects">
            {projects.map((project) => {
              const contextName = project.contextName ?? project.contextId;
              return (
                <li key={`${project.project.path}:${project.contextId}`}>
                  <Button
                    type="button"
                    variant="ghost"
                    className="h-auto w-full justify-between gap-4 px-0 py-3 text-left normal-case tracking-normal"
                    disabled={onSelect === undefined}
                    onClick={() => onSelect?.(project)}
                  >
                    <span className="min-w-0">
                      <span className="block truncate text-sm font-semibold" title={project.project.name}>{project.project.name}</span>
                      <span className="mt-1 block truncate font-mono text-xs text-muted-foreground" title={project.project.path}>{project.project.path}</span>
                    </span>
                    <span className="shrink-0 text-right text-xs text-muted-foreground">
                      <span className="block">{contextName}</span>
                      <span className="mt-1 block">{formatRecentProjectTime(project.lastLaunchedAt)}</span>
                    </span>
                  </Button>
                </li>
              );
            })}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function formatRecentProjectTime(value: string): string {
  const time = new Date(value);
  return Number.isNaN(time.getTime()) ? "Unknown" : time.toLocaleString();
}

function HomeProjectSection({
  project,
  onReviewLaunchOptions,
}: {
  project: HomeDashboardState["project"];
  onReviewLaunchOptions: () => void;
}) {
  return (
    <Card as="section" hierarchy="primary" className="border border-border py-0" aria-labelledby="home-project-heading">
      <CardContent className="space-y-4 p-5">
        <div className="min-w-0">
          <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Selected project</p>
          <h2 id="home-project-heading" className="mt-1 truncate text-xl font-semibold" title={project.name}>{project.name}</h2>
          <p className="mt-1 truncate font-mono text-sm text-muted-foreground" title={project.path}>{project.path}</p>
        </div>
        <dl className="grid grid-cols-2 gap-4 border-t border-border pt-4 text-sm">
          <HomeUnavailableMetadata label="Git branch" />
          <HomeUnavailableMetadata label="Last opened" />
        </dl>
        <div className="border-t border-border pt-4">
          <Button type="button" variant="outline" size="sm" onClick={onReviewLaunchOptions}>Review launch options</Button>
        </div>
      </CardContent>
    </Card>
  );
}

function HomeUnavailableMetadata({ label }: { label: string }) {
  return (
    <div>
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="mt-1 font-medium text-muted-foreground">Unavailable</dd>
    </div>
  );
}

function HomeCurrentContextSection({ currentContext }: { currentContext: HomeDashboardState["currentContext"] }) {
  if (currentContext === undefined) {
    return (
      <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="home-context-heading">
        <CardContent className="p-5">
          <h2 id="home-context-heading" className="font-semibold">Current context</h2>
          <p className="mt-1 text-sm text-muted-foreground">Choose a context in the selector before launching this project.</p>
        </CardContent>
      </Card>
    );
  }

  const status = currentContext.confidence.status;
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="home-context-heading">
      <CardContent className="space-y-3 p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">Current context</p>
            <h2 id="home-context-heading" className="mt-1 truncate text-lg font-semibold">{currentContext.name}</h2>
            <p className="mt-1 text-sm text-muted-foreground">Coding tool: {currentContext.tool.name}</p>
          </div>
          <StatusIndicator status={homeStatus(status)} />
        </div>
        <p className="border-t border-border pt-3 text-sm text-muted-foreground">{homeConfidenceSummary(status)}</p>
      </CardContent>
    </Card>
  );
}

function HomeQuickLaunchSection({
  dashboard,
  pending,
  error,
  onQuickLaunch,
}: {
  dashboard: HomeDashboardState;
  pending: boolean;
  error?: DisplayError;
  onQuickLaunch: () => void;
}) {
  const context = dashboard.currentContext;
  const canLaunch = context !== undefined && context.confidence.status !== "blocked";
  const label = context ? `Launch ${context.name}` : "Launch context";

  return (
    <Card as="section" hierarchy="tertiary" className="py-0" aria-labelledby="home-quick-launch-heading">
      <CardContent className="space-y-3 p-5">
        <div>
          <h2 id="home-quick-launch-heading" className="font-semibold">Quick launch</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            {context ? `Open ${dashboard.project.name} with ${context.name}.` : "A context must be selected before launch."}
          </p>
        </div>
        {error ? <p className="text-sm text-destructive" role="alert">{error.message}</p> : null}
        <Button type="button" disabled={!canLaunch || pending} onClick={onQuickLaunch}>
          {pending ? `Launching ${context?.name ?? "context"}...` : label}
        </Button>
      </CardContent>
    </Card>
  );
}

function homeStatus(status: "ready" | "needs_attention" | "blocked") {
  return status;
}

function homeConfidenceSummary(status: "ready" | "needs_attention" | "blocked"): string {
  switch (status) {
    case "ready":
      return "All systems required for this context are ready.";
    case "needs_attention":
      return "This context can launch, but review its setup before continuing.";
    case "blocked":
      return "This context is blocked until its required setup is resolved.";
  }
}

export { HomeView, formatRecentProjectTime, homeConfidenceSummary };
