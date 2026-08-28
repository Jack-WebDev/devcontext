import type { DisplayError, RecentProjectState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface RecentProjectConfirmationDialogProps {
  project: RecentProjectState;
  launchPending: boolean;
  error?: DisplayError;
  onCancel: () => void;
  onConfirm: () => void;
}

function RecentProjectConfirmationDialog({
  project,
  launchPending,
  error,
  onCancel,
  onConfirm,
}: RecentProjectConfirmationDialogProps) {
  const contextName = project.contextName ?? project.contextId;
  return (
    <Card
      as="section"
      aria-labelledby="recent-project-confirmation-title"
      aria-modal="true"
      className="border border-border py-0"
      role="dialog"
    >
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="recent-project-confirmation-title" className="text-base font-semibold text-foreground">
            Launch recent project?
          </h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Dev Context will check this project and context before opening it.
          </p>
        </div>

        <dl className="grid gap-2 text-sm">
          <RecentProjectDetail label="Project" value={project.project.name} />
          <RecentProjectDetail label="Path" value={project.project.path} />
          <RecentProjectDetail label="Context" value={contextName} />
        </dl>

        {error ? <p className="text-sm text-destructive" role="alert">{error.message}</p> : null}

        <div className="flex justify-end gap-3">
          <Button type="button" variant="outline" disabled={launchPending} onClick={onCancel}>
            Cancel
          </Button>
          <Button type="button" disabled={launchPending} onClick={onConfirm}>
            {launchPending ? "Checking launch..." : `Launch ${contextName}`}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function RecentProjectDetail({label, value}: {label: string; value: string}) {
  return (
    <div className="grid gap-1">
      <dt className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</dt>
      <dd className="truncate font-mono text-sm text-foreground" title={value}>{value}</dd>
    </div>
  );
}

export { RecentProjectConfirmationDialog };
