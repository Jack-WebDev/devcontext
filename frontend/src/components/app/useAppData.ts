import { useEffect, useState } from "react";

import type {
  ContextListItem,
  HistoryState,
  HomeDashboardState,
  LaunchState,
  ProjectsState,
  RecentProjectState,
  RunningEnvironmentsState,
  SettingsState,
  TrustCenterState,
} from "../../lib/devctx-api";
import { devContextApi } from "../../lib/devctx-api";
import type { AppRoute } from "../shell/routes";
import { loadStateFromResult, type LoadState } from "./load-state";

export function useAppData(activeRoute: AppRoute) {
  const [launchState, setLaunchState] = useState<LoadState<LaunchState>>({ status: "loading" });
  const [homeDashboard, setHomeDashboard] = useState<LoadState<HomeDashboardState>>({ status: "loading" });
  const [recentProjects, setRecentProjects] = useState<LoadState<RecentProjectState[]>>({ status: "loading" });
  const [contexts, setContexts] = useState<LoadState<ContextListItem[]>>({ status: "loading" });
  const [projects, setProjects] = useState<LoadState<ProjectsState>>({ status: "loading" });
  const [history, setHistory] = useState<LoadState<HistoryState>>({ status: "loading" });
  const [running, setRunning] = useState<LoadState<RunningEnvironmentsState>>({ status: "loading" });
  const [settings, setSettings] = useState<LoadState<SettingsState>>({ status: "loading" });
  const [trustCenter, setTrustCenter] = useState<LoadState<TrustCenterState>>({ status: "loading" });

  async function refreshHomeDashboard() {
    setHomeDashboard(loadStateFromResult(await devContextApi.getHomeDashboard()));
  }

  async function refreshRecentProjects() {
    const result = await devContextApi.getRecentProjects();
    setRecentProjects(result.ok ? { status: "loaded", data: result.data.projects } : { status: "error", error: result.error });
  }

  async function refreshContexts() {
    const result = await devContextApi.getContexts();
    setContexts(result.ok ? { status: "loaded", data: result.data.contexts } : { status: "error", error: result.error });
  }

  async function refreshProjects() {
    setProjects(loadStateFromResult(await devContextApi.getProjects()));
  }

  async function refreshHistory() {
    setHistory(loadStateFromResult(await devContextApi.getHistory()));
  }

  async function refreshRunningEnvironments() {
    setRunning(loadStateFromResult(await devContextApi.getRunningEnvironments()));
  }

  async function refreshSettings() {
    setSettings(loadStateFromResult(await devContextApi.getSettings()));
  }

  async function refreshTrustCenter() {
    setTrustCenter(loadStateFromResult(await devContextApi.getTrustCenter()));
  }

  useEffect(() => {
    let active = true;
    void devContextApi.getLaunchState().then((result) => {
      if (active) {
        setLaunchState(loadStateFromResult(result));
      }
    });
    return () => {
      active = false;
    };
  }, []);

  useEffect(() => {
    void refreshSettings();
    void refreshRecentProjects();
    void refreshContexts();
    void refreshProjects();
    void refreshHomeDashboard();
  }, []);

  useEffect(() => {
    if (activeRoute === "history") {
      void refreshHistory();
    }
    if (activeRoute === "running") {
      void refreshRunningEnvironments();
    }
    if (activeRoute === "trust") {
      void refreshTrustCenter();
    }
  }, [activeRoute]);

  return {
    launchState,
    setLaunchState,
    homeDashboard,
    recentProjects,
    contexts,
    projects,
    history,
    running,
    settings,
    trustCenter,
    refreshHomeDashboard,
    refreshRecentProjects,
    refreshContexts,
    refreshProjects,
    refreshRunningEnvironments,
    setSettings,
    refreshTrustCenter,
  };
}
