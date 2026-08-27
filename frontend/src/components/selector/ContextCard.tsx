import type { KeyboardEvent, Ref } from "react";

import type { ContextState, LaunchConfidenceStatus, ProviderState } from "../../lib/devctx-api";
import { Badge } from "../ui/badge.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import type { ContextNavigationDirection } from "./selection-state";

interface ContextCardProps {
  context: ContextState;
  selected?: boolean;
  recommendation?: string;
  disabled?: boolean;
  tabIndex?: number;
  buttonRef?: Ref<HTMLButtonElement>;
  onSelect?: (contextId: string) => void;
  onNavigate?: (contextId: string, direction: ContextNavigationDirection) => void;
  onLaunchSelected?: () => void;
  onProviderSetup?: (contextId: string, providerId: string) => void;
}

function ContextCard({
  context,
  selected = false,
  recommendation,
  disabled = false,
  tabIndex = 0,
  buttonRef,
  onSelect,
  onNavigate,
  onLaunchSelected,
  onProviderSetup,
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
  const contextDescription = context.description;
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
      {selected ? (
        <span
          className="pointer-events-none absolute inset-y-0 left-0 z-20 w-1 bg-primary"
          data-context-selection-marker
          aria-hidden="true"
        />
      ) : null}
      {onSelect ? (
        <Button
          ref={buttonRef}
          type="button"
          variant="ghost"
          className="absolute inset-0 z-10 h-auto w-full p-0 focus-visible:ring-0 disabled:cursor-not-allowed"
          aria-labelledby={contextNameId}
          aria-current={selected ? "true" : undefined}
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
        <ContextIdentity
          context={context}
          description={contextDescription}
          accent={contextAccent}
          selected={selected}
          recommendation={recommendation}
        />
        <ToolStatusRow context={context} />
        <ProviderSummary context={context} providers={enabledProviders} onProviderSetup={onProviderSetup} />
        <ContextHealthSummary context={context} enabledProviderCount={enabledProviders.length} />
      </CardContent>
    </Card>
  );
}

function ProviderSummary({
  context,
  providers,
  onProviderSetup,
}: {
  context: ContextState;
  providers: ProviderState[];
  onProviderSetup?: (contextId: string, providerId: string) => void;
}) {
  return (
    <section aria-label={`Enabled providers for ${context.name}`}>
      <p className="mb-2 text-xs font-semibold tracking-wide text-muted-foreground uppercase">Enabled providers</p>
      {providers.length > 0 ? (
        <ul className="space-y-2">
          {providers.map((provider) => (
            <ProviderStatusRow key={provider.id} context={context} provider={provider} onProviderSetup={onProviderSetup} />
          ))}
        </ul>
      ) : (
        <p className="text-sm text-muted-foreground">No providers enabled.</p>
      )}
    </section>
  );
}

function ContextHealthSummary({
  context,
  enabledProviderCount,
}: {
  context: ContextState;
  enabledProviderCount: number;
}) {
  if (context.confidence === undefined) {
    return null;
  }

  const providerChecks = context.confidence.checks.filter((check) => check.component === "provider");
  const toolCheck = context.confidence.checks.find(
    (check) => check.component === "tool" && check.toolId === context.tool.id,
  );
  const isolationChecks = context.confidence.checks.filter((check) => check.component === "isolation");
  const providerStatus = mostSevereStatus(providerChecks.map((check) => check.severity));
  const isolationStatus = mostSevereStatus(isolationChecks.map((check) => check.severity));

  if (providerStatus === undefined && toolCheck === undefined && isolationStatus === undefined) {
    return null;
  }

  return (
    <section className="border-t border-border pt-3" aria-label={`Context health for ${context.name}`}>
      <p className="mb-2 text-xs font-semibold tracking-wide text-muted-foreground uppercase">Context health</p>
      <dl className="space-y-2 text-sm">
        {providerStatus ? (
          <HealthSummaryItem
            label={enabledProviderCount === 1 ? "1 provider" : `${enabledProviderCount} providers`}
            status={providerStatus}
          />
        ) : null}
        {toolCheck ? <HealthSummaryItem label={context.tool.name} status={toolCheck.severity} /> : null}
        {isolationStatus ? <HealthSummaryItem label="Isolation" status={isolationStatus} isolation /> : null}
      </dl>
    </section>
  );
}

function HealthSummaryItem({
  label,
  status,
  isolation = false,
}: {
  label: string;
  status: LaunchConfidenceStatus;
  isolation?: boolean;
}) {
  const presentation = healthStatusPresentation(status, isolation);

  return (
    <div className="flex items-center justify-between gap-3">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className={`font-medium ${presentation.className}`}>{presentation.label}</dd>
    </div>
  );
}

function mostSevereStatus(statuses: LaunchConfidenceStatus[]): LaunchConfidenceStatus | undefined {
  if (statuses.includes("blocked")) {
    return "blocked";
  }
  if (statuses.includes("needs_attention")) {
    return "needs_attention";
  }
  if (statuses.includes("ready")) {
    return "ready";
  }
  return undefined;
}

function healthStatusPresentation(
  status: LaunchConfidenceStatus,
  isolation: boolean,
): { label: string; className: string } {
  if (isolation && status === "ready") {
    return {label: "Protected", className: "text-emerald-700"};
  }

  return toolStatusPresentation(status);
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
  recommendation,
}: {
  context: ContextState;
  description?: string;
  accent: ContextAccentName;
  selected: boolean;
  recommendation?: string;
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
            <div className="flex shrink-0 items-center gap-2">
              {recommendation ? (
                <Badge
                  variant="secondary"
                  className="border border-foreground/20 bg-muted/50 px-2 py-0.5 text-xs font-semibold text-foreground"
                  title={recommendation}
                >
                  Recommended
                </Badge>
              ) : null}
              <Badge
                variant={selected ? "default" : "secondary"}
                className={`border px-2 py-0.5 text-xs font-semibold ${
                  selected
                    ? "border-primary bg-primary text-primary-foreground"
                    : "border-border bg-muted/30 text-muted-foreground"
                }`}
              >
                {selected ? "Selected" : "Not selected"}
              </Badge>
            </div>
          </div>
          {description ? (
            <p className="mt-1 truncate text-sm text-muted-foreground" title={description}>
              {description}
            </p>
          ) : null}
          {recommendation ? <p className="mt-2 text-xs text-muted-foreground">{recommendation}</p> : null}
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

function ProviderStatusRow({
  context,
  provider,
  onProviderSetup,
}: {
  context: ContextState;
  provider: ProviderState;
  onProviderSetup?: (contextId: string, providerId: string) => void;
}) {
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
          <ProviderIdentityLine provider={provider} />
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
      {provider.setupAction?.state === "open_and_configure" ? (
        <div className="pointer-events-auto relative z-20 mt-2 border border-border bg-muted/30 p-2">
          <p className="text-xs text-muted-foreground">{provider.setupAction.message}</p>
          <Button
            type="button"
            variant="outline"
            size="sm"
            className="mt-2"
            disabled={onProviderSetup === undefined}
            onClick={() => onProviderSetup?.(context.id, provider.id)}
          >
            {provider.setupAction.label}
          </Button>
        </div>
      ) : null}
    </li>
  );
}

function ProviderIdentityLine({ provider }: { provider: ProviderState }) {
  const { identity } = provider;

  if (identity.status === "verified" && identity.fields.length > 0) {
    const details = identity.fields.map((field) => `${field.label}: ${field.value}`).join(" · ");
    return <p className="mt-1 truncate text-xs text-muted-foreground" title={details}>Account: {details}</p>;
  }

  if (identity.status === "mismatch_evidence" && identity.message) {
    return <p className="mt-1 text-xs text-muted-foreground">{identity.message}</p>;
  }

  return <p className="mt-1 text-xs text-muted-foreground">Account identity unavailable</p>;
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
