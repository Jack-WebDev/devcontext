import type { ContextState, LaunchConfidenceStatus } from "../../lib/devctx-api";
import { Card } from "../ui/card.js";

interface SelectorConfidenceSummaryProps {
  context?: ContextState;
}

function SelectorConfidenceSummary({ context }: SelectorConfidenceSummaryProps) {
  if (context === undefined || context.confidence === undefined) {
    return (
      <Card as="div" size="sm" className="border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
        Select a context to review launch readiness.
      </Card>
    );
  }

  const status = confidenceStatusPresentation(context.confidence.status);

  return (
    <Card as="div" size="sm" className="border border-border bg-muted/30 p-3 text-sm">
      <div className="flex min-w-0 items-center justify-between gap-3">
        <div className="min-w-0">
          <h3 className="truncate font-medium">Confidence summary</h3>
          <p className="mt-1 truncate text-muted-foreground" title={context.name}>
            {context.name}
          </p>
        </div>
        <span className={`shrink-0 font-medium ${status.className}`}>{status.label}</span>
      </div>
    </Card>
  );
}

function confidenceStatusPresentation(status: LaunchConfidenceStatus): { label: string; className: string } {
  switch (status) {
    case "ready":
      return { label: "Ready", className: "text-emerald-700" };
    case "needs_attention":
      return { label: "Needs attention", className: "text-amber-700" };
    case "blocked":
      return { label: "Blocked", className: "text-destructive" };
    default:
      return { label: "Needs attention", className: "text-muted-foreground" };
  }
}

export { SelectorConfidenceSummary, confidenceStatusPresentation };
