import { useEffect, useState } from "react";
import type {
  ApiResult,
  ContextDetailsState,
  ContextListItem,
  CreateContextResult,
  CreateContextRequest,
  ContextTemplateState,
  DuplicateContextRequest,
  DuplicateContextResult,
  DisplayError,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "../ui/sheet.js";

type CustomContextRequest = CreateContextRequest & {
  name: string;
  description: string;
  icon: string;
  accent: string;
  toolId: string;
  enabledProviderIds: string[];
};

export function ContextDetailsDrawer({
  contextId,
  onClose,
  load,
  duplicate,
}: {
  contextId: string;
  onClose: () => void;
  load: (id: string) => Promise<ApiResult<ContextDetailsState>>;
  duplicate: (
    request: DuplicateContextRequest,
  ) => Promise<ApiResult<DuplicateContextResult>>;
}) {
  const [result, setResult] = useState<
    ApiResult<ContextDetailsState> | undefined
  >();
  const [duplicateID, setDuplicateID] = useState("");
  const [duplicateName, setDuplicateName] = useState("");
  const [duplicateError, setDuplicateError] = useState<DisplayError>();
  const [duplicatePending, setDuplicatePending] = useState(false);
  useEffect(() => {
    void load(contextId).then(setResult);
  }, [contextId, load]);
  useEffect(() => {
    if (!result?.ok) {
      return;
    }
    setDuplicateID(`${result.data.context.id}-copy`);
    setDuplicateName(`${result.data.context.name} copy`);
  }, [result]);
  async function submitDuplicate() {
    if (!result?.ok) {
      return;
    }
    setDuplicatePending(true);
    setDuplicateError(undefined);
    const duplicated = await duplicate({
      sourceContextId: result.data.context.id,
      contextId: duplicateID,
      name: duplicateName,
    });
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
          <SheetDescription>
            Backend-owned context information.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-3 px-8 pb-8">
          {!result ? (
            <p>Loading context details...</p>
          ) : !result.ok ? (
            <p className="text-destructive">{result.error.message}</p>
          ) : (
            <>
              <Detail label="Name" value={result.data.context.name} />
              <Detail label="Location" value={result.data.location} />
              <Detail
                label="Created"
                value={new Date(result.data.createdAt).toLocaleString()}
              />
              <Detail
                label="Projects"
                value={String(result.data.projectCount)}
              />
              <Detail
                label="Coding tool"
                value={result.data.context.tool.name}
              />
              <Detail
                label="Providers"
                value={
                  result.data.enabledProviders.map((p) => p.name).join(", ") ||
                  "None"
                }
              />
              <section className="space-y-3 border-t border-border pt-4" aria-labelledby="duplicate-context-heading">
                <div>
                  <h3 id="duplicate-context-heading" className="font-medium">Duplicate context</h3>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Copies context metadata, provider settings, and coding-tool settings into a new isolated context. Credentials are not copied.
                  </p>
                </div>
                <Field label="New name" value={duplicateName} onChange={setDuplicateName} />
                <Field label="New ID" value={duplicateID} onChange={setDuplicateID} />
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
  return (
    <div>
      <p className="text-xs uppercase text-muted-foreground">{label}</p>
      <p className="mt-1 break-all">{value}</p>
    </div>
  );
}

export function CreateContextDialog({
  contexts,
  onClose,
  create,
  loadTemplates,
}: {
  contexts: ContextListItem[];
  onClose: () => void;
  create: (
    request: CreateContextRequest,
  ) => Promise<ApiResult<CreateContextResult>>;
  loadTemplates: () => Promise<ApiResult<{ templates: ContextTemplateState[] }>>;
}) {
  const [name, setName] = useState("");
  const [contextId, setContextID] = useState("");
  const [description, setDescription] = useState("");
  const [icon, setIcon] = useState("");
  const [accent, setAccent] = useState("custom");
  const [toolId, setToolID] = useState(contexts[0]?.context.tool.id ?? "");
  const [providers, setProviders] = useState<string[]>([]);
  const [templates, setTemplates] = useState<ContextTemplateState[]>([]);
  const [templateID, setTemplateID] = useState("custom");
  const [error, setError] = useState<DisplayError>();
  const [pending, setPending] = useState(false);
  const options = contexts[0]?.context.availableTools ?? [];
  const providerOptions = contexts
    .flatMap((c) => c.context.providers)
    .filter((p, i, all) => all.findIndex((x) => x.id === p.id) === i);
  useEffect(() => { void loadTemplates().then((result) => { if (result.ok) setTemplates(result.data.templates); }); }, [loadTemplates]);
  function selectTemplate(id: string) {
    setTemplateID(id);
    const template = templates.find((item) => item.id === id);
    if (!template) return;
    setName(template.name);
    setDescription(template.description);
    setIcon(template.icon ?? "");
    setAccent(template.accent);
  }
  async function submit() {
    setPending(true);
    setError(undefined);
    const request: CustomContextRequest = {
      contextId,
      templateId: templateID,
      name,
      description,
      icon,
      accent,
      toolId,
      enabledProviderIds: providers,
    };
    const result = await create(request);
    setPending(false);
    if (!result.ok) {
      setError(result.error);
      return;
    }
    onClose();
  }
  return (
    <Sheet open onOpenChange={(open) => !open && onClose()}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>New context</SheetTitle>
          <SheetDescription>
            Create an isolated development identity.
          </SheetDescription>
        </SheetHeader>
        <div className="space-y-4 px-8 pb-8">
          <label className="block text-sm">
            Start from a template
            <select className="mt-1 w-full border p-2" value={templateID} onChange={(e) => selectTemplate(e.target.value)}>
              {templates.map((template) => <option key={template.id} value={template.id}>{template.name}</option>)}
            </select>
          </label>
          <Field label="Name" value={name} onChange={setName} />
          <Field label="ID" value={contextId} onChange={setContextID} />
          <Field
            label="Description"
            value={description}
            onChange={setDescription}
          />
          <Field label="Icon" value={icon} onChange={setIcon} />
          <label className="block text-sm">
            Accent
            <select
              className="mt-1 w-full border p-2"
              value={accent}
              onChange={(e) => setAccent(e.target.value)}
            >
              {["sage", "slate-blue", "amber", "custom"].map((value) => (
                <option key={value}>{value}</option>
              ))}
            </select>
          </label>
          <label className="block text-sm">
            Coding tool
            <select
              className="mt-1 w-full border p-2"
              value={toolId}
              onChange={(e) => setToolID(e.target.value)}
            >
              {options.map((tool) => (
                <option key={tool.id} value={tool.id}>
                  {tool.name}
                </option>
              ))}
            </select>
          </label>
          <fieldset>
            <legend className="text-sm">Providers</legend>
            {providerOptions.map((provider) => (
              <label key={provider.id} className="block">
                <input
                  type="checkbox"
                  checked={providers.includes(provider.id)}
                  onChange={() =>
                    setProviders((current) =>
                      current.includes(provider.id)
                        ? current.filter((id) => id !== provider.id)
                        : [...current, provider.id],
                    )
                  }
                />{" "}
                {provider.name}
              </label>
            ))}
          </fieldset>
          {error ? <p className="text-destructive">{error.message}</p> : null}
          <Button
            type="button"
            disabled={pending || !name || !contextId || !toolId}
            onClick={() => void submit()}
          >
            {pending ? "Creating..." : "Create context"}
          </Button>
        </div>
      </SheetContent>
    </Sheet>
  );
}
function Field({
  label,
  value,
  onChange,
}: {
  label: string;
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <label className="block text-sm">
      {label}
      <input
        className="mt-1 w-full border p-2"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </label>
  );
}
