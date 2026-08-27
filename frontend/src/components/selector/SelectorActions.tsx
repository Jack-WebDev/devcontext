import type { LaunchConfidenceCheck, LaunchConfidenceState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Separator } from "../ui/separator.js";

interface SelectorActionsProps {
  launchDisabled: boolean;
  launchPending: boolean;
  contextName?: string;
  confidence?: LaunchConfidenceState;
  onLaunch: () => void;
  onCancel: () => void;
}

function SelectorActions({
  launchDisabled,
  launchPending,
  contextName,
  confidence,
  onLaunch,
  onCancel,
}: SelectorActionsProps) {
  const feedback = launchConfidenceFeedback(confidence, contextName);
  const launchLabel = contextName ? `Launch ${contextName}` : "Launch";

  return (
    <div className="space-y-4">
      {feedback ? <LaunchConfidenceFeedback feedback={feedback} /> : null}
      <Separator />
      <div className="flex justify-end gap-3">
        <Button type="button" variant="outline" disabled={launchPending} onClick={onCancel}>
          Cancel
        </Button>
        <Button type="button" disabled={launchDisabled || launchPending} onClick={onLaunch}>
          {launchPending ? "Launching..." : launchLabel}
        </Button>
      </div>
    </div>
  );
}

function LaunchConfidenceFeedback({ feedback }: { feedback: LaunchConfidenceFeedbackState }) {
  const role = feedback.status === "blocked" ? "alert" : "status";
  const className = feedback.status === "blocked"
    ? "border-destructive/40 bg-destructive/5"
    : feedback.status === "needs_attention"
      ? "border-amber-500/40 bg-amber-500/5"
      : "border-emerald-600/30 bg-emerald-600/5";

  return (
    <div className={`border p-3 text-sm ${className}`} role={role}>
      <p className="font-medium">{feedback.title}</p>
      <p className="mt-1 text-muted-foreground">{feedback.message}</p>
      {feedback.checks.length > 0 ? (
        <ul className="mt-2 space-y-1 text-xs text-muted-foreground">
          {feedback.checks.map((check) => (
            <li key={`${check.component}-${check.providerId ?? check.toolId ?? check.label}`}>
              <span className="font-medium text-foreground">{check.label}:</span> {check.actionHint ?? check.message}
            </li>
          ))}
        </ul>
      ) : null}
    </div>
  );
}

interface LaunchConfidenceFeedbackState {
  status: "ready" | "needs_attention" | "blocked";
  title: string;
  message: string;
  checks: LaunchConfidenceCheck[];
}

function launchConfidenceFeedback(
  confidence: LaunchConfidenceState | undefined,
  contextName: string | undefined,
): LaunchConfidenceFeedbackState | undefined {
  if (confidence === undefined || contextName === undefined) {
    return undefined;
  }

  switch (confidence.status) {
    case "blocked":
      return {
        status: "blocked",
        title: `Launch blocked for ${contextName}`,
        message: "Dev Context cannot confirm a safe launch until these issues are resolved.",
        checks: confidence.checks.filter((check) => check.severity === "blocked"),
      };
    case "needs_attention":
      return {
        status: "needs_attention",
        title: `Review ${contextName} before launch`,
        message: "Launch is available, but these items need your attention.",
        checks: confidence.checks.filter((check) => check.severity === "needs_attention"),
      };
    case "ready":
      return {
        status: "ready",
        title: `${contextName} is ready to launch`,
        message: "Everything required for a safe launch is available.",
        checks: [],
      };
    default:
      return undefined;
  }
}

export { SelectorActions, launchConfidenceFeedback };
