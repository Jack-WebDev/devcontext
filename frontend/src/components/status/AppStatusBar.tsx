import { ShieldCheck } from "lucide-react";
import type { LaunchState } from "../../lib/devctx-api";

function AppStatusBar({ launchState }: { launchState?: LaunchState }) {
  const needsAttention =
    launchState?.confidence?.status === "blocked" ||
    (launchState?.warnings.length ?? 0) > 0;
  const isolation =
    launchState?.confidence === undefined
      ? "Checking isolation"
      : launchState.confidence.status === "blocked"
        ? "Isolation needs attention"
        : "Isolation enabled";
  return (
    <footer
      className="h-[55px] border-t border-border bg-[#faf9f7]"
      aria-label="Application status"
    >
      <div className="flex h-full items-center text-[11px] text-muted-foreground">
        <div className="flex h-full w-[238px] items-center gap-2 border-r border-border px-5">
          <ShieldCheck className="size-4" />
          {isolation}
        </div>
        <div className="flex items-center gap-2 px-5">
          <span
            className={`size-2 rounded-full ${needsAttention ? "bg-warning" : "bg-success"}`}
          />
          {needsAttention
            ? "System needs attention"
            : "All systems operational"}
        </div>
      </div>
    </footer>
  );
}

export { AppStatusBar };
