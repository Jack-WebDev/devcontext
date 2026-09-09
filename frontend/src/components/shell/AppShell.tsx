import {
	Command,
	CirclePlay,
	Clock3,
	FolderKanban,
	House,
	Layers3,
	Plus,
	Settings,
} from "lucide-react";
import type { ReactNode } from "react";

import type { ProjectState } from "../../lib/devctx-api";
import { type AppRoute, appRouteDefinition, appRoutes } from "./routes.js";

interface AppShellProps {
	activeRoute: AppRoute;
	onNavigate: (route: AppRoute) => void;
	isFirstRun?: boolean;
	currentProject?: ProjectState;
	statusBar?: ReactNode;
	onOpenCommandPalette?: () => void;
	children: ReactNode;
}

function AppShell({
	activeRoute,
	onNavigate,
	isFirstRun = false,
	currentProject,
	statusBar,
	onOpenCommandPalette,
	children,
}: AppShellProps) {
	return (
		<div
			className="app-shell grid h-full grid-rows-[minmax(0,1fr)_48px] overflow-hidden text-foreground"
			data-app-shell
		>
			<div className="app-shell-grid grid min-h-0 overflow-hidden">
				<aside className="app-sidebar flex min-h-0 flex-col overflow-hidden border-r border-sidebar-border/80 bg-sidebar text-sidebar-foreground">
					<div className="px-6 pt-7 pb-7">
						<h1 className="flex items-center gap-3 text-[15px] font-bold tracking-[-0.025em] text-sidebar-foreground">
							<span className="grid size-8 place-items-center rounded-[10px] bg-primary text-primary-foreground shadow-sm">
								<Layers3 className="size-5" />
							</span>
							Dev Context
						</h1>
					</div>
					<nav
						className="flex flex-col gap-1 px-3"
						aria-label="Primary navigation"
					>
						<p className="px-3 pb-1.5 text-[10px] font-bold tracking-[0.08em] text-sidebar-foreground/45 uppercase">
							Workspace
						</p>
						{appRoutes
							.filter((route) => route.id !== "settings")
							.map((route) => (
								<button
									key={route.id}
									type="button"
									className="flex h-[36px] min-w-0 items-center gap-3 rounded-lg px-3 text-left text-[13px] font-medium text-sidebar-foreground/70 transition-colors duration-150 hover:bg-sidebar-accent/75 hover:text-sidebar-accent-foreground focus-visible:outline-2 focus-visible:outline-sidebar-ring focus-visible:outline-offset-2 data-[active=true]:bg-sidebar-accent data-[active=true]:font-semibold data-[active=true]:text-sidebar-accent-foreground motion-reduce:transition-none"
									data-active={activeRoute === route.id}
									aria-current={activeRoute === route.id ? "page" : undefined}
									onClick={() => onNavigate(route.id)}
								>
									<NavIcon route={route.id} />
									{route.label}
								</button>
							))}
					</nav>
					<div className="mx-4 mt-5 border-t border-sidebar-border/80 pt-4">
						<button
							type="button"
							className="flex h-[36px] w-full items-center gap-3 rounded-lg px-3 text-left text-[13px] font-medium text-sidebar-foreground/70 transition-colors duration-150 hover:bg-sidebar-accent/75 hover:text-sidebar-accent-foreground focus-visible:outline-2 focus-visible:outline-sidebar-ring focus-visible:outline-offset-2 data-[active=true]:bg-sidebar-accent data-[active=true]:font-semibold data-[active=true]:text-sidebar-accent-foreground motion-reduce:transition-none"
							data-active={activeRoute === "settings"}
							onClick={() => onNavigate("settings")}
						>
							<Settings className="size-[18px]" />
							Settings
						</button>
					</div>
					<div className="flex-1" />
					{isFirstRun || !currentProject ? null : (
						<CurrentProjectSummary project={currentProject} />
					)}
				</aside>
				<main className="min-h-0 min-w-0 overflow-x-hidden overflow-y-auto">
					<div className="sticky top-0 z-10">
						<div className="app-toolbar">
							<span className="app-toolbar-title">
								{appRouteDefinition(activeRoute).label}
							</span>
							{onOpenCommandPalette ? (
								<button
									type="button"
									className="app-toolbar-command"
									onClick={onOpenCommandPalette}
									aria-label="Open command palette"
								>
									<Command className="size-3.5" />
									Quick switch <kbd>⌘ K</kbd>
								</button>
							) : null}
						</div>
					</div>
					<div className="app-page-container">{children}</div>
				</main>
			</div>
			{statusBar}
		</div>
	);
}

function NavIcon({ route }: { route: AppRoute }) {
	const className = "size-4 shrink-0";
	switch (route) {
		case "home":
			return <House className={className} />;
		case "contexts":
			return <Layers3 className={className} />;
		case "projects":
			return <FolderKanban className={className} />;
		case "running":
			return <CirclePlay className={className} />;
		case "history":
			return <Clock3 className={className} />;
		case "settings":
			return <Settings className={className} />;
		case "diagnostics":
			return <Plus className={className} />;
	}
}

function CurrentProjectSummary({ project }: { project: ProjectState }) {
	return (
		<section
			className="min-w-0 border-t border-sidebar-border/80 px-5 py-4"
			aria-labelledby="shell-current-project-title"
		>
			<p
				id="shell-current-project-title"
				className="text-[10px] font-semibold tracking-[0.08em] text-sidebar-foreground/55 uppercase"
			>
				Current project
			</p>
			<div className="mt-2 flex items-center gap-2">
				<span
					className="size-2 shrink-0 rounded-full bg-success shadow-[0_0_0_3px_color-mix(in_srgb,var(--success)_14%,transparent)]"
					aria-hidden="true"
				/>
				<p className="truncate text-sm font-semibold" title={project.name}>
					{project.name}
				</p>
			</div>
			<p
				className="mt-1 truncate font-mono text-xs text-sidebar-foreground/70"
				title={project.path}
			>
				{project.path}
			</p>
		</section>
	);
}

export type { AppShellProps };
export { AppShell, CurrentProjectSummary };
