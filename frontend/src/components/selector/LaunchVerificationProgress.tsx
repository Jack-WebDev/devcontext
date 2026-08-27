import type { LaunchVerificationStep, LaunchVerificationStepStatus } from "../../lib/devctx-api";
import { Card } from "../ui/card.js";

interface LaunchVerificationProgressProps {
  projectName: string;
  contextName: string;
  steps?: LaunchVerificationStep[];
}

function LaunchVerificationProgress({ projectName, contextName, steps = [] }: LaunchVerificationProgressProps) {
  return (
    <Card
      as="section"
      size="sm"
      className="border border-border bg-muted/30 p-4 text-sm"
      aria-labelledby="launch-verification-title"
      aria-live="polite"
      role="status"
    >
      <h3 id="launch-verification-title" className="font-medium">Launch verification</h3>
      <p className="mt-1 text-muted-foreground">Launching {projectName} as {contextName}...</p>
      {steps.length === 0 ? (
        <p className="mt-3 text-muted-foreground">Preparing launch verification...</p>
      ) : (
        <ol className="mt-3 space-y-3 border-t border-border pt-3">
          {steps.map((step) => (
            <VerificationStepRow key={step.id} step={step} />
          ))}
        </ol>
      )}
    </Card>
  );
}

function VerificationStepRow({ step }: { step: LaunchVerificationStep }) {
  const presentation = verificationStepPresentation(step.status);

  return (
    <li className="flex min-w-0 items-start justify-between gap-3">
      <div className="min-w-0">
        <p className="font-medium">{step.label}</p>
        <p className="mt-1 text-xs text-muted-foreground">{step.message}</p>
      </div>
      <span className={`shrink-0 text-xs font-medium ${presentation.className}`}>{presentation.label}</span>
    </li>
  );
}

function verificationStepPresentation(status: LaunchVerificationStepStatus): { label: string; className: string } {
  switch (status) {
    case "ready":
      return {label: "Ready", className: "text-emerald-700"};
    case "needs_attention":
      return {label: "Needs attention", className: "text-amber-700"};
    case "blocked":
      return {label: "Blocked", className: "text-destructive"};
    case "pending":
      return {label: "Pending", className: "text-muted-foreground"};
  }
}

export { LaunchVerificationProgress, verificationStepPresentation };
