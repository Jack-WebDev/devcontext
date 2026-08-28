import { useEffect, useState } from "react";
import type {
  ApiResult,
  ContextDetailsState,
  ContextListItem,
  CreateContextResult,
  CreateContextRequest,
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
}: {
  contextId: string;
  onClose: () => void;
  load: (id: string) => Promise<ApiResult<ContextDetailsState>>;
}) {
  const [result, setResult] = useState<
    ApiResult<ContextDetailsState> | undefined
  >();
  useEffect(() => {
    void load(contextId).then(setResult);
  }, [contextId, load]);
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
}: {
  contexts: ContextListItem[];
  onClose: () => void;
  create: (
    request: CreateContextRequest,
  ) => Promise<ApiResult<CreateContextResult>>;
}) {
  const [name, setName] = useState("");
  const [contextId, setContextID] = useState("");
  const [description, setDescription] = useState("");
  const [icon, setIcon] = useState("");
  const [accent, setAccent] = useState("custom");
  const [toolId, setToolID] = useState(contexts[0]?.context.tool.id ?? "");
  const [providers, setProviders] = useState<string[]>([]);
  const [error, setError] = useState<DisplayError>();
  const [pending, setPending] = useState(false);
  const options = contexts[0]?.context.availableTools ?? [];
  const providerOptions = contexts
    .flatMap((c) => c.context.providers)
    .filter((p, i, all) => all.findIndex((x) => x.id === p.id) === i);
  async function submit() {
    setPending(true);
    setError(undefined);
    const request: CustomContextRequest = {
      contextId,
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
