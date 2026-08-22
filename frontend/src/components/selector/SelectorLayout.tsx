import type { ReactNode } from "react";

interface SelectorLayoutProps {
  projectIdentity: ReactNode;
  contextCards: ReactNode;
  confidenceSummary: ReactNode;
  rememberControl: ReactNode;
  launchActions: ReactNode;
}

function SelectorLayout({
  projectIdentity,
  contextCards,
  confidenceSummary,
  rememberControl,
  launchActions,
}: SelectorLayoutProps) {
  return (
    <div className="space-y-6" data-selector-layout="context-selector">
      <div data-selector-layout-section="project-identity">{projectIdentity}</div>
      <section className="space-y-3" aria-label="Context choices" data-selector-layout-section="context-cards">
        {contextCards}
      </section>
      <section aria-label="Confidence summary" data-selector-layout-section="confidence-summary">
        {confidenceSummary}
      </section>
      <section aria-label="Project remembering" data-selector-layout-section="remember-control">
        {rememberControl}
      </section>
      <section aria-label="Launch actions" data-selector-layout-section="launch-actions">
        {launchActions}
      </section>
    </div>
  );
}

export { SelectorLayout };
