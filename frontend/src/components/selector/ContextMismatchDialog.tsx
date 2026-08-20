import type { ContextMismatch, ContextState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ContextMismatchDialogProps {
  mismatch: ContextMismatch;
  contexts: ContextState[];
  launchPending: boolean;
  onCancel: () => void;
  onOpenAnyway: () => void;
}

function ContextMismatchDialog({
  mismatch,
  contexts,
  launchPending,
  onCancel,
  onOpenAnyway,
}: ContextMismatchDialogProps) {
  return (
    <Card
      as="section"
      aria-labelledby="context-mismatch-title"
      aria-modal="true"
      className="border border-destructive/30 py-0"
      role="dialog"
    >
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="context-mismatch-title" className="text-base font-semibold text-foreground">
            Context mismatch
          </h3>
          <p className="mt-1 text-muted-foreground">
            This project is remembered for a different context. Opening anyway can expose project files to the
            requested context's local tools and credentials.
          </p>
        </div>

        <dl className="grid gap-2 text-sm">
          <MismatchDetail label="Project" value={mismatch.projectPath} />
          <MismatchDetail label="Remembered context" value={contextName(contexts, mismatch.boundContextId)} />
          <MismatchDetail label="Requested context" value={contextName(contexts, mismatch.requestedContextId)} />
        </dl>

        <div className="flex justify-end gap-3">
          <Button
            type="button"
            variant="outline"
            disabled={launchPending}
            onClick={onCancel}
          >
            Cancel
          </Button>
          <Button
            type="button"
            variant="destructive"
            disabled={launchPending}
            onClick={onOpenAnyway}
          >
            {launchPending ? "Opening..." : "Open Anyway"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function MismatchDetail({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1">
      <dt className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</dt>
      <dd className="truncate font-mono text-sm text-foreground" title={value}>
        {value}
      </dd>
    </div>
  );
}

function contextName(contexts: ContextState[], contextId: string): string {
  const context = contexts.find((candidate) => candidate.id === contextId);
  return context?.name ?? contextId;
}

export { ContextMismatchDialog, contextName };
