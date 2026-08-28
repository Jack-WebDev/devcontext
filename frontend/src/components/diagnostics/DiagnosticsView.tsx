import { useEffect, useState } from "react";

import type { ApiResult, ContextListItem, DiagnosticsState, DisplayError, RepairAction, RepairActionsState, RunRepairActionResult } from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface DiagnosticsViewProps {
  contexts: ContextListItem[];
  load: (contextId: string) => Promise<ApiResult<DiagnosticsState>>;
  loadRepairActions: (contextId: string) => Promise<ApiResult<RepairActionsState>>;
  runRepairAction: (contextId: string, actionId: string, confirmDestructive: boolean) => Promise<ApiResult<RunRepairActionResult>>;
}

type DiagnosticsLoad =
  | {status: "loading"}
  | {status: "loaded"; data: DiagnosticsState}
  | {status: "error"; error: DisplayError};

function DiagnosticsView({contexts, load, loadRepairActions, runRepairAction}: DiagnosticsViewProps) {
  const [contextID, setContextID] = useState(contexts[0]?.context.id ?? "");
  const [diagnostics, setDiagnostics] = useState<DiagnosticsLoad>({status: "loading"});
  const [repairActions, setRepairActions] = useState<RepairAction[]>([]);
  const [repairAction, setRepairAction] = useState<RepairAction>();
  const [repairPending, setRepairPending] = useState(false);
  const [repairError, setRepairError] = useState<DisplayError>();

  useEffect(() => {
    if (!contexts.some((item) => item.context.id === contextID)) {
      setContextID(contexts[0]?.context.id ?? "");
    }
  }, [contextID, contexts]);

  useEffect(() => {
    if (contextID === "") {
      setDiagnostics({status: "loaded", data: {groups: []}});
      return;
    }
    let active = true;
    setDiagnostics({status: "loading"});
    load(contextID).then((result) => {
      if (!active) {
        return;
      }
      setDiagnostics(result.ok ? {status: "loaded", data: result.data} : {status: "error", error: result.error});
    });
    return () => {
      active = false;
    };
  }, [contextID, load]);

  useEffect(() => {
    if (contextID === "") {
      setRepairActions([]);
      return;
    }
    loadRepairActions(contextID).then((result) => setRepairActions(result.ok ? result.data.actions : []));
  }, [contextID, loadRepairActions]);

  async function run(action: RepairAction, confirmDestructive = false) {
    if (repairPending) {
      return;
    }
    setRepairPending(true);
    setRepairError(undefined);
    try {
      const result = await runRepairAction(contextID, action.id, confirmDestructive);
      if (!result.ok) {
        setRepairError(result.error);
        return;
      }
      setDiagnostics({status: "loaded", data: result.data.diagnostics});
      setRepairAction(undefined);
      const actions = await loadRepairActions(contextID);
      if (actions.ok) {
        setRepairActions(actions.data.actions);
      }
    } finally {
      setRepairPending(false);
    }
  }

  return (
    <section aria-labelledby="diagnostics-heading" className="space-y-6">
      <div>
        <p className="text-sm text-muted-foreground">Technical checks</p>
        <h2 id="diagnostics-heading" className="text-2xl font-semibold">Diagnostics</h2>
        <p className="mt-1 text-sm text-muted-foreground">Review local context storage, provider setup, and coding-tool readiness.</p>
      </div>

      <label className="grid max-w-sm gap-2 text-sm font-medium" htmlFor="diagnostics-context-select">
        Context
        <select
          id="diagnostics-context-select"
          className="h-10 border border-input bg-background px-3 text-sm text-foreground"
          value={contextID}
          disabled={contexts.length === 0}
          onChange={(event) => setContextID(event.currentTarget.value)}
        >
          {contexts.map((item) => <option key={item.context.id} value={item.context.id}>{item.context.name}</option>)}
        </select>
      </label>

      {renderDiagnostics(diagnostics, contexts.length)}
      {repairActions.length > 0 ? <RepairActions actions={repairActions} pending={repairPending} onSelect={(action) => action.destructive ? setRepairAction(action) : void run(action)} /> : null}
      {repairAction ? <RepairConfirmation action={repairAction} pending={repairPending} error={repairError} onCancel={() => !repairPending && setRepairAction(undefined)} onConfirm={() => void run(repairAction, true)} /> : null}
    </section>
  );
}

function RepairActions({actions, pending, onSelect}: {actions: RepairAction[]; pending: boolean; onSelect: (action: RepairAction) => void}) {
  return (
    <Card as="section" hierarchy="tertiary" className="py-0" aria-labelledby="repair-actions-heading">
      <CardContent className="space-y-4 p-5">
        <div><h3 id="repair-actions-heading" className="text-lg font-semibold">Repair</h3><p className="mt-1 text-sm text-muted-foreground">Repair actions affect only this context’s isolated storage.</p></div>
        <div className="flex flex-wrap gap-3">{actions.map((action) => <Button key={action.id} type="button" variant={action.destructive ? "destructive" : "outline"} size="sm" disabled={pending} onClick={() => onSelect(action)}>{action.label}</Button>)}</div>
      </CardContent>
    </Card>
  );
}

function RepairConfirmation({action, pending, error, onCancel, onConfirm}: {action: RepairAction; pending: boolean; error?: DisplayError; onCancel: () => void; onConfirm: () => void}) {
  return (
    <Card as="section" aria-labelledby="repair-confirmation-heading" aria-modal="true" className="border border-destructive/30 py-0" role="dialog">
      <CardContent className="space-y-4 p-5">
        <div><h3 id="repair-confirmation-heading" className="text-lg font-semibold">Confirm {action.label}</h3><p className="mt-1 text-sm text-muted-foreground">{action.description}</p></div>
        <p className="text-sm text-destructive">This permanently removes the following context-owned items.</p>
        {action.targets.length === 0 ? <p className="text-sm text-muted-foreground">No files are currently present.</p> : <ul className="max-h-48 space-y-1 overflow-y-auto border border-border p-3 font-mono text-xs">{action.targets.map((target) => <li key={target.path}>{target.path}</li>)}</ul>}
        {error ? <p className="text-sm text-destructive" role="alert">{error.message}</p> : null}
        <div className="flex justify-end gap-3"><Button type="button" variant="outline" disabled={pending} onClick={onCancel}>Cancel</Button><Button type="button" variant="destructive" disabled={pending} onClick={onConfirm}>{pending ? "Resetting..." : "Reset storage"}</Button></div>
      </CardContent>
    </Card>
  );
}

function renderDiagnostics(diagnostics: DiagnosticsLoad, contextCount: number) {
  if (contextCount === 0) {
    return <EmptyDiagnostics message="Create a context before running diagnostics." />;
  }
  if (diagnostics.status === "loading") {
    return <p className="text-sm text-muted-foreground">Running diagnostics...</p>;
  }
  if (diagnostics.status === "error") {
    return <p className="text-sm text-destructive" role="alert">{diagnostics.error.message}</p>;
  }
  if (diagnostics.data.groups.length === 0) {
    return <EmptyDiagnostics message="No diagnostics are available for this context." />;
  }
  return <div className="space-y-4" aria-label="Diagnostic groups">{diagnostics.data.groups.map((group) => <DiagnosticGroupCard key={group.id} group={group} />)}</div>;
}

function DiagnosticGroupCard({group}: {group: DiagnosticsState["groups"][number]}) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby={`diagnostics-${group.id}-heading`}>
      <CardContent className="space-y-4 p-5">
        <h3 id={`diagnostics-${group.id}-heading`} className="text-lg font-semibold">{group.label}</h3>
        <ul className="divide-y divide-border border-y border-border">
          {group.checks.map((check) => <DiagnosticCheckRow key={check.id} check={check} />)}
        </ul>
      </CardContent>
    </Card>
  );
}

function DiagnosticCheckRow({check}: {check: DiagnosticsState["groups"][number]["checks"][number]}) {
  const visibleDetails = check.details.filter((detail) => !detail.isPath);
  const pathDetails = check.details.filter((detail) => detail.isPath);
  return (
    <li className="space-y-3 py-4">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h4 className="font-medium">{check.label}</h4>
          <p className="mt-1 text-sm text-muted-foreground">{check.message}</p>
        </div>
        <StatusIndicator status={check.severity} />
      </div>
      {visibleDetails.length > 0 ? <DiagnosticDetails details={visibleDetails} /> : null}
      {pathDetails.length > 0 ? (
        <details className="text-sm">
          <summary className="cursor-pointer font-medium text-muted-foreground">Show paths</summary>
          <DiagnosticDetails details={pathDetails} className="mt-3" />
        </details>
      ) : null}
    </li>
  );
}

function DiagnosticDetails({details, className = ""}: {details: {label: string; value: string}[]; className?: string}) {
  return (
    <dl className={`grid gap-2 text-sm ${className}`}>
      {details.map((detail) => (
        <div key={`${detail.label}:${detail.value}`} className="grid gap-1 sm:grid-cols-[10rem_1fr] sm:gap-3">
          <dt className="text-muted-foreground">{detail.label}</dt>
          <dd className="break-all font-mono text-xs" title={detail.value}>{detail.value}</dd>
        </div>
      ))}
    </dl>
  );
}

function EmptyDiagnostics({message}: {message: string}) {
  return <Card as="section" hierarchy="secondary" className="py-0"><CardContent className="p-5 text-sm text-muted-foreground">{message}</CardContent></Card>;
}

export { DiagnosticsView, DiagnosticCheckRow, RepairConfirmation, renderDiagnostics };
