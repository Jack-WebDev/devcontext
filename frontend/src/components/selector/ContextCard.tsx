import type { ContextState, ProviderState } from "../../lib/devctx-api";

interface ContextCardProps {
  context: ContextState;
  selected?: boolean;
  disabled?: boolean;
  onSelect?: (contextId: string) => void;
}

function ContextCard({ context, selected = false, disabled = false, onSelect }: ContextCardProps) {
  const interactive = onSelect !== undefined && !disabled;
  const unselectedClassName = interactive ? "border-border hover:border-foreground/20" : "border-border";
  const selectedClassName = selected
    ? "border-primary ring-2 ring-ring/40"
    : unselectedClassName;
  const disabledClassName = disabled ? "opacity-60" : "";
  const className = `min-w-0 border bg-card p-5 text-left text-card-foreground shadow-sm transition-colors ${selectedClassName} ${disabledClassName}`;
  const enabledProviders = context.providers.filter((provider) => provider.enabled);

  return (
    <article
      className={className}
      aria-labelledby={`context-${context.id}-name`}
      data-selected={selected ? "true" : undefined}
    >
      <div className="min-w-0 space-y-4">
        {onSelect ? (
          <button
            type="button"
            className="block min-w-0 text-left"
            aria-labelledby={`context-${context.id}-name`}
            aria-pressed={selected}
            disabled={disabled}
            onClick={() => onSelect(context.id)}
          >
            <ContextIdentity context={context} />
          </button>
        ) : (
          <ContextIdentity context={context} />
        )}

        {enabledProviders.length > 0 ? (
          <ul className="space-y-2">
            {enabledProviders.map((provider) => (
              <ProviderStatusRow key={provider.id} provider={provider} />
            ))}
          </ul>
        ) : null}
      </div>
    </article>
  );
}

function ContextIdentity({ context }: { context: ContextState }) {
  return (
    <div className="min-w-0 space-y-2">
      <h3 id={`context-${context.id}-name`} className="truncate text-base font-semibold" title={context.name}>
        {context.name}
      </h3>
      <p className="truncate font-mono text-xs text-muted-foreground" title={context.id}>
        {context.id}
      </p>
    </div>
  );
}

function ProviderStatusRow({ provider }: { provider: ProviderState }) {
  const status = providerStatusPresentation(provider.state);
  const accessibleStatus = `${provider.name} local status: ${status.label}`;

  return (
    <li className="flex min-w-0 items-start justify-between gap-3 text-sm">
      <div className="min-w-0">
        <p className="truncate font-medium" title={provider.name}>
          {provider.name}
        </p>
        {provider.explanation ? (
          <p className="mt-1 truncate text-xs text-muted-foreground" title={provider.explanation}>
            {provider.explanation}
          </p>
        ) : null}
      </div>
      <div className="flex shrink-0 items-center gap-2">
        <span
          className={`size-2.5 rounded-full ${status.indicatorClassName}`}
          role="img"
          aria-label={accessibleStatus}
        />
        <span className="text-xs font-medium text-muted-foreground">{status.label}</span>
      </div>
    </li>
  );
}

function providerStatusPresentation(state: string): { label: string; indicatorClassName: string } {
  switch (state) {
    case "ready":
      return { label: "Ready", indicatorClassName: "bg-emerald-600" };
    case "not_configured":
      return { label: "Not configured", indicatorClassName: "bg-amber-500" };
    case "directory_missing":
      return { label: "Directory missing", indicatorClassName: "bg-orange-600" };
    case "unavailable":
      return { label: "Unavailable", indicatorClassName: "bg-destructive" };
    default:
      return { label: "Unknown", indicatorClassName: "bg-muted-foreground" };
  }
}

export { ContextCard };
