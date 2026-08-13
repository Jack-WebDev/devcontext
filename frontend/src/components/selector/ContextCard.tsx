import type { KeyboardEvent, Ref } from "react";

import type { ContextState, ProviderState } from "../../lib/devctx-api";
import type { ContextNavigationDirection } from "./selection-state";

interface ContextCardProps {
  context: ContextState;
  selected?: boolean;
  disabled?: boolean;
  tabIndex?: number;
  buttonRef?: Ref<HTMLButtonElement>;
  onSelect?: (contextId: string) => void;
  onNavigate?: (contextId: string, direction: ContextNavigationDirection) => void;
  onLaunchSelected?: () => void;
}

function ContextCard({
  context,
  selected = false,
  disabled = false,
  tabIndex = 0,
  buttonRef,
  onSelect,
  onNavigate,
  onLaunchSelected,
}: ContextCardProps) {
  const interactive = onSelect !== undefined && !disabled;
  const unselectedClassName = interactive ? "border-border hover:border-foreground/20" : "border-border";
  const selectedClassName = selected
    ? "border-primary ring-2 ring-ring/40"
    : unselectedClassName;
  const focusClassName = onSelect
    ? "focus-within:border-ring focus-within:ring-2 focus-within:ring-ring/50"
    : "";
  const disabledClassName = disabled ? "opacity-60" : "";
  const className = `min-w-0 border bg-card p-5 text-left text-card-foreground shadow-sm transition-colors ${selectedClassName} ${focusClassName} ${disabledClassName}`;
  const enabledProviders = context.providers.filter((provider) => provider.enabled);
  const contextNameId = `context-${context.id}-name`;
  const contextSelectionId = `context-${context.id}-selection`;

  function handleButtonKeyDown(event: KeyboardEvent<HTMLButtonElement>) {
    const direction = contextNavigationDirectionForKey(event.key);
    if (direction !== undefined) {
      event.preventDefault();
      onNavigate?.(context.id, direction);
      return;
    }

    if (event.key === "Enter" && selected) {
      event.preventDefault();
      onLaunchSelected?.();
    }
  }

  return (
    <article
      className={className}
      aria-labelledby={contextNameId}
      aria-describedby={contextSelectionId}
      data-selected={selected ? "true" : undefined}
    >
      <div className="min-w-0 space-y-4">
        {onSelect ? (
          <button
            ref={buttonRef}
            type="button"
            className="block min-w-0 text-left outline-none focus-visible:outline-none disabled:cursor-not-allowed"
            aria-labelledby={contextNameId}
            aria-describedby={contextSelectionId}
            aria-pressed={selected}
            disabled={disabled}
            tabIndex={disabled ? undefined : tabIndex}
            onClick={() => onSelect(context.id)}
            onKeyDown={handleButtonKeyDown}
          >
            <ContextIdentity context={context} selected={selected} />
          </button>
        ) : (
          <ContextIdentity context={context} selected={selected} />
        )}

        {enabledProviders.length > 0 ? (
          <ul className="space-y-2">
            {enabledProviders.map((provider) => (
              <ProviderStatusRow key={provider.id} context={context} provider={provider} />
            ))}
          </ul>
        ) : null}
      </div>
    </article>
  );
}

function ContextIdentity({ context, selected }: { context: ContextState; selected: boolean }) {
  return (
    <div className="min-w-0 space-y-2">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <h3 id={`context-${context.id}-name`} className="truncate text-base font-semibold" title={context.name}>
          {context.name}
        </h3>
        <span
          id={`context-${context.id}-selection`}
          className={`shrink-0 border px-2 py-0.5 text-xs font-semibold ${
            selected
              ? "border-primary bg-primary text-primary-foreground"
              : "border-border bg-muted/30 text-muted-foreground"
          }`}
        >
          {selected ? "Selected" : "Not selected"}
        </span>
      </div>
      <p className="truncate font-mono text-xs text-muted-foreground" title={context.id}>
        {context.id}
      </p>
    </div>
  );
}

function ProviderStatusRow({ context, provider }: { context: ContextState; provider: ProviderState }) {
  const status = providerStatusPresentation(provider.state);
  const accessibleStatus = `${provider.name} local status: ${status.label}`;
  const authenticationGuidance = providerAuthenticationGuidance(context, provider);

  return (
    <li className="min-w-0 text-sm">
      <div className="flex min-w-0 items-start justify-between gap-3">
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
      </div>
      {authenticationGuidance ? (
        <p className="mt-2 border border-border bg-muted/30 p-2 text-xs text-muted-foreground">
          {authenticationGuidance}
        </p>
      ) : null}
    </li>
  );
}

function providerAuthenticationGuidance(context: ContextState, provider: ProviderState): string | undefined {
  if (provider.state !== "not_configured") {
    return undefined;
  }

  switch (provider.id) {
    case "claude":
    case "codex":
      return `${provider.name} is enabled for ${context.name} but is not signed in yet. Open ${context.name}, then sign in with ${provider.name} inside that tool. Dev Context will not copy credentials or ask for passwords or tokens.`;
    default:
      return undefined;
  }
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

function contextNavigationDirectionForKey(key: string): ContextNavigationDirection | undefined {
  if (key === "ArrowDown") {
    return "next";
  }

  if (key === "ArrowUp") {
    return "previous";
  }

  return undefined;
}

export { ContextCard };
