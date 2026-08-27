import { useEffect, useState } from "react";

import { GuiErrorNotice } from "./components/selector/GuiErrorNotice";
import { SelectorView } from "./components/selector/SelectorView";
import { createOnboardingContextAndRefresh } from "./components/selector/onboarding-action";
import { HomeView } from "./components/home/HomeView";
import { AppShell } from "./components/shell/AppShell";
import { appRouteDefinition, appRouteFromHash, type AppRoute } from "./components/shell/routes";
import {
  devContextApi,
  type ApiResult,
  type CreateContextResult,
  type DisplayError,
  type HomeDashboardState,
  type LaunchState,
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

function App() {
  const [launchState, setLaunchState] = useState<LaunchStateLoad>({
    status: "loading",
  });
  const [homeDashboard, setHomeDashboard] = useState<HomeDashboardLoad>({status: "loading"});
  const [homeLaunchPending, setHomeLaunchPending] = useState(false);
  const [homeLaunchError, setHomeLaunchError] = useState<DisplayError | undefined>(undefined);
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
      return { ok: true, data: result.created };
    }

    return { ok: false, error: result.error };
  }

  async function refreshHomeDashboard() {
    const result = await devContextApi.getHomeDashboard();
    setHomeDashboard(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error});
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
    } finally {
      setHomeLaunchPending(false);
    }
  }

  function handleReviewLaunchOptions() {
    document.getElementById("context-selector")?.scrollIntoView({behavior: "smooth", block: "start"});
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
          {renderHomeDashboard(homeDashboard, homeLaunchPending, homeLaunchError, handleHomeQuickLaunch, handleReviewLaunchOptions)}
          <section id="context-selector" aria-labelledby="context-selector-heading" className="space-y-6">
            <div>
              <p className="text-sm text-muted-foreground">Launch options</p>
              <h2 id="context-selector-heading" className="text-xl font-semibold">Context selector</h2>
            </div>
          {renderSelectorContent(launchState, handleCreateContext)}
          </section>
        </section>
      ) : (
        <PlaceholderScreen route={activeRoute} />
      )}
    </AppShell>
  );
}

function renderHomeDashboard(
  dashboard: HomeDashboardLoad,
  launchPending: boolean,
  launchError: DisplayError | undefined,
  onQuickLaunch: () => void,
  onReviewLaunchOptions: () => void,
) {
  if (dashboard.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading Home dashboard...</p>;
  }
  if (dashboard.status === "error") {
    return <GuiErrorNotice error={dashboard.error} />;
  }
  return (
    <HomeView
      dashboard={dashboard.data}
      launchPending={launchPending}
      launchError={launchError}
      onQuickLaunch={onQuickLaunch}
      onReviewLaunchOptions={onReviewLaunchOptions}
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
