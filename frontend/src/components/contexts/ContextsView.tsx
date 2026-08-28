import type { ContextListItem } from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ContextsViewProps {
  contexts: ContextListItem[];
}

function ContextsView({contexts}: ContextsViewProps) {
  return (
    <section aria-labelledby="contexts-heading" className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm text-muted-foreground">Development identities</p>
          <h2 id="contexts-heading" className="text-2xl font-semibold">Contexts</h2>
        </div>
        <Button type="button" disabled title="Context creation will be available in a later update.">
          New context
        </Button>
      </div>

      {contexts.length === 0 ? (
        <Card as="section" hierarchy="secondary" className="py-0">
          <CardContent className="p-5 text-sm text-muted-foreground">
            No contexts are configured yet. Create a context to set up an isolated development identity.
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 lg:grid-cols-2" aria-label="Configured contexts">
          {contexts.map((item) => <ContextListCard key={item.context.id} item={item} />)}
        </div>
      )}
    </section>
  );
}

function ContextListCard({item}: {item: ContextListItem}) {
  const {context} = item;
  return (
    <Card as="article" hierarchy="secondary" className="py-0" aria-labelledby={`context-${context.id}-heading`}>
      <CardContent className="space-y-4 p-5">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h3 id={`context-${context.id}-heading`} className="truncate text-lg font-semibold" title={context.name}>
              {context.name}
            </h3>
            {context.description ? <p className="mt-1 text-sm text-muted-foreground">{context.description}</p> : null}
          </div>
          <StatusIndicator status={context.confidence?.status ?? "blocked"} />
        </div>

        <dl className="grid gap-3 border-t border-border pt-4 text-sm sm:grid-cols-2">
          <ContextListDetail label="Coding tool" value={context.tool.name} />
          <ContextListDetail label="Projects" value={projectCountLabel(item.projectCount)} />
          <ContextListDetail label="Providers" value={providerSummary(item)} />
          <ContextListDetail label="Last used" value={formatContextTime(item.lastUsedAt)} />
        </dl>
      </CardContent>
    </Card>
  );
}

function ContextListDetail({label, value}: {label: string; value: string}) {
  return (
    <div className="min-w-0">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="mt-1 truncate font-medium" title={value}>{value}</dd>
    </div>
  );
}

function providerSummary(item: ContextListItem): string {
  if (item.enabledProviders.length === 0) {
    return "No providers enabled";
  }
  return item.enabledProviders.map((provider) => provider.name).join(", ");
}

function projectCountLabel(count: number): string {
  return `${count} ${count === 1 ? "project" : "projects"}`;
}

function formatContextTime(value: string | undefined): string {
  if (value === undefined) {
    return "Never launched";
  }
  const time = new Date(value);
  return Number.isNaN(time.getTime()) ? "Unavailable" : time.toLocaleString();
}

export { ContextsView, formatContextTime, projectCountLabel, providerSummary };
