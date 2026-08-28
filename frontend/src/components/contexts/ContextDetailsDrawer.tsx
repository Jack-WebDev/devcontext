import { useEffect, useState } from "react";

import type { ApiResult, ContextDetailsState, DisplayError, DuplicateContextRequest, DuplicateContextResult } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "../ui/sheet.js";
import { ContextField } from "./ContextField";

interface ContextDetailsDrawerProps {
  contextId: string;
  onClose: () => void;
  load: (id: string) => Promise<ApiResult<ContextDetailsState>>;
  duplicate: (request: DuplicateContextRequest) => Promise<ApiResult<DuplicateContextResult>>;
}

export function ContextDetailsDrawer({ contextId, onClose, load, duplicate }: ContextDetailsDrawerProps) {
  const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
  const [duplicateID, setDuplicateID] = useState("");
  const [duplicateName, setDuplicateName] = useState("");
  const [duplicateError, setDuplicateError] = useState<DisplayError>();
  const [duplicatePending, setDuplicatePending] = useState(false);

  useEffect(() => {
    void load(contextId).then(setResult);
  }, [contextId, load]);

  useEffect(() => {
    if (result?.ok) {
      setDuplicateID(`${result.data.context.id}-copy`);
      setDuplicateName(`${result.data.context.name} copy`);
    }
  }, [result]);

  async function submitDuplicate() {
    if (!result?.ok) return;

    setDuplicatePending(true);
    setDuplicateError(undefined);
    const duplicated = await duplicate({ sourceContextId: result.data.context.id, contextId: duplicateID, name: duplicateName });
    setDuplicatePending(false);

    if (!duplicated.ok) {
      setDuplicateError(duplicated.error);
      return;
    }
    onClose();
  }

  return (
    <Sheet open onOpenChange={(open) => !open && onClose()}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>Context details</SheetTitle>
          <SheetDescription>Backend-owned context information.</SheetDescription>
        </SheetHeader>
        <div className="space-y-3 px-8 pb-8">
          {!result ? <p>Loading context details...</p> : !result.ok ? <p className="text-destructive">{result.error.message}</p> : (
            <>
              <Detail label="Name" value={result.data.context.name} />
              <Detail label="Location" value={result.data.location} />
              <Detail label="Created" value={new Date(result.data.createdAt).toLocaleString()} />
              <Detail label="Projects" value={String(result.data.projectCount)} />
              <Detail label="Coding tool" value={result.data.context.tool.name} />
              <Detail label="Providers" value={result.data.enabledProviders.map((provider) => provider.name).join(", ") || "None"} />
              <section className="space-y-3 border-t border-border pt-4" aria-labelledby="duplicate-context-heading">
                <div>
                  <h3 id="duplicate-context-heading" className="font-medium">Duplicate context</h3>
                  <p className="mt-1 text-sm text-muted-foreground">Copies context metadata, provider settings, and coding-tool settings into a new isolated context. Credentials are not copied.</p>
                </div>
                <ContextField label="New name" value={duplicateName} onChange={setDuplicateName} />
                <ContextField label="New ID" value={duplicateID} onChange={setDuplicateID} />
                {duplicateError ? <p className="text-destructive">{duplicateError.message}</p> : null}
                <Button type="button" disabled={duplicatePending || !duplicateID || !duplicateName} onClick={() => void submitDuplicate()}>
                  {duplicatePending ? "Duplicating..." : "Duplicate context"}
                </Button>
              </section>
            </>
          )}
        </div>
      </SheetContent>
    </Sheet>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return <div><p className="text-xs uppercase text-muted-foreground">{label}</p><p className="mt-1 break-all">{value}</p></div>;
}
