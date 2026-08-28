import { GuiErrorNotice } from "../selector/GuiErrorNotice";
import { SelectorView } from "../selector/SelectorView";
import { HomeView } from "../home/HomeView";
import { ProjectsView } from "../projects/ProjectsView";
import { ContextsView } from "../contexts/ContextsView";
import { HistoryView } from "../history/HistoryView";
import { RunningView } from "../running/RunningView";
import { devContextApi } from "../../lib/devctx-api";
import { devContextWindow } from "../../lib/devctx-window";
import type {
  ApiResult,
  ContextListItem,
  CreateContextResult,
  DisplayError,
  HistoryState,
  HomeDashboardState,
  LaunchState,
  ProjectListItem,
  ProjectsState,
  RecentProjectState,
  RunningEnvironmentsState,
  SettingsState,
} from "../../lib/devctx-api";
import { appRouteDefinition, type AppRoute } from "../shell/routes";
import { notifyCodingToolLaunched } from "../notifications/notifications";
import type { LoadState } from "./load-state";

export function HistoryContent({ history }: { history: LoadState<HistoryState> }) {
  if (history.status === "loading") return <LoadingMessage>Loading history...</LoadingMessage>;
  if (history.status === "error") return <GuiErrorNotice error={history.error} />;
  return <HistoryView entries={history.data.entries} />;
}

export function RunningContent({ running }: { running: LoadState<RunningEnvironmentsState> }) {
  if (running.status === "loading") return <LoadingMessage>Refreshing active environments...</LoadingMessage>;
  if (running.status === "error") return <GuiErrorNotice error={running.error} />;
  return <RunningView environments={running.data.environments} />;
}

export function ContextsContent({ contexts, onSelect, onNew }: {
  contexts: LoadState<ContextListItem[]>;
  onSelect: (id: string) => void;
  onNew: () => void;
}) {
  if (contexts.status === "loading") return <LoadingMessage>Loading contexts...</LoadingMessage>;
  if (contexts.status === "error") return <GuiErrorNotice error={contexts.error} />;
  return <ContextsView contexts={contexts.data} onSelect={onSelect} onNew={onNew} />;
}

export function ProjectsContent({ projects, launchingProjectPath, errorProjectPath, launchError, onLaunch, onChangeContext, onOpenFolder }: {
  projects: LoadState<ProjectsState>;
  launchingProjectPath?: string;
  errorProjectPath?: string;
  launchError?: DisplayError;
  onLaunch: (project: ProjectListItem) => void;
  onChangeContext?: (project: ProjectListItem) => void;
  onOpenFolder: (project: ProjectListItem) => void;
}) {
  if (projects.status === "loading") return <LoadingMessage>Loading projects...</LoadingMessage>;
  if (projects.status === "error") return <GuiErrorNotice error={projects.error} />;
  return <ProjectsView projects={projects.data.projects} launchingProjectPath={launchingProjectPath} errorProjectPath={errorProjectPath} launchError={launchError?.message} onLaunch={onLaunch} onChangeContext={onChangeContext} onOpenFolder={onOpenFolder} />;
}

export function HomeDashboardContent({ dashboard, recentProjects, launchPending, launchError, onQuickLaunch, onReviewLaunchOptions, onRecentProjectSelect }: {
  dashboard: LoadState<HomeDashboardState>;
  recentProjects: LoadState<RecentProjectState[]>;
  launchPending: boolean;
  launchError?: DisplayError;
  onQuickLaunch: () => void;
  onReviewLaunchOptions: () => void;
  onRecentProjectSelect: (project: RecentProjectState) => void;
}) {
  if (dashboard.status === "loading") return <LoadingMessage>Loading Home dashboard...</LoadingMessage>;
  if (dashboard.status === "error") return <GuiErrorNotice error={dashboard.error} />;
  const projects = recentProjects.status === "loaded" ? recentProjects.data : [];
  return <HomeView dashboard={{ ...dashboard.data, recentProjects: projects }} launchPending={launchPending} launchError={launchError} onQuickLaunch={onQuickLaunch} onReviewLaunchOptions={onReviewLaunchOptions} onRecentProjectSelect={onRecentProjectSelect} />;
}

export function SelectorContent({ launchState, onCreateContext, onRunDiagnostics, settings }: {
  launchState: LoadState<LaunchState>;
  onCreateContext: (contextId: string, importProviderIds?: string[]) => Promise<ApiResult<CreateContextResult>>;
  onRunDiagnostics: () => void;
  settings?: SettingsState;
}) {
  if (launchState.status === "loading") return <LoadingMessage>Loading selector...</LoadingMessage>;
  if (launchState.status === "error") return <GuiErrorNotice error={launchState.error} />;
  return <SelectorView launchState={launchState.data} onBindProject={devContextApi.bindProject} onPreflightLaunchProject={devContextApi.preflightLaunchProject} onLaunchProject={devContextApi.launchProject} onCancel={devContextWindow.closeSelector} onCreatePersonalContext={(providerIds) => onCreateContext("personal", providerIds)} onCreateCompanyContext={(providerIds) => onCreateContext("company", providerIds)} onRunDiagnostics={onRunDiagnostics} onCodingToolLaunched={notifyLaunch} launchSuccessCloseBehavior={settings?.closeAfterLaunch ? "close_selector" : "keep_open"} showLaunchVerification={settings?.launchVerification ?? true} />;
}

export function PlaceholderScreen({ route }: { route: AppRoute }) {
  const definition = appRouteDefinition(route);
  return (
    <section className="max-w-2xl space-y-2" aria-labelledby={`${route}-heading`}>
      <p className="text-sm text-muted-foreground">{definition.label}</p>
      <h2 id={`${route}-heading`} className="text-2xl font-semibold">{definition.label}</h2>
      <p className="text-sm text-muted-foreground">This section will be available as its supporting API is added.</p>
    </section>
  );
}

function LoadingMessage({ children }: { children: string }) {
  return <p className="text-sm text-muted-foreground">{children}</p>;
}

function notifyLaunch(result: { project: { name: string }; context: { name: string; tool: { name: string } } }) {
  notifyCodingToolLaunched({ projectName: result.project.name, contextName: result.context.name, toolName: result.context.tool.name });
}
