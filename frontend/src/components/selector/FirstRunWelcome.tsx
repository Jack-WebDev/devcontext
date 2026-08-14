import type { DisplayError, LaunchState, ProviderCredentialSession } from "../../lib/devctx-api";
import { ProviderCredentialClassification, type ProviderSessionAssignments } from "./ProviderCredentialClassification.js";

interface FirstRunWelcomeProps {
  launchState: LaunchState;
  providerCredentialSessions?: ProviderCredentialSession[];
  providerSessionAssignments?: ProviderSessionAssignments;
  onCreatePersonal?: () => void;
  onCreateCompany?: () => void;
  onClassifyProviderSession?: (providerId: string, contextId: "personal" | "company") => void;
  pendingContextId?: string;
  error?: DisplayError;
}

function FirstRunWelcome({
  launchState,
  providerCredentialSessions = [],
  providerSessionAssignments = {},
  onCreatePersonal,
  onCreateCompany,
  onClassifyProviderSession,
  pendingContextId,
  error,
}: FirstRunWelcomeProps) {
  const pending = pendingContextId !== undefined;
  const canClassify = providerCredentialSessions.length === 0 || onClassifyProviderSession !== undefined;
  const handleClassifyProviderSession = onClassifyProviderSession ?? (() => {});
  const classificationComplete = providerCredentialSessions.every(
    (session) => providerSessionAssignments[session.providerId] !== undefined,
  );

  return (
    <section aria-labelledby="first-run-title" className="space-y-6">
      <div className="space-y-2">
        <h3 id="first-run-title" className="text-xl font-semibold">
          Create your first development context
        </h3>
        <p className="max-w-2xl text-sm text-muted-foreground">
          Dev Context creates local identities for this machine so each project can open with the right provider
          folders and authentication setup.
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <GuidanceItem
          title="Local first"
          description="Context files live under your local Dev Context home. Nothing is synced or uploaded by Dev Context."
        />
        <GuidanceItem
          title="Isolated tools"
          description="Each context gets separate Codex and Claude directories while VS Code keeps your normal profile."
        />
        <GuidanceItem
          title="Separate authentication"
          description="Classify detected sessions before import, or sign in inside each provider yourself. Dev Context does not store passwords, tokens, or cloud accounts."
        />
      </div>

      <ProviderCredentialClassification
        sessions={providerCredentialSessions}
        assignments={providerSessionAssignments}
        disabled={pending}
        onClassify={handleClassifyProviderSession}
      />

      <div className="grid gap-4 sm:grid-cols-2" role="group" aria-label="Create a default context">
        <OnboardingAction
          title="Personal"
          description="Use this for personal repositories, experiments, and tools tied to your own accounts."
          buttonLabel={pendingContextId === "personal" ? "Creating..." : "Create Personal"}
          disabled={pending || !canClassify || !classificationComplete || onCreatePersonal === undefined}
          onClick={onCreatePersonal}
        />
        <OnboardingAction
          title="Company"
          description="Use this for work repositories and tools tied to employer or client accounts."
          buttonLabel={pendingContextId === "company" ? "Creating..." : "Create Company"}
          disabled={pending || !canClassify || !classificationComplete || onCreateCompany === undefined}
          onClick={onCreateCompany}
        />
      </div>

      {pendingContextId ? (
        <p className="border border-border bg-muted/30 p-3 text-sm text-muted-foreground" role="status">
          Creating {pendingContextId === "personal" ? "Personal" : "Company"} context...
        </p>
      ) : null}

      {error ? <FirstRunErrorNotice error={error} /> : null}

      <p className="border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
        Current project: <span className="font-medium text-foreground">{launchState.project.path}</span>
      </p>
    </section>
  );
}

function GuidanceItem({ title, description }: { title: string; description: string }) {
  return (
    <div className="border border-border bg-card p-4">
      <h4 className="text-sm font-semibold">{title}</h4>
      <p className="mt-2 text-sm text-muted-foreground">{description}</p>
    </div>
  );
}

function OnboardingAction({
  title,
  description,
  buttonLabel,
  disabled,
  onClick,
}: {
  title: string;
  description: string;
  buttonLabel: string;
  disabled: boolean;
  onClick?: () => void;
}) {
  return (
    <div className="border border-border bg-card p-5">
      <div className="space-y-2">
        <h4 className="text-base font-semibold">{title}</h4>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <button
        type="button"
        className="mt-4 bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled}
        onClick={onClick}
      >
        {buttonLabel}
      </button>
    </div>
  );
}

function FirstRunErrorNotice({ error }: { error: DisplayError }) {
  return (
    <div className="border border-destructive/40 bg-destructive/5 p-4 text-sm" role="alert">
      <p className="font-medium text-destructive">{error.message}</p>
      <p className="mt-1 text-muted-foreground">{error.recovery}</p>
    </div>
  );
}

function shouldRenderFirstRunWelcome(launchState: LaunchState): boolean {
  return launchState.firstRun;
}

export { FirstRunWelcome, shouldRenderFirstRunWelcome };
