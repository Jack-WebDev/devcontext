import type { ReactNode } from "react";
import { CirclePlay, Clock3, FolderKanban, House, Layers3, Plus, Settings, ShieldCheck } from "lucide-react";

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
    <div className="flex min-h-screen min-w-0 flex-col bg-[#faf9f7] text-foreground" data-app-shell>
      <div className="flex min-h-0 flex-1">
      <aside className="flex w-56 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
        <div className="px-6 py-9">
          <h1 className="flex items-center gap-3 text-lg font-semibold tracking-tight"><span className="grid size-8 place-items-center rounded-md bg-[#566d5a] text-white"><Layers3 className="size-5" /></span>Dev Context</h1>
        </div>
        <nav className="flex flex-1 flex-col gap-1 px-4" aria-label="Primary navigation">
          {appRoutes.map((route) => (
            <button
              key={route.id}
              type="button"
              className={`flex min-w-0 items-center gap-3 rounded-md px-3 py-2.5 text-left text-sm font-medium transition-colors ${
                activeRoute === route.id
                  ? "bg-sidebar-accent text-sidebar-foreground"
                  : "text-sidebar-foreground hover:bg-sidebar-accent"
              }`}
              aria-current={activeRoute === route.id ? "page" : undefined}
              onClick={() => onNavigate(route.id)}
            >
              <NavIcon route={route.id} />
              {route.label}
            </button>
          ))}
        </nav>
        <CurrentProjectSummary project={currentProject} />
      </aside>
      <main className="min-w-0 flex-1 overflow-x-hidden">
        <div className="mx-auto w-full max-w-6xl px-6 py-8 lg:px-9">{children}</div>
      </main>
      </div>
      {statusBar}
    </div>
  );
}

function NavIcon({ route }: { route: AppRoute }) {
  const className = "size-4 shrink-0";
  switch (route) {
    case "home": return <House className={className} />;
    case "contexts": return <Layers3 className={className} />;
    case "projects": return <FolderKanban className={className} />;
    case "running": return <CirclePlay className={className} />;
    case "history": return <Clock3 className={className} />;
    case "settings": return <Settings className={className} />;
    case "trust": return <ShieldCheck className={className} />;
    case "diagnostics": return <Plus className={className} />;
  }
}

function CurrentProjectSummary({ project }: { project?: ProjectState }) {
  if (project === undefined) {
    return (
      <div className="border-t border-sidebar-border px-5 py-4 text-xs text-sidebar-foreground/70">
        Current project unavailable
      </div>
    );
  }

  return (
    <section className="min-w-0 border-t border-sidebar-border px-5 py-4" aria-labelledby="shell-current-project-title">
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
