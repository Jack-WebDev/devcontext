import type { ProviderCredentialSession } from "../../lib/devctx-api";
import { Badge } from "../ui/badge.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

type ProviderSessionAssignments = Record<string, "personal" | "company" | undefined>;

interface ProviderCredentialClassificationProps {
  sessions: ProviderCredentialSession[];
  assignments: ProviderSessionAssignments;
  disabled?: boolean;
  onClassify: (providerId: string, contextId: "personal" | "company") => void;
}

function ProviderCredentialClassification({
  sessions,
  assignments,
  disabled = false,
  onClassify,
}: ProviderCredentialClassificationProps) {
  if (sessions.length === 0) {
    return null;
  }

  return (
    <section className="space-y-3" aria-label="Classify imported provider sessions">
      <div>
        <h4 className="text-sm font-semibold">Classify detected provider sessions</h4>
        <p className="mt-1 text-sm text-muted-foreground">
          Dev Context found currently signed-in local sessions. Review the metadata and choose which context should
          receive each session.
        </p>
      </div>
      <div className="grid gap-3 sm:grid-cols-2">
        {sessions.map((session) => (
          <ProviderCredentialSessionCard
            key={session.providerId}
            session={session}
            assignment={assignments[session.providerId]}
            disabled={disabled}
            onClassify={onClassify}
          />
        ))}
      </div>
    </section>
  );
}

function ProviderCredentialSessionCard({
  session,
  assignment,
  disabled,
  onClassify,
}: {
  session: ProviderCredentialSession;
  assignment?: "personal" | "company";
  disabled: boolean;
  onClassify: (providerId: string, contextId: "personal" | "company") => void;
}) {
  return (
    <Card as="article" size="sm" className="border border-border py-0">
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h5 className="text-sm font-semibold">{session.name}</h5>
            <p className="mt-1 text-xs text-muted-foreground">{providerSessionSourceLabel(session)}</p>
            <ProviderCredentialMetadata session={session} />
          </div>
          <Badge variant={assignment === undefined ? "secondary" : "default"} className="text-xs font-medium">
            {assignment === undefined
              ? "Unassigned"
              : assignment === "personal"
                ? "Will import to Personal"
                : "Will import to Company"}
          </Badge>
        </div>
        <div className="mt-4 flex flex-wrap gap-2" role="group" aria-label={`Classify ${session.name} session`}>
          <Button
            type="button"
            variant={assignment === "personal" ? "default" : "outline"}
            size="sm"
            disabled={disabled}
            onClick={() => onClassify(session.providerId, "personal")}
          >
            Import to Personal
          </Button>
          <Button
            type="button"
            variant={assignment === "company" ? "default" : "outline"}
            size="sm"
            disabled={disabled}
            onClick={() => onClassify(session.providerId, "company")}
          >
            Import to Company
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

function ProviderCredentialMetadata({ session }: { session: ProviderCredentialSession }) {
  const rows = providerCredentialRows(session);

  if (!session.metadataAvailable || rows.length === 0) {
    return (
      <p className="mt-2 text-sm text-muted-foreground">
        Account metadata unavailable. Refresh this provider sign-in, then reopen Dev Context to identify the session.
      </p>
    );
  }

  return (
    <dl className="mt-2 space-y-1 text-sm">
      {rows.map((row) => (
        <div key={row.label} className="grid grid-cols-[auto_1fr] gap-x-2">
          <dt className="text-muted-foreground">{row.label}:</dt>
          <dd className="min-w-0 truncate font-medium text-foreground" title={row.value}>
            {row.value}
          </dd>
        </div>
      ))}
    </dl>
  );
}

function providerCredentialRows(session: ProviderCredentialSession): Array<{ label: string; value: string }> {
  if (session.providerId === "codex" && session.codex) {
    return [
      { label: "Email", value: session.codex.email },
      { label: "Plan", value: session.codex.chatgptPlanType },
      { label: "Account", value: session.codex.chatgptAccountId },
    ].filter(rowHasValue);
  }

  if (session.providerId === "claude" && session.claude) {
    return [
      { label: "Subscription", value: session.claude.subscriptionType },
      { label: "Organization", value: session.claude.organizationName },
      { label: "Organization ID", value: session.claude.organizationUuid },
    ].filter(rowHasValue);
  }

  return [];
}

function providerSessionSourceLabel(session: ProviderCredentialSession): string {
  switch (session.providerId) {
    case "codex":
      return "Current global Codex session";
    case "claude":
      return "Current global Claude session";
    default:
      return "Current global provider session";
  }
}

function rowHasValue(row: { label: string; value?: string }): row is { label: string; value: string } {
  return row.value !== undefined && row.value !== "";
}

export { ProviderCredentialClassification };
export type { ProviderSessionAssignments };
