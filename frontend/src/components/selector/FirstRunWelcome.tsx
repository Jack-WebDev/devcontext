import type { LaunchState } from "../../lib/devctx-api";

interface FirstRunWelcomeProps {
  launchState: LaunchState;
  onCreatePersonal?: () => void;
  onCreateCompany?: () => void;
}

function FirstRunWelcome({ launchState, onCreatePersonal, onCreateCompany }: FirstRunWelcomeProps) {
  return (
    <section aria-labelledby="first-run-title" className="space-y-6">
      <div className="space-y-2">
        <h3 id="first-run-title" className="text-xl font-semibold">
          Create your first development context
        </h3>
        <p className="max-w-2xl text-sm text-muted-foreground">
          Dev Context creates local identities for this machine so each project can open with the right editor state,
          provider folders, and authentication setup.
        </p>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <GuidanceItem
          title="Local first"
          description="Context files live under your local Dev Context home. Nothing is synced or uploaded by Dev Context."
        />
        <GuidanceItem
          title="Isolated tools"
          description="Each context gets separate editor and provider directories so project work does not share local state accidentally."
        />
        <GuidanceItem
          title="Separate authentication"
          description="Sign in inside each provider yourself. Dev Context does not store passwords, tokens, or cloud accounts."
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2" role="group" aria-label="Create a default context">
        <OnboardingAction
          title="Personal"
          description="Use this for personal repositories, experiments, and tools tied to your own accounts."
          buttonLabel="Create Personal"
          disabled={onCreatePersonal === undefined}
          onClick={onCreatePersonal}
        />
        <OnboardingAction
          title="Company"
          description="Use this for work repositories and tools tied to employer or client accounts."
          buttonLabel="Create Company"
          disabled={onCreateCompany === undefined}
          onClick={onCreateCompany}
        />
      </div>

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

function shouldRenderFirstRunWelcome(launchState: LaunchState): boolean {
  return launchState.firstRun;
}

export { FirstRunWelcome, shouldRenderFirstRunWelcome };
