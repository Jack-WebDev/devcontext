import type { DisplayError, RunningEnvironmentConflict } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface RunningEnvironmentConflictDialogProps {
  conflict: RunningEnvironmentConflict;
  launchPending: boolean;
  error?: DisplayError;
  onCancel: () => void;
  onLaunchAnother: () => void;
}

function RunningEnvironmentConflictDialog({conflict, launchPending, error, onCancel, onLaunchAnother}: RunningEnvironmentConflictDialogProps) {
  const sameContext = conflict.kind === "same_context";
  const environment = conflict.environment;
  return <Card as="section" role="dialog" aria-modal="true" aria-labelledby="running-environment-conflict-heading" className="border border-border py-0"><CardContent className="space-y-4 p-5"><div><h3 id="running-environment-conflict-heading" className="text-base font-semibold">{sameContext ? "Environment already running" : "Project is running in another context"}</h3><p className="mt-1 text-sm text-muted-foreground">{sameContext ? "This project is already open with the selected context." : "This project is already open with a different context. Contexts stay isolated for the lifetime of each environment."}</p></div><dl className="grid gap-2 text-sm"><ConflictDetail label="Project" value={environment.project.name} /><ConflictDetail label="Running context" value={environment.context.name} /><ConflictDetail label="Coding tool" value={environment.tool.name} /></dl>{error ? <p className="text-sm text-destructive" role="alert">{error.message}</p> : null}<div className="flex flex-wrap justify-end gap-3"><Button type="button" variant="outline" disabled={launchPending} onClick={onCancel}>Cancel</Button><Button type="button" variant="outline" disabled title="Switching to an existing window is not available for this coding tool yet.">Switch to existing window</Button><Button type="button" disabled={launchPending} onClick={onLaunchAnother}>{launchPending ? "Launching..." : "Launch another window"}</Button></div></CardContent></Card>;
}

function ConflictDetail({label, value}: {label: string; value: string}) { return <div className="grid gap-1"><dt className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">{label}</dt><dd className="font-medium">{value}</dd></div>; }

export { RunningEnvironmentConflictDialog };
export type { RunningEnvironmentConflictDialogProps };
