import type { ContextState, LaunchConfidenceCheck, LaunchConfidenceStatus, ProjectState } from "../../lib/devctx-api";
import { Card } from "../ui/card.js";

interface SelectorConfidenceSummaryProps {
  context?: ContextState;
  project?: ProjectState;
}

function SelectorConfidenceSummary({ context, project }: SelectorConfidenceSummaryProps) {
  if (context === undefined || context.confidence === undefined || project === undefined) {
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
          <h3 className="truncate font-medium">Launch confidence</h3>
          <p className="mt-1 text-muted-foreground">Review the selected development identity before launch.</p>
        </div>
        <span className={`shrink-0 font-medium ${status.className}`}>{status.label}</span>
      </div>
      <dl className="mt-4 space-y-3 border-t border-border pt-3">
        <ConfidenceSummaryRow label="Project" value={project.name} />
        <ConfidenceSummaryRow label="Context" value={context.name} />
        {context.confidence.checks
          .filter((check) => check.component === "provider")
          .map((check) => <ConfidenceCheckRow key={`provider-${check.providerId}`} check={check} />)}
        {context.confidence.checks
          .filter((check) => check.component === "tool" && check.toolId === context.tool.id)
          .map((check) => <ConfidenceCheckRow key={`tool-${check.toolId}`} check={check} />)}
        <IsolationConfidenceRow checks={context.confidence.checks.filter((check) => check.component === "isolation")} />
      </dl>
    </Card>
  );
}

function ConfidenceSummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex min-w-0 items-baseline justify-between gap-3 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="truncate font-medium" title={value}>{value}</dd>
    </div>
  );
}

function ConfidenceCheckRow({ check }: { check: LaunchConfidenceCheck }) {
  const status = confidenceStatusPresentation(check.severity);

  return (
    <div className="space-y-1 text-sm">
      <div className="flex min-w-0 items-baseline justify-between gap-3">
        <dt className="truncate text-muted-foreground" title={check.label}>{check.label}</dt>
        <dd className={`shrink-0 font-medium ${status.className}`}>{status.label}</dd>
      </div>
      <p className="text-xs text-muted-foreground">{check.message}</p>
    </div>
  );
}

function IsolationConfidenceRow({ checks }: { checks: LaunchConfidenceCheck[] }) {
  const status = mostSevereConfidenceStatus(checks);
  if (status === undefined) {
    return null;
  }

  const presentation = confidenceStatusPresentation(status);
  const label = status === "ready" ? "Protected" : presentation.label;

  return (
    <div className="space-y-1 text-sm">
      <div className="flex min-w-0 items-baseline justify-between gap-3">
        <dt className="text-muted-foreground">Isolation</dt>
        <dd className={`shrink-0 font-medium ${presentation.className}`}>{label}</dd>
      </div>
      <p className="text-xs text-muted-foreground">{checks.map((check) => check.message).join(" ")}</p>
    </div>
  );
}

function mostSevereConfidenceStatus(checks: LaunchConfidenceCheck[]): LaunchConfidenceStatus | undefined {
  if (checks.some((check) => check.severity === "blocked")) {
    return "blocked";
  }
  if (checks.some((check) => check.severity === "needs_attention")) {
    return "needs_attention";
  }
  if (checks.some((check) => check.severity === "ready")) {
    return "ready";
  }
  return undefined;
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
