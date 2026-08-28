import type { TrustCenterState, TrustIsolationProtection } from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Card, CardContent } from "../ui/card.js";

function TrustCenterView({ state }: { state: TrustCenterState }) {
  return (
    <section aria-labelledby="trust-center-heading" className="space-y-6">
      <div>
        <p className="text-sm text-muted-foreground">Protection boundaries</p>
        <h2 id="trust-center-heading" className="text-2xl font-semibold">
          Trust Center
        </h2>
        <p className="mt-1 text-sm text-muted-foreground">
          Review the local protection state Dev Context can verify.
        </p>
      </div>

      <CredentialSyncCard enabled={state.credentialSync.enabled} message={state.credentialSync.message} />
      <TrustContextProtections contexts={state.contexts} />
      <TrustProjectMappings mappings={state.projectMappings} />
      <TrustIntegrationBoundaries boundaries={state.integrationBoundaries} />
    </section>
  );
}

function CredentialSyncCard({ enabled, message }: TrustCenterState["credentialSync"]) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="credential-sync-heading">
      <CardContent className="space-y-2 p-5">
        <div className="flex items-center justify-between gap-4">
          <h3 id="credential-sync-heading" className="font-semibold">Credential sync</h3>
          <StatusIndicator status={enabled ? "needs_attention" : "ready"} />
        </div>
        <p className="text-sm text-muted-foreground">{message}</p>
      </CardContent>
    </Card>
  );
}

function TrustContextProtections({ contexts }: { contexts: TrustCenterState["contexts"] }) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="context-protection-heading">
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="context-protection-heading" className="font-semibold">Context isolation</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Each enabled provider and selected coding tool has its own context-owned storage.
          </p>
        </div>
        {contexts.length === 0 ? (
          <p className="text-sm text-muted-foreground">No configured contexts to evaluate.</p>
        ) : (
          <div className="space-y-4">
            {contexts.map((context) => (
              <section key={context.id} className="border-t border-border pt-4" aria-label={`${context.name} protection`}>
                <h4 className="font-medium">{context.name}</h4>
                <div className="mt-3 space-y-3">
                  <IsolationRow label={context.tool.name} isolation={context.tool.isolation} />
                  {context.providers.map((provider) => (
                    <IsolationRow key={provider.id} label={provider.name} isolation={provider.isolation} />
                  ))}
                </div>
              </section>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

function IsolationRow({ label, isolation }: { label: string; isolation: TrustIsolationProtection }) {
  return (
    <div className="flex items-start justify-between gap-4 text-sm">
      <div className="min-w-0">
        <p className="font-medium">{label}</p>
        <p className="mt-1 text-muted-foreground">{isolation.message}</p>
      </div>
      <StatusIndicator status={isolation.status} />
    </div>
  );
}

function TrustProjectMappings({ mappings }: { mappings: TrustCenterState["projectMappings"] }) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="project-mappings-heading">
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="project-mappings-heading" className="font-semibold">Remembered project mappings</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            These explicit mappings suggest a context; they do not bypass launch safety checks.
          </p>
        </div>
        {mappings.length === 0 ? (
          <p className="text-sm text-muted-foreground">No projects are remembered.</p>
        ) : (
          <ul className="space-y-3">
            {mappings.map((mapping) => (
              <li key={mapping.project.path} className="border-t border-border pt-3 text-sm">
                <p className="font-medium">{mapping.project.name}</p>
                <p className="mt-1 break-all font-mono text-xs text-muted-foreground">{mapping.project.path}</p>
                <p className="mt-1 text-muted-foreground">Suggested context: {mapping.contextName}</p>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

function TrustIntegrationBoundaries({ boundaries }: { boundaries: TrustCenterState["integrationBoundaries"] }) {
  return (
    <Card as="section" hierarchy="secondary" className="py-0" aria-labelledby="integration-boundaries-heading">
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="integration-boundaries-heading" className="font-semibold">Coding-tool integration data</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Integration data stays within selected coding-tool storage.
          </p>
        </div>
        {boundaries.length === 0 ? (
          <p className="text-sm text-muted-foreground">No coding-tool integrations are configured.</p>
        ) : (
          <ul className="space-y-3">
            {boundaries.map((boundary) => (
              <li key={boundary.toolId} className="border-t border-border pt-3 text-sm">
                <div className="flex items-center justify-between gap-4">
                  <p className="font-medium">{boundary.toolName}</p>
                  <span className="text-xs text-muted-foreground">
                    {boundary.statusDataAvailable ? "Safe status data available" : "No exported integration data"}
                  </span>
                </div>
                <p className="mt-1 text-muted-foreground">{boundary.message}</p>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}

export { TrustCenterView, TrustIntegrationBoundaries, TrustProjectMappings };
