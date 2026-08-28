import { useEffect, useState } from "react";

import type {
  ApiResult,
  ContextDetailsState,
  ContextMetadataExport,
  DisplayError,
  DuplicateContextRequest,
  DuplicateContextResult,
  ExportContextMetadataRequest,
  ImportContextMetadataRequest,
  ImportContextMetadataResult,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "../ui/sheet.js";
import { ContextField } from "./ContextField";
import { parseContextMetadataExport } from "./context-transfer";

interface ContextDetailsDrawerProps {
  contextId: string;
  onClose: () => void;
  load: (id: string) => Promise<ApiResult<ContextDetailsState>>;
  duplicate: (request: DuplicateContextRequest) => Promise<ApiResult<DuplicateContextResult>>;
  exportMetadata: (request: ExportContextMetadataRequest) => Promise<ApiResult<ContextMetadataExport>>;
  importMetadata: (request: ImportContextMetadataRequest) => Promise<ApiResult<ImportContextMetadataResult>>;
}

export function ContextDetailsDrawer({ contextId, onClose, load, duplicate, exportMetadata, importMetadata }: ContextDetailsDrawerProps) {
  const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
  const [duplicateID, setDuplicateID] = useState("");
  const [duplicateName, setDuplicateName] = useState("");
  const [duplicateError, setDuplicateError] = useState<DisplayError>();
  const [duplicatePending, setDuplicatePending] = useState(false);
  const [exportedMetadata, setExportedMetadata] = useState("");
  const [exportError, setExportError] = useState<DisplayError>();
  const [exportPending, setExportPending] = useState(false);
  const [importedMetadata, setImportedMetadata] = useState("");
  const [importID, setImportID] = useState("");
  const [importError, setImportError] = useState<string>();
  const [importPending, setImportPending] = useState(false);

  useEffect(() => {
    void load(contextId).then(setResult);
  }, [contextId, load]);

  useEffect(() => {
    if (result?.ok) {
      setDuplicateID(`${result.data.context.id}-copy`);
      setDuplicateName(`${result.data.context.name} copy`);
      setImportID(`${result.data.context.id}-imported`);
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

  async function prepareExport() {
    if (!result?.ok) return;
    setExportPending(true);
    setExportError(undefined);
    const exported = await exportMetadata({contextId: result.data.context.id});
    setExportPending(false);
    if (!exported.ok) {
      setExportError(exported.error);
      return;
    }
    setExportedMetadata(JSON.stringify(exported.data, null, 2));
  }

  async function submitImport() {
    let exported: ContextMetadataExport;
    try {
      exported = parseContextMetadataExport(importedMetadata);
    } catch {
      setImportError("Paste a valid context metadata export before importing.");
      return;
    }
    setImportPending(true);
    setImportError(undefined);
    const imported = await importMetadata({contextId: importID, export: exported});
    setImportPending(false);
    if (!imported.ok) {
      setImportError(imported.error.message);
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
              <section className="space-y-3 border-t border-border pt-4" aria-labelledby="context-transfer-heading">
                <div>
                  <h3 id="context-transfer-heading" className="font-medium">Import and export</h3>
                  <p className="mt-1 text-sm text-muted-foreground">Exports include context metadata and non-secret provider and coding-tool settings. Credentials are never included or imported.</p>
                </div>
                <Button type="button" variant="outline" disabled={exportPending} onClick={() => void prepareExport()}>
                  {exportPending ? "Preparing export..." : "Prepare safe export"}
                </Button>
                {exportError ? <p className="text-destructive">{exportError.message}</p> : null}
                {exportedMetadata ? <label className="block text-sm">Safe context metadata<textarea aria-label="Safe context metadata export" className="mt-1 min-h-40 w-full border p-2 font-mono text-xs" value={exportedMetadata} readOnly /></label> : null}
                <label className="block text-sm">Import context metadata<textarea aria-label="Import context metadata" className="mt-1 min-h-40 w-full border p-2 font-mono text-xs" value={importedMetadata} onChange={(event) => setImportedMetadata(event.target.value)} placeholder="Paste a safe context metadata export" /></label>
                <ContextField label="New context ID" value={importID} onChange={setImportID} />
                {importError ? <p className="text-destructive">{importError}</p> : null}
                <Button type="button" disabled={importPending || !importID || !importedMetadata} onClick={() => void submitImport()}>
                  {importPending ? "Importing..." : "Import as new context"}
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
