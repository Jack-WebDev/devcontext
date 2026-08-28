import type { LaunchState } from "../../lib/devctx-api";

function AppStatusBar({launchState}: {launchState?: LaunchState}) {
  const warningCount = launchState?.warnings.length ?? 0;
  const health = launchState?.confidence?.status === "blocked" ? "Needs attention" : warningCount > 0 ? "Needs attention" : "Healthy";
  const isolation = launchState?.confidence === undefined ? "Checking isolation" : launchState.confidence.status === "blocked" ? "Isolation needs attention" : "Isolation enabled";
  const trayIndicator = health === "Healthy" ? "Normal" : "Warning";
  return <footer className="border-t border-border bg-background px-4 py-2 text-xs text-muted-foreground" aria-label="Application status"><div className="mx-auto flex max-w-6xl flex-wrap gap-x-5 gap-y-1"><span>{isolation}</span><span>System health: {health}</span><span>Warnings: {warningCount}</span><span>Tray indicator: {trayIndicator}</span><span>Dev Context</span></div></footer>;
}

export { AppStatusBar };
