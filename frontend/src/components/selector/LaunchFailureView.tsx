import type { DisplayError } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { GuiErrorNotice } from "./GuiErrorNotice.js";

interface LaunchFailureViewProps {
  error: DisplayError;
  onRetry: () => void;
  onCancel: () => void;
  onRunDiagnostics?: () => void;
  onOpenConfiguration?: () => void;
}

function LaunchFailureView({
  error,
  onRetry,
  onCancel,
  onRunDiagnostics,
  onOpenConfiguration,
}: LaunchFailureViewProps) {
  return (
    <section className="space-y-4" aria-labelledby="launch-failure-actions-title">
      <GuiErrorNotice error={error} />
      <div>
        <h3 id="launch-failure-actions-title" className="text-sm font-medium">Next steps</h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Dev Context is still open so you can resolve the issue or try the launch again.
        </p>
      </div>
      <div className="flex flex-wrap gap-3">
        <Button type="button" onClick={onRetry}>Retry</Button>
        <Button type="button" variant="outline" disabled={onRunDiagnostics === undefined} onClick={onRunDiagnostics}>
          Run diagnostics
        </Button>
        <Button type="button" variant="outline" disabled={onOpenConfiguration === undefined} onClick={onOpenConfiguration}>
          Open configuration
        </Button>
        <Button type="button" variant="ghost" onClick={onCancel}>Cancel</Button>
      </div>
      {error.technicalDetails ? (
        <details className="border-t border-border pt-4 text-sm">
          <summary className="cursor-pointer font-medium">Technical details</summary>
          <pre className="mt-3 overflow-x-auto whitespace-pre-wrap wrap-break-word bg-muted/40 p-3 text-xs text-muted-foreground">
            {error.technicalDetails}
          </pre>
        </details>
      ) : null}
    </section>
  );
}

export { LaunchFailureView };
export type { LaunchFailureViewProps };
