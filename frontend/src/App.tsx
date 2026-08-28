import { useEffect, useState } from "react";

import { GuiErrorNotice } from "./components/selector/GuiErrorNotice";
import { SelectorView } from "./components/selector/SelectorView";
import { createOnboardingContextAndRefresh } from "./components/selector/onboarding-action";
import { HomeView } from "./components/home/HomeView";
import { RecentProjectConfirmationDialog } from "./components/home/RecentProjectConfirmationDialog";
import { ProjectsView } from "./components/projects/ProjectsView";
import { ProjectContextChangeDialog } from "./components/projects/ProjectContextChangeDialog";
import { ContextsView } from "./components/contexts/ContextsView";
import { ContextDetailsDrawer, CreateContextDialog } from "./components/contexts/ContextManagement";
import { AppShell } from "./components/shell/AppShell";
import { appRouteDefinition, appRouteFromHash, type AppRoute } from "./components/shell/routes";
import {
  devContextApi,
  type ApiResult,
  type CreateContextResult,
  type ContextListItem,
  type DisplayError,
  type HomeDashboardState,
  type LaunchState,
  type ProjectListItem,
  type ProjectsState,
  type RecentProjectState,
} from "./lib/devctx-api";
import { devContextWindow } from "./lib/devctx-window";

type LaunchStateLoad =
  | { status: "loading" }
  | { status: "loaded"; data: LaunchState }
  | { status: "error"; error: DisplayError };

type HomeDashboardLoad =
  | { status: "loading" }
  | { status: "loaded"; data: HomeDashboardState }
  | { status: "error"; error: DisplayError };

type RecentProjectsLoad =
  | { status: "loading" }
  | { status: "loaded"; data: RecentProjectState[] }
  | { status: "error"; error: DisplayError };

type ContextsLoad =
  | { status: "loading" }
  | { status: "loaded"; data: ContextListItem[] }
  | { status: "error"; error: DisplayError };

type ProjectsLoad =
  | { status: "loading" }
  | { status: "loaded"; data: ProjectsState }
  | { status: "error"; error: DisplayError };

function App() {
  const [launchState, setLaunchState] = useState<LaunchStateLoad>({
    status: "loading",
  });
  const [homeDashboard, setHomeDashboard] = useState<HomeDashboardLoad>({status: "loading"});
  const [recentProjects, setRecentProjects] = useState<RecentProjectsLoad>({status: "loading"});
  const [contexts, setContexts] = useState<ContextsLoad>({status: "loading"});
  const [projects, setProjects] = useState<ProjectsLoad>({status: "loading"});
  const [contextDetailsID, setContextDetailsID] = useState<string>();
  const [creatingContext, setCreatingContext] = useState(false);
  const [homeLaunchPending, setHomeLaunchPending] = useState(false);
  const [homeLaunchError, setHomeLaunchError] = useState<DisplayError | undefined>(undefined);
  const [recentProjectToLaunch, setRecentProjectToLaunch] = useState<RecentProjectState | undefined>(undefined);
  const [recentProjectLaunchPending, setRecentProjectLaunchPending] = useState(false);
  const [recentProjectLaunchError, setRecentProjectLaunchError] = useState<DisplayError | undefined>(undefined);
  const [projectLaunchPath, setProjectLaunchPath] = useState<string>();
  const [projectErrorPath, setProjectErrorPath] = useState<string>();
  const [projectLaunchError, setProjectLaunchError] = useState<DisplayError>();
  const [projectContextChange, setProjectContextChange] = useState<ProjectListItem>();
  const [projectContextChangePending, setProjectContextChangePending] = useState(false);
  const [projectContextChangeError, setProjectContextChangeError] = useState<DisplayError>();
  const [activeRoute, setActiveRoute] = useState<AppRoute>(() => appRouteFromHash(window.location.hash));

  useEffect(() => {
    let active = true;

    devContextApi.getLaunchState().then((result) => {
      if (!active) {
        return;
      }

      if (result.ok) {
        setLaunchState({ status: "loaded", data: result.data });
        return;
      }

      setLaunchState({ status: "error", error: result.error });
    });

    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    void refreshRecentProjects();
  }, []);

  useEffect(() => {
    void refreshContexts();
  }, []);

  useEffect(() => {
    void refreshProjects();
  }, []);

  useEffect(() => {
    let active = true;
    devContextApi.getHomeDashboard().then((result) => {
      if (!active) {
        return;
      }
      setHomeDashboard(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error});
    });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    function syncRouteFromHash() {
      setActiveRoute(appRouteFromHash(window.location.hash));
    }

    window.addEventListener("hashchange", syncRouteFromHash);
    return () => window.removeEventListener("hashchange", syncRouteFromHash);
  }, []);

  function handleNavigate(route: AppRoute) {
    if (route !== activeRoute) {
      window.location.hash = route;
    }
  }

  async function handleCreateContext(
    contextId: string,
    importProviderIds: string[] = [],
  ): Promise<ApiResult<CreateContextResult>> {
    const result = await createOnboardingContextAndRefresh({
      contextId,
      importProviderIds,
      createContext: (requestedContextId, requestedImportProviderIds) =>
        devContextApi.createContext({ contextId: requestedContextId, importProviderIds: requestedImportProviderIds }),
      getLaunchState: () => devContextApi.getLaunchState(),
    });
    if (result.ok) {
      setLaunchState({ status: "loaded", data: result.launchState });
      void refreshHomeDashboard();
      void refreshRecentProjects();
      void refreshContexts();
      void refreshProjects();
      return { ok: true, data: result.created };
    }

    return { ok: false, error: result.error };
  }

  async function refreshHomeDashboard() {
    const result = await devContextApi.getHomeDashboard();
    setHomeDashboard(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error});
  }

  async function refreshRecentProjects() {
    const result = await devContextApi.getRecentProjects();
    setRecentProjects(result.ok ? {status: "loaded", data: result.data.projects} : {status: "error", error: result.error});
  }

  async function refreshContexts() {
    const result = await devContextApi.getContexts();
    setContexts(result.ok ? {status: "loaded", data: result.data.contexts} : {status: "error", error: result.error});
  }

  async function refreshProjects() {
    const result = await devContextApi.getProjects();
    setProjects(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error});
  }

  async function handleHomeQuickLaunch() {
    if (homeDashboard.status !== "loaded" || homeDashboard.data.currentContext === undefined || homeLaunchPending) {
      return;
    }

    const {project, currentContext} = homeDashboard.data;
    setHomeLaunchPending(true);
    setHomeLaunchError(undefined);
    try {
      const request = {projectPath: project.path, contextId: currentContext.id};
      const preflight = await devContextApi.preflightLaunchProject(request);
      if (!preflight.ok) {
        setHomeLaunchError(preflight.error);
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setHomeLaunchError(launch.error);
        return;
      }
      await refreshHomeDashboard();
      await refreshRecentProjects();
      await refreshProjects();
    } finally {
      setHomeLaunchPending(false);
    }
  }

  function handleRecentProjectSelect(project: RecentProjectState) {
    setRecentProjectToLaunch(project);
    setRecentProjectLaunchError(undefined);
  }

  function handleRecentProjectCancel() {
    if (!recentProjectLaunchPending) {
      setRecentProjectToLaunch(undefined);
      setRecentProjectLaunchError(undefined);
    }
  }

  async function handleRecentProjectConfirm() {
    if (recentProjectToLaunch === undefined || recentProjectLaunchPending) {
      return;
    }

    setRecentProjectLaunchPending(true);
    setRecentProjectLaunchError(undefined);
    try {
      const request = {projectPath: recentProjectToLaunch.project.path, contextId: recentProjectToLaunch.contextId};
      const preflight = await devContextApi.preflightLaunchProject(request);
      if (!preflight.ok) {
        setRecentProjectLaunchError(preflight.error);
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setRecentProjectLaunchError(launch.error);
        return;
      }
      setRecentProjectToLaunch(undefined);
      await refreshRecentProjects();
      await refreshProjects();
    } finally {
      setRecentProjectLaunchPending(false);
    }
  }

  function handleReviewLaunchOptions() {
    document.getElementById("context-selector")?.scrollIntoView({behavior: "smooth", block: "start"});
  }

  async function handleProjectLaunch(project: ProjectListItem) {
    if (project.contextId === undefined || projectLaunchPath !== undefined) {
      return;
    }

    setProjectLaunchPath(project.project.path);
    setProjectErrorPath(project.project.path);
    setProjectLaunchError(undefined);
    try {
      const request = {projectPath: project.project.path, contextId: project.contextId};
      const preflight = await devContextApi.preflightLaunchProject(request);
      if (!preflight.ok) {
        setProjectLaunchError(preflight.error);
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setProjectLaunchError(launch.error);
        return;
      }
      await Promise.all([refreshHomeDashboard(), refreshRecentProjects(), refreshProjects()]);
    } finally {
      setProjectLaunchPath(undefined);
    }
  }

  function handleProjectChangeContext(project: ProjectListItem) {
    setProjectContextChange(project);
    setProjectContextChangeError(undefined);
  }

  async function handleProjectContextChange(contextId: string) {
    if (projectContextChange === undefined || projectContextChangePending) {
      return;
    }

    setProjectContextChangePending(true);
    setProjectContextChangeError(undefined);
    try {
      const result = await devContextApi.bindProject({projectPath: projectContextChange.project.path, contextId});
      if (!result.ok) {
        setProjectContextChangeError(result.error);
        return;
      }
      setProjectContextChange(undefined);
      await Promise.all([refreshProjects(), refreshHomeDashboard(), refreshContexts()]);
    } finally {
      setProjectContextChangePending(false);
    }
  }

  function handleProjectOpenFolder(project: ProjectListItem) {
    window.open(new URL(project.project.path, "file://").href, "_blank", "noopener,noreferrer");
  }

  return (
    <AppShell
      activeRoute={activeRoute}
      onNavigate={handleNavigate}
      currentProject={launchState.status === "loaded" ? launchState.data.project : undefined}
    >
      {activeRoute === "home" ? (
        <section aria-labelledby="home-heading" className="space-y-8">
          <div>
            <p className="text-sm text-muted-foreground">{appRouteDefinition(activeRoute).label}</p>
            <h2 id="home-heading" className="text-2xl font-semibold">Home</h2>
          </div>
          {renderHomeDashboard(
            homeDashboard,
            recentProjects,
            homeLaunchPending,
            homeLaunchError,
            handleHomeQuickLaunch,
            handleReviewLaunchOptions,
            handleRecentProjectSelect,
          )}
          {recentProjectToLaunch ? (
            <RecentProjectConfirmationDialog
              project={recentProjectToLaunch}
              launchPending={recentProjectLaunchPending}
              error={recentProjectLaunchError}
              onCancel={handleRecentProjectCancel}
              onConfirm={() => void handleRecentProjectConfirm()}
            />
          ) : null}
          <section id="context-selector" aria-labelledby="context-selector-heading" className="space-y-6">
            <div>
              <p className="text-sm text-muted-foreground">Launch options</p>
              <h2 id="context-selector-heading" className="text-xl font-semibold">Context selector</h2>
            </div>
          {renderSelectorContent(launchState, handleCreateContext)}
          </section>
        </section>
      ) : activeRoute === "contexts" ? (
        <>{renderContexts(contexts, setContextDetailsID, () => setCreatingContext(true))}{contextDetailsID ? <ContextDetailsDrawer contextId={contextDetailsID} onClose={() => setContextDetailsID(undefined)} load={(contextId) => devContextApi.getContextDetails({contextId})}/> : null}{creatingContext && contexts.status === "loaded" ? <CreateContextDialog contexts={contexts.data} onClose={() => setCreatingContext(false)} create={async (request) => { const result = await devContextApi.createContext(request); if (result.ok) { await refreshContexts(); setCreatingContext(false); } return result; }}/> : null}</>
      ) : activeRoute === "projects" ? (
        <>
          {renderProjects(projects, projectLaunchPath, projectErrorPath, projectLaunchError, handleProjectLaunch, contexts.status === "loaded" ? handleProjectChangeContext : undefined, handleProjectOpenFolder)}
          {projectContextChange && contexts.status === "loaded" ? (
            <ProjectContextChangeDialog
              project={projectContextChange}
              contexts={contexts.data.map((item) => item.context)}
              pending={projectContextChangePending}
              error={projectContextChangeError}
              onCancel={() => !projectContextChangePending && setProjectContextChange(undefined)}
              onConfirm={(contextId) => void handleProjectContextChange(contextId)}
            />
          ) : null}
        </>
      ) : (
        <PlaceholderScreen route={activeRoute} />
      )}
    </AppShell>
  );
}

function renderContexts(contexts: ContextsLoad, onSelect: (id: string) => void, onNew: () => void) {
  if (contexts.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading contexts...</p>;
  }
  if (contexts.status === "error") {
    return <GuiErrorNotice error={contexts.error} />;
  }
  return <ContextsView contexts={contexts.data} onSelect={onSelect} onNew={onNew} />;
}

function renderProjects(
  projects: ProjectsLoad,
  launchingProjectPath: string | undefined,
  errorProjectPath: string | undefined,
  launchError: DisplayError | undefined,
  onLaunch: (project: ProjectListItem) => void,
  onChangeContext: ((project: ProjectListItem) => void) | undefined,
  onOpenFolder: (project: ProjectListItem) => void,
) {
  if (projects.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading projects...</p>;
  }
  if (projects.status === "error") {
    return <GuiErrorNotice error={projects.error} />;
  }
  return <ProjectsView projects={projects.data.projects} launchingProjectPath={launchingProjectPath} errorProjectPath={errorProjectPath} launchError={launchError?.message} onLaunch={onLaunch} onChangeContext={onChangeContext} onOpenFolder={onOpenFolder} />;
}

function renderHomeDashboard(
  dashboard: HomeDashboardLoad,
  recentProjects: RecentProjectsLoad,
  launchPending: boolean,
  launchError: DisplayError | undefined,
  onQuickLaunch: () => void,
  onReviewLaunchOptions: () => void,
  onRecentProjectSelect: (project: RecentProjectState) => void,
) {
  if (dashboard.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading Home dashboard...</p>;
  }
  if (dashboard.status === "error") {
    return <GuiErrorNotice error={dashboard.error} />;
  }
  const projects = recentProjects.status === "loaded" ? recentProjects.data : [];
  return (
    <HomeView
      dashboard={{...dashboard.data, recentProjects: projects}}
      launchPending={launchPending}
      launchError={launchError}
      onQuickLaunch={onQuickLaunch}
      onReviewLaunchOptions={onReviewLaunchOptions}
      onRecentProjectSelect={onRecentProjectSelect}
    />
  );
}

function PlaceholderScreen({ route }: { route: AppRoute }) {
  const definition = appRouteDefinition(route);
  return (
    <section className="max-w-2xl space-y-2" aria-labelledby={`${route}-heading`}>
      <p className="text-sm text-muted-foreground">{definition.label}</p>
      <h2 id={`${route}-heading`} className="text-2xl font-semibold">{definition.label}</h2>
      <p className="text-sm text-muted-foreground">This section will be available as its supporting API is added.</p>
    </section>
  );
}

function renderSelectorContent(
  launchState: LaunchStateLoad,
  onCreateContext: (
    contextId: string,
    importProviderIds?: string[],
  ) => Promise<ApiResult<CreateContextResult>>,
) {
  if (launchState.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading selector...</p>;
  }

  if (launchState.status === "error") {
    return <GuiErrorNotice error={launchState.error} />;
  }

  return (
    <SelectorView
      launchState={launchState.data}
      onBindProject={(request) => devContextApi.bindProject(request)}
      onPreflightLaunchProject={(request) => devContextApi.preflightLaunchProject(request)}
      onLaunchProject={(request) => devContextApi.launchProject(request)}
      onCancel={() => devContextWindow.closeSelector()}
      onCreatePersonalContext={(importProviderIds) => onCreateContext("personal", importProviderIds)}
      onCreateCompanyContext={(importProviderIds) => onCreateContext("company", importProviderIds)}
    />
  );
}

export default App;
