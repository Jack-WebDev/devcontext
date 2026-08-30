import {
	CirclePlay,
	Clock3,
	FolderKanban,
	House,
	Layers3,
	Plus,
	Settings,
	ShieldCheck,
} from "lucide-react";
import type { ReactNode } from "react";

import type { ProjectState } from "../../lib/devctx-api";
import { type AppRoute, appRoutes } from "./routes.js";

interface AppShellProps {
	activeRoute: AppRoute;
	onNavigate: (route: AppRoute) => void;
	isFirstRun?: boolean;
	currentProject?: ProjectState;
	statusBar?: ReactNode;
	children: ReactNode;
}

function AppShell({
	activeRoute,
	onNavigate,
	isFirstRun = false,
	currentProject,
	statusBar,
	children,
}: AppShellProps) {
	return (
		<div
			className="app-shell grid h-screen grid-rows-[minmax(0,1fr)_48px] overflow-hidden text-foreground"
			data-app-shell
		>
			<div className="app-shell-grid grid min-h-0 overflow-hidden">
				<aside className="flex min-h-0 flex-col overflow-hidden border-r border-sidebar-border bg-[#fbfaf8] text-sidebar-foreground">
					<div className="px-7 pt-9 pb-8">
						<h1 className="flex items-center gap-3 text-lg font-semibold tracking-tight">
							<span className="grid size-8 place-items-center rounded-md bg-[#566d5a] text-white">
								<Layers3 className="size-5" />
							</span>
							Dev Context
						</h1>
					</div>
					<nav
						className="flex flex-col gap-px px-4"
						aria-label="Primary navigation"
					>
						{appRoutes
							.filter((route) => route.id !== "settings")
							.map((route) => (
								<button
									key={route.id}
									type="button"
									className="flex h-[38px] min-w-0 items-center gap-3 rounded-[7px] px-3.5 text-left text-sm font-medium text-muted-foreground transition-colors duration-150 hover:bg-[#f3f1ed] hover:text-foreground data-[active=true]:bg-[#efede9] data-[active=true]:text-foreground"
									data-active={activeRoute === route.id}
									aria-current={activeRoute === route.id ? "page" : undefined}
									onClick={() => onNavigate(route.id)}
								>
									<NavIcon route={route.id} />
									{route.label}
								</button>
							))}
					</nav>
					<div className="mx-5 mt-5 border-t border-sidebar-border pt-4">
						<button
							type="button"
							className="flex h-[38px] w-full items-center gap-3 rounded-[7px] px-3.5 text-left text-sm font-medium text-muted-foreground transition-colors duration-150 hover:bg-[#f3f1ed] hover:text-foreground data-[active=true]:bg-[#efede9] data-[active=true]:text-foreground"
							data-active={activeRoute === "settings"}
							onClick={() => onNavigate("settings")}
						>
							<Settings className="size-[18px]" />
							Settings
						</button>
					</div>
					<div className="flex-1" />
					{isFirstRun ? null : <SidebarShortcuts />}
					{currentProject ? (
						<span className="sr-only">
							Current project {currentProject.name} {currentProject.path}
						</span>
					) : null}
				</aside>
				<main className="min-h-0 min-w-0 overflow-x-hidden overflow-y-auto">
					<div className="app-page-container">{children}</div>
				</main>
			</div>
			{statusBar}
		</div>
	);
}

function SidebarShortcuts() {
	return (
		<div className="mt-auto px-3 pb-5">
			<section
				className="rounded-lg border border-border bg-[#faf9f7] p-3"
				aria-label="Keyboard shortcuts"
			>
				<p className="mb-2.5 text-[11px] font-semibold">Keyboard shortcuts</p>
				<Shortcut label="Quick Switch" keycap="⌘ K" />
			</section>
		</div>
	);
}

function Shortcut({ label, keycap }: { label: string; keycap: string }) {
	return (
		<div className="flex h-[29px] items-center justify-between text-[11px] text-muted-foreground">
			<span>{label}</span>
			<kbd className="min-w-9 rounded border border-border bg-white px-1.5 py-0.5 text-center text-[10px] text-foreground">
				{keycap}
			</kbd>
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
		case "trust":
			return <ShieldCheck className={className} />;
		case "diagnostics":
			return <Plus className={className} />;
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
		<section
			className="min-w-0 border-t border-sidebar-border px-5 py-4"
			aria-labelledby="shell-current-project-title"
		>
			<p
				id="shell-current-project-title"
				className="text-xs font-semibold tracking-wide text-sidebar-foreground/70 uppercase"
			>
				Current project
			</p>
			<p className="mt-1 truncate text-sm font-medium" title={project.name}>
				{project.name}
			</p>
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
