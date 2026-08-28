import type { ReactNode } from "react";

import type { ProjectState } from "../../lib/devctx-api";
import { appRoutes, type AppRoute } from "./routes.js";

interface AppShellProps {
  activeRoute: AppRoute;
  onNavigate: (route: AppRoute) => void;
  currentProject?: ProjectState;
  statusBar?: ReactNode;
  children: ReactNode;
}

function AppShell({ activeRoute, onNavigate, currentProject, statusBar, children }: AppShellProps) {
  return (
    <div className="flex min-h-screen min-w-0 flex-col bg-background text-foreground" data-app-shell>
      <div className="flex min-h-0 flex-1">
      <aside className="flex w-48 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground sm:w-56">
        <div className="border-b border-sidebar-border px-4 py-5 sm:px-6">
          <h1 className="text-base font-semibold">Dev Context</h1>
        </div>
        <nav className="flex flex-1 flex-col gap-1 p-3" aria-label="Primary navigation">
          {appRoutes.map((route) => (
            <button
              key={route.id}
              type="button"
              className={`min-w-0 px-3 py-2 text-left text-sm font-medium transition-colors ${
                activeRoute === route.id
                  ? "bg-sidebar-primary text-sidebar-primary-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent"
              }`}
              aria-current={activeRoute === route.id ? "page" : undefined}
              onClick={() => onNavigate(route.id)}
            >
              {route.label}
            </button>
          ))}
        </nav>
        <CurrentProjectSummary project={currentProject} />
      </aside>
      <main className="min-w-0 flex-1 overflow-x-hidden">
        <div className="mx-auto w-full max-w-6xl px-4 py-6 sm:px-6 sm:py-8 lg:px-8">{children}</div>
      </main>
      </div>
      {statusBar}
    </div>
  );
}

function CurrentProjectSummary({ project }: { project?: ProjectState }) {
  if (project === undefined) {
    return (
      <div className="border-t border-sidebar-border px-4 py-4 text-xs text-sidebar-foreground/70">
        Current project unavailable
      </div>
    );
  }

  return (
    <section className="min-w-0 border-t border-sidebar-border px-4 py-4" aria-labelledby="shell-current-project-title">
      <p id="shell-current-project-title" className="text-xs font-semibold tracking-wide text-sidebar-foreground/70 uppercase">
        Current project
      </p>
      <p className="mt-1 truncate text-sm font-medium" title={project.name}>{project.name}</p>
      <p className="mt-1 truncate font-mono text-xs text-sidebar-foreground/70" title={project.path}>{project.path}</p>
    </section>
  );
}

export { AppShell, CurrentProjectSummary };
export type { AppShellProps };
