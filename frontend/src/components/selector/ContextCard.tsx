import type { KeyboardEvent, Ref } from "react";

import type { ContextState, LaunchConfidenceStatus, ProviderState } from "../../lib/devctx-api";
import { Badge } from "../ui/badge.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
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
  const className = `min-w-0 border py-0 text-left transition-colors ${selectedClassName} ${focusClassName} ${disabledClassName}`;
  const enabledProviders = context.providers.filter((provider) => provider.enabled);
  const contextNameId = `context-${context.id}-name`;
  const contextDescription = context.metadata?.description;
  const contextAccent = contextAccentName(context.metadata?.accent);

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
    <Card
      as="article"
      size="sm"
      className={`relative ${className}`}
      aria-labelledby={contextNameId}
      data-selected={selected ? "true" : undefined}
      data-context-accent={contextAccent}
    >
      {onSelect ? (
        <Button
          ref={buttonRef}
          type="button"
          variant="ghost"
          className="absolute inset-0 z-10 h-auto w-full p-0 focus-visible:ring-0 disabled:cursor-not-allowed"
          aria-labelledby={contextNameId}
          aria-pressed={selected}
          disabled={disabled}
          tabIndex={disabled ? undefined : tabIndex}
          onClick={() => onSelect(context.id)}
          onKeyDown={handleButtonKeyDown}
        >
          <span className="sr-only">Select {context.name}</span>
        </Button>
      ) : null}
      <CardContent className="pointer-events-none min-w-0 space-y-4 p-5">
        <ContextIdentity context={context} description={contextDescription} accent={contextAccent} selected={selected} />
        <ToolStatusRow context={context} />
        <ProviderSummary context={context} providers={enabledProviders} />
      </CardContent>
    </Card>
  );
}

function ProviderSummary({ context, providers }: { context: ContextState; providers: ProviderState[] }) {
  return (
    <section aria-label={`Enabled providers for ${context.name}`}>
      <p className="mb-2 text-xs font-semibold tracking-wide text-muted-foreground uppercase">Enabled providers</p>
      {providers.length > 0 ? (
        <ul className="space-y-2">
          {providers.map((provider) => (
            <ProviderStatusRow key={provider.id} context={context} provider={provider} />
          ))}
        </ul>
      ) : (
        <p className="text-sm text-muted-foreground">No providers enabled.</p>
      )}
    </section>
  );
}

function ToolStatusRow({ context }: { context: ContextState }) {
  const status = toolStatusPresentation(context.tool.status);

  return (
    <div className="min-w-0 border border-border bg-muted/30 p-3 text-sm">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate font-medium" title={context.tool.name}>
            {context.tool.name}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">Coding tool</p>
        </div>
        <Badge variant="secondary" className={`shrink-0 text-xs font-medium normal-case tracking-normal ${status.className}`}>
          {status.label}
        </Badge>
      </div>
      <p className="mt-2 text-xs text-muted-foreground">{context.tool.message}</p>
      {context.tool.actionHint ? (
        <p className="mt-2 border border-border bg-background p-2 text-xs text-muted-foreground">{context.tool.actionHint}</p>
      ) : null}
    </div>
  );
}

function ContextIdentity({
  context,
  description,
  accent,
  selected,
}: {
  context: ContextState;
  description?: string;
  accent: ContextAccentName;
  selected: boolean;
}) {
  return (
    <div className="min-w-0 space-y-2">
      <div className="flex min-w-0 items-start gap-3">
        <span className={`mt-1.5 size-2 shrink-0 ${contextAccentClassName(accent)}`} aria-hidden="true" />
        <div className="min-w-0 flex-1">
          <div className="flex min-w-0 items-start justify-between gap-3">
            <h3 id={`context-${context.id}-name`} className="truncate text-base font-semibold" title={context.name}>
              {context.name}
            </h3>
            <Badge
              variant={selected ? "default" : "secondary"}
              className={`shrink-0 border px-2 py-0.5 text-xs font-semibold ${
                selected
                  ? "border-primary bg-primary text-primary-foreground"
                  : "border-border bg-muted/30 text-muted-foreground"
              }`}
            >
              {selected ? "Selected" : "Not selected"}
            </Badge>
          </div>
          {description ? (
            <p className="mt-1 truncate text-sm text-muted-foreground" title={description}>
              {description}
            </p>
          ) : null}
        </div>
      </div>
      <p className="truncate font-mono text-xs text-muted-foreground" title={context.id}>
        {context.id}
      </p>
    </div>
  );
}

type ContextAccentName = "sage" | "slate-blue" | "amber" | "custom" | "neutral";

function contextAccentName(value: string | undefined): ContextAccentName {
  switch (value) {
    case "sage":
    case "slate-blue":
    case "amber":
    case "custom":
      return value;
    default:
      return "neutral";
  }
}

function contextAccentClassName(accent: ContextAccentName): string {
  switch (accent) {
    case "sage":
      return "bg-emerald-600";
    case "slate-blue":
      return "bg-blue-700";
    case "amber":
      return "bg-amber-500";
    case "custom":
      return "bg-violet-600";
    default:
      return "bg-muted-foreground";
  }
}

function ProviderStatusRow({ context, provider }: { context: ContextState; provider: ProviderState }) {
  const status = providerStatusPresentation(provider.state);
  const accessibleStatus = `${provider.name} local status: ${status.label}`;
  const setupGuidance = providerSetupGuidance(context, provider);

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
          <Badge variant="secondary" className="text-xs font-medium normal-case tracking-normal">
            {status.label}
          </Badge>
        </div>
      </div>
      {setupGuidance ? (
        <p className="mt-2 border border-border bg-muted/30 p-2 text-xs text-muted-foreground">
          {setupGuidance}
        </p>
      ) : null}
    </li>
  );
}

function providerSetupGuidance(context: ContextState, provider: ProviderState): string | undefined {
  if (provider.state !== "not_configured") {
    return undefined;
  }

	return provider.actionHint ?? `${provider.name} is enabled for ${context.name} but is not configured. Open ${context.name} and complete ${provider.name} setup.`;
}

function providerStatusPresentation(state: ProviderState["state"]): { label: string; indicatorClassName: string } {
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

function toolStatusPresentation(status: LaunchConfidenceStatus): { label: string; className: string } {
  switch (status) {
    case "ready":
      return { label: "Ready", className: "text-emerald-700" };
    case "needs_attention":
      return { label: "Needs attention", className: "text-amber-700" };
    case "blocked":
      return { label: "Blocked", className: "text-destructive" };
    default:
      return { label: "Needs attention", className: "text-muted-foreground" };
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
