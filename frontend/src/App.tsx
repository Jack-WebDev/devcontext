import { useEffect, useState } from "react";

import { GuiErrorNotice } from "./components/selector/GuiErrorNotice";
import { createOnboardingContextAndRefresh } from "./components/selector/onboarding-action";
import { RecentProjectConfirmationDialog } from "./components/home/RecentProjectConfirmationDialog";
import { ProjectContextChangeDialog } from "./components/projects/ProjectContextChangeDialog";
import { DiagnosticsView } from "./components/diagnostics/DiagnosticsView";
import { RunningEnvironmentConflictDialog } from "./components/running/RunningEnvironmentConflictDialog";
import { ContextDetailsDrawer, CreateContextDialog } from "./components/contexts/ContextManagement";
import { AppShell } from "./components/shell/AppShell";
import { CommandPalette } from "./components/command-palette/CommandPalette";
import { SettingsView } from "./components/settings/SettingsView";
import { AppStatusBar } from "./components/status/AppStatusBar";
import { notifyCodingToolLaunched } from "./components/notifications/notifications";
import { launchContextActions, navigationActions } from "./components/command-palette/actions";
import { isCommandPaletteShortcut } from "./components/command-palette/shortcut";
import { appRouteDefinition, appRouteDefinitions, appRouteFromHash, type AppRoute } from "./components/shell/routes";
import {
  devContextApi,
  type ApiResult,
  type CreateContextResult,
  type DisplayError,
  type ImportContextMetadataRequest,
  type ImportContextMetadataResult,
  type ProjectListItem,
  type RecentProjectState,
  type RunningEnvironmentConflict,
  type SettingsState,
} from "./lib/devctx-api";
import { useAppData } from "./components/app/useAppData";
import { ContextsContent, HistoryContent, HomeDashboardContent, PlaceholderScreen, ProjectsContent, RunningContent, SelectorContent, TrustCenterContent } from "./components/app/AppContent";

interface PendingRunningEnvironmentLaunch { conflict: RunningEnvironmentConflict; request: {projectPath: string; contextId: string}; }

function App() {
  const [activeRoute, setActiveRoute] = useState<AppRoute>(() => appRouteFromHash(window.location.hash));
  const {
    launchState, setLaunchState, homeDashboard, recentProjects, contexts, projects, history, running, settings, trustCenter,
    refreshHomeDashboard, refreshRecentProjects, refreshContexts, refreshProjects, refreshRunningEnvironments, setSettings,
  } = useAppData(activeRoute);
  const [settingsPending, setSettingsPending] = useState(false);
  const [onboardingReplayVisible, setOnboardingReplayVisible] = useState(false);
  const [pendingRunningEnvironmentLaunch, setPendingRunningEnvironmentLaunch] = useState<PendingRunningEnvironmentLaunch>();
  const [runningEnvironmentLaunchPending, setRunningEnvironmentLaunchPending] = useState(false);
  const [runningEnvironmentLaunchError, setRunningEnvironmentLaunchError] = useState<DisplayError>();
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
  const [commandPaletteOpen, setCommandPaletteOpen] = useState(false);
  const [commandPaletteLaunchPending, setCommandPaletteLaunchPending] = useState(false);
  const [commandPaletteLaunchError, setCommandPaletteLaunchError] = useState<DisplayError>();


  useEffect(() => {
    function syncRouteFromHash() {
      setActiveRoute(appRouteFromHash(window.location.hash));
    }

    window.addEventListener("hashchange", syncRouteFromHash);
    return () => window.removeEventListener("hashchange", syncRouteFromHash);
  }, []);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (!isCommandPaletteShortcut(event)) {
        return;
      }

      event.preventDefault();
      setCommandPaletteOpen((open) => !open);
    }

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);


  function handleNavigate(route: AppRoute) {
    if (route !== "home") {
      setOnboardingReplayVisible(false);
    }
    if (route !== activeRoute) {
      window.location.hash = route;
    }
  }

  async function handleCommandPaletteLaunch(contextId: string) {
    if (launchState.status !== "loaded" || commandPaletteLaunchPending) {
      return;
    }

    const context = launchState.data.contexts.find((item) => item.id === contextId);
    if (context === undefined || context.confidence?.status === "blocked") {
      return;
    }

    setCommandPaletteLaunchPending(true);
    setCommandPaletteLaunchError(undefined);
    try {
      const request = {projectPath: launchState.data.project.path, contextId};
      const preflight = await devContextApi.preflightLaunchProject(request);
      if (!preflight.ok) {
        setCommandPaletteLaunchError(preflight.error);
        return;
      }
      if (deferRunningEnvironmentConflict(preflight.data, request)) {
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setCommandPaletteLaunchError(launch.error);
        return;
      }
      notifyLaunch(launch.data);
      await Promise.all([refreshHomeDashboard(), refreshRecentProjects(), refreshProjects(), refreshRunningEnvironments()]);
    } finally {
      setCommandPaletteLaunchPending(false);
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

  async function handleImportContextMetadata(request: ImportContextMetadataRequest): Promise<ApiResult<ImportContextMetadataResult>> {
    const result = await devContextApi.importContextMetadata(request);
    if (result.ok) {
      await refreshContexts();
    }
    return result;
  }

  async function handleSettingsChange(next: SettingsState) { if (settingsPending) return; setSettingsPending(true); const result = await devContextApi.updateSettings(next); setSettings(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error}); setSettingsPending(false); }

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
      if (deferRunningEnvironmentConflict(preflight.data, request)) {
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setHomeLaunchError(launch.error);
        return;
      }
      notifyLaunch(launch.data);
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
      if (deferRunningEnvironmentConflict(preflight.data, request)) {
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setRecentProjectLaunchError(launch.error);
        return;
      }
      notifyLaunch(launch.data);
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
      if (deferRunningEnvironmentConflict(preflight.data, request)) {
        return;
      }
      const launch = await devContextApi.launchProject(request);
      if (!launch.ok) {
        setProjectLaunchError(launch.error);
        return;
      }
      notifyLaunch(launch.data);
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

  function deferRunningEnvironmentConflict(preflight: {runningEnvironmentConflict?: RunningEnvironmentConflict}, request: {projectPath: string; contextId: string}) {
    if (preflight.runningEnvironmentConflict === undefined) {
      return false;
    }
    setRunningEnvironmentLaunchError(undefined);
    setPendingRunningEnvironmentLaunch({conflict: preflight.runningEnvironmentConflict, request});
    return true;
  }

  async function handleLaunchAnotherWindow() {
    if (pendingRunningEnvironmentLaunch === undefined || runningEnvironmentLaunchPending) {
      return;
    }
    setRunningEnvironmentLaunchPending(true);
    setRunningEnvironmentLaunchError(undefined);
    try {
      const result = await devContextApi.launchProject(pendingRunningEnvironmentLaunch.request);
      if (!result.ok) {
        setRunningEnvironmentLaunchError(result.error);
        return;
      }
      notifyLaunch(result.data);
      setPendingRunningEnvironmentLaunch(undefined);
      await Promise.all([refreshHomeDashboard(), refreshRecentProjects(), refreshProjects(), refreshRunningEnvironments()]);
    } finally {
      setRunningEnvironmentLaunchPending(false);
    }
  }

  return (
    <AppShell
      activeRoute={activeRoute}
      onNavigate={handleNavigate}
      currentProject={launchState.status === "loaded" ? launchState.data.project : undefined}
      statusBar={<AppStatusBar launchState={launchState.status === "loaded" ? launchState.data : undefined} />}
    >
      {activeRoute === "home" ? (
        <section aria-labelledby="home-heading" className="space-y-8">
          <div>
            <p className="text-sm text-muted-foreground">{appRouteDefinition(activeRoute).label}</p>
            <h2 id="home-heading" className="text-2xl font-semibold">Home</h2>
          </div>
          <HomeDashboardContent dashboard={homeDashboard} recentProjects={recentProjects} launchPending={homeLaunchPending} launchError={homeLaunchError} onQuickLaunch={handleHomeQuickLaunch} onReviewLaunchOptions={handleReviewLaunchOptions} onRecentProjectSelect={handleRecentProjectSelect} />
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
          <SelectorContent launchState={launchState} onCreateContext={handleCreateContext} onRunDiagnostics={() => handleNavigate("diagnostics")} settings={settings.status === "loaded" ? settings.data : undefined} showOnboardingReplay={onboardingReplayVisible} onDismissOnboardingReplay={() => setOnboardingReplayVisible(false)} />
          </section>
        </section>
      ) : activeRoute === "contexts" ? (
        <><ContextsContent contexts={contexts} onSelect={setContextDetailsID} onNew={() => setCreatingContext(true)} />{contextDetailsID ? <ContextDetailsDrawer contextId={contextDetailsID} onClose={() => setContextDetailsID(undefined)} load={(contextId) => devContextApi.getContextDetails({contextId})} duplicate={async (request) => { const result = await devContextApi.duplicateContext(request); if (result.ok) { await refreshContexts(); } return result; }} exportMetadata={devContextApi.exportContextMetadata} importMetadata={handleImportContextMetadata}/> : null}{creatingContext && contexts.status === "loaded" ? <CreateContextDialog contexts={contexts.data} onClose={() => setCreatingContext(false)} create={async (request) => { const result = await devContextApi.createContext(request); if (result.ok) { await refreshContexts(); setCreatingContext(false); } return result; }} loadTemplates={() => devContextApi.getContextTemplates()}/> : null}</>
      ) : activeRoute === "projects" ? (
        <>
          <ProjectsContent projects={projects} launchingProjectPath={projectLaunchPath} errorProjectPath={projectErrorPath} launchError={projectLaunchError} onLaunch={handleProjectLaunch} onChangeContext={contexts.status === "loaded" ? handleProjectChangeContext : undefined} onOpenFolder={handleProjectOpenFolder} />
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
      ) : activeRoute === "diagnostics" ? (
        <DiagnosticsView contexts={contexts.status === "loaded" ? contexts.data : []} load={(contextId) => devContextApi.getDiagnostics({contextId})} loadRepairActions={(contextId) => devContextApi.getRepairActions({contextId})} runRepairAction={(contextId, actionId, confirmDestructive) => devContextApi.runRepairAction({contextId, actionId, confirmDestructive})} />
      ) : activeRoute === "history" ? (
        <HistoryContent history={history} />
      ) : activeRoute === "running" ? (
        <RunningContent running={running} />
      ) : activeRoute === "settings" ? (
        settings.status === "loaded" ? <SettingsView settings={settings.data} pending={settingsPending} onChange={(next) => void handleSettingsChange(next)} onReplayOnboarding={() => { setOnboardingReplayVisible(true); handleNavigate("home"); }} /> : settings.status === "error" ? <GuiErrorNotice error={settings.error} /> : <p className="text-sm text-muted-foreground">Loading settings...</p>
      ) : activeRoute === "trust" ? (
        <TrustCenterContent trustCenter={trustCenter} />
      ) : (
        <PlaceholderScreen route={activeRoute} />
      )}
      {commandPaletteLaunchError ? <GuiErrorNotice error={commandPaletteLaunchError} /> : null}
      <CommandPalette
        open={commandPaletteOpen}
        onOpenChange={setCommandPaletteOpen}
        launchActions={launchState.status === "loaded" ? launchContextActions(launchState.data.contexts, (contextId) => void handleCommandPaletteLaunch(contextId)) : []}
        navigationActions={navigationActions(appRouteDefinitions, handleNavigate)}
      />
      {pendingRunningEnvironmentLaunch ? <RunningEnvironmentConflictDialog conflict={pendingRunningEnvironmentLaunch.conflict} launchPending={runningEnvironmentLaunchPending} error={runningEnvironmentLaunchError} onCancel={() => !runningEnvironmentLaunchPending && setPendingRunningEnvironmentLaunch(undefined)} onLaunchAnother={() => void handleLaunchAnotherWindow()} /> : null}
    </AppShell>
  );
}

function notifyLaunch(result: {project: {name: string}; context: {name: string; tool: {name: string}}}) {
  notifyCodingToolLaunched({
    projectName: result.project.name,
    contextName: result.context.name,
    toolName: result.context.tool.name,
  });
}

export default App;
