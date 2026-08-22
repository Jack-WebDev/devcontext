import { type KeyboardEvent, useEffect, useRef, useState } from "react";

import type {
  ApiResult,
  BindProjectRequest,
  CreateContextResult,
  DisplayError,
  LaunchProjectRequest,
  LaunchProjectResult,
  LaunchState,
  PreflightLaunchProjectRequest,
  PreflightLaunchProjectResult,
  ProviderCredentialSession,
  ProjectBindingState,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
import { ContextCard } from "./ContextCard";
import { FirstRunWelcome, shouldRenderFirstRunWelcome } from "./FirstRunWelcome";
import { GuiErrorNotice } from "./GuiErrorNotice";
import { ProviderCredentialClassification, type ProviderSessionAssignments } from "./ProviderCredentialClassification.js";
import { ProjectIdentity } from "./ProjectIdentity";
import { RememberProjectControl } from "./RememberProjectControl";
import { SelectorActions } from "./SelectorActions";
import { SelectorConfidenceSummary } from "./SelectorConfidenceSummary";
import { SelectorLayout } from "./SelectorLayout";
import { cancelSelector } from "./cancel-action";
import { missingDefaultContextIds } from "./default-context-actions";
import { createLaunchRequestGuard, launchSelectedContext } from "./launch-action";
import {
  initialRovingContextId,
  initialSelectedContextId,
  nextKeyboardContextId,
  nextSelectedContextId,
  type ContextNavigationDirection,
} from "./selection-state";
import { canLaunchSelectedContextFromKeyboard, escapeKeyboardAction } from "./selector-keyboard";

interface SelectorViewProps {
  launchState: LaunchState;
  onBindProject: (request: BindProjectRequest) => Promise<ApiResult<ProjectBindingState>>;
  onPreflightLaunchProject: (request: PreflightLaunchProjectRequest) => Promise<ApiResult<PreflightLaunchProjectResult>>;
  onLaunchProject: (request: LaunchProjectRequest) => Promise<ApiResult<LaunchProjectResult>>;
  onCancel: () => Promise<void> | void;
  onCreatePersonalContext?: (importProviderIds: string[]) => Promise<ApiResult<CreateContextResult>>;
  onCreateCompanyContext?: (importProviderIds: string[]) => Promise<ApiResult<CreateContextResult>>;
}

function SelectorView({
  launchState,
  onBindProject,
  onPreflightLaunchProject,
  onLaunchProject,
  onCancel,
  onCreatePersonalContext,
  onCreateCompanyContext,
}: SelectorViewProps) {
  const [selectedContextId, setSelectedContextId] = useState<string | undefined>(() =>
    initialSelectedContextId(launchState),
  );
  const [rovingContextId, setRovingContextId] = useState<string | undefined>(() =>
    initialRovingContextId(launchState),
  );
  const [rememberProject, setRememberProject] = useState(false);
  const [launchPending, setLaunchPending] = useState(false);
  const [launchError, setLaunchError] = useState<DisplayError | undefined>(undefined);
  const [mismatchError, setMismatchError] = useState<DisplayError | undefined>(undefined);
  const [onboardingPendingContextId, setOnboardingPendingContextId] = useState<string | undefined>(undefined);
  const [onboardingError, setOnboardingError] = useState<DisplayError | undefined>(undefined);
  const [providerSessionAssignments, setProviderSessionAssignments] = useState<ProviderSessionAssignments>({});
  const contextButtonRefs = useRef(new Map<string, HTMLButtonElement>());
  const launchGuard = useRef(createLaunchRequestGuard());
  const mismatchDialogOpen = mismatchError?.contextMismatch !== undefined;
  const selectedContext = launchState.contexts.find((context) => context.id === selectedContextId);
  const keyboardLaunchAvailable = canLaunchSelectedContextFromKeyboard({
    selectedContextId,
    launchPending,
    mismatchDialogOpen,
  });

  useEffect(() => {
    setSelectedContextId(initialSelectedContextId(launchState));
    setRovingContextId(initialRovingContextId(launchState));
    setRememberProject(false);
    setLaunchPending(false);
    setLaunchError(undefined);
    setMismatchError(undefined);
    setOnboardingPendingContextId(undefined);
    setOnboardingError(undefined);
    setProviderSessionAssignments({});
    launchGuard.current = createLaunchRequestGuard();
  }, [launchState]);

  function setContextButtonRef(contextId: string) {
    return (button: HTMLButtonElement | null) => {
      if (button) {
        contextButtonRefs.current.set(contextId, button);
      } else {
        contextButtonRefs.current.delete(contextId);
      }
    };
  }

  function handleSelectContext(contextId: string) {
    const nextContextId = nextSelectedContextId(launchState.contexts, contextId);
    setSelectedContextId(nextContextId);
    setRovingContextId(nextContextId);
  }

  function handleContextNavigation(contextId: string, direction: ContextNavigationDirection) {
    if (launchPending) {
      return;
    }

    const nextContextId = nextKeyboardContextId(launchState.contexts, contextId, direction);
    if (nextContextId === undefined) {
      return;
    }

    setSelectedContextId(nextContextId);
    setRovingContextId(nextContextId);
    contextButtonRefs.current.get(nextContextId)?.focus();
  }

  function handleSelectorKeyDown(event: KeyboardEvent<HTMLDivElement>) {
    if (event.key !== "Escape") {
      return;
    }

    const action = escapeKeyboardAction({
      selectedContextId,
      launchPending,
      mismatchDialogOpen,
    });
    if (action === "none") {
      return;
    }

    event.preventDefault();
    event.stopPropagation();

    if (action === "close-dialog") {
      setMismatchError(undefined);
      return;
    }

    void cancelSelector({ closeSelector: onCancel });
  }

  async function handleLaunch(confirmContextMismatch = false) {
    await launchGuard.current.run(async () => {
      setLaunchPending(true);
      setLaunchError(undefined);
      if (confirmContextMismatch) {
        setMismatchError(undefined);
      }

      try {
        const result = await launchSelectedContext({
          projectPath: launchState.project.path,
          selectedContextId,
          rememberProject,
          confirmContextMismatch,
          bindProject: onBindProject,
          preflightLaunchProject: onPreflightLaunchProject,
          launchProject: onLaunchProject,
        });

        if (result && !result.ok) {
          if (result.error.code === "context_mismatch_requires_confirmation" && result.error.contextMismatch) {
            setMismatchError(result.error);
          } else {
            setLaunchError(result.error);
          }
        }
      } finally {
        setLaunchPending(false);
      }
    });
  }

  function handleClassifyProviderSession(providerId: string, contextId: "personal" | "company") {
    setProviderSessionAssignments((current) => ({
      ...current,
      [providerId]: contextId,
    }));
  }

  function importProviderIdsForContext(contextId: "personal" | "company"): string[] {
    return launchState.providerCredentialSessions
      .filter((session) => providerSessionAssignments[session.providerId] === contextId)
      .map((session) => session.providerId);
  }

  async function handleCreateContext(
    contextId: "personal" | "company",
    createContext: (importProviderIds: string[]) => Promise<ApiResult<CreateContextResult>>,
  ) {
    setOnboardingPendingContextId(contextId);
    setOnboardingError(undefined);

    try {
      const result = await createContext(importProviderIdsForContext(contextId));
      if (!result.ok) {
        setOnboardingError(result.error);
      }
    } finally {
      setOnboardingPendingContextId(undefined);
    }
  }

  return (
    <div className="space-y-8" onKeyDown={handleSelectorKeyDown}>
      {shouldRenderFirstRunWelcome(launchState) ? (
        <>
          <ProjectIdentity project={launchState.project} />
          <FirstRunWelcome
            launchState={launchState}
            providerCredentialSessions={launchState.providerCredentialSessions}
            providerSessionAssignments={providerSessionAssignments}
            pendingContextId={onboardingPendingContextId}
            error={onboardingError}
            onClassifyProviderSession={handleClassifyProviderSession}
            onCreatePersonal={
              onCreatePersonalContext
                ? () => void handleCreateContext("personal", onCreatePersonalContext)
                : undefined
            }
            onCreateCompany={
              onCreateCompanyContext
                ? () => void handleCreateContext("company", onCreateCompanyContext)
                : undefined
            }
          />
        </>
      ) : (
        <SelectorLayout
          projectIdentity={<ProjectIdentity project={launchState.project} />}
          contextCards={
            <>
              {selectedContextId === undefined ? (
                <p className="text-sm text-muted-foreground">No context selected</p>
              ) : null}

              <div className="grid gap-4 sm:grid-cols-2" role="group" aria-label="Available contexts">
                {launchState.contexts.map((context) => (
                  <ContextCard
                    key={context.id}
                    context={context}
                    selected={selectedContextId === context.id}
                    disabled={launchPending}
                    tabIndex={rovingContextId === context.id ? 0 : -1}
                    buttonRef={setContextButtonRef(context.id)}
                    onSelect={handleSelectContext}
                    onNavigate={handleContextNavigation}
                    onLaunchSelected={keyboardLaunchAvailable ? () => void handleLaunch() : undefined}
                  />
                ))}
              </div>

              <MissingDefaultContextActions
                launchState={launchState}
                providerCredentialSessions={launchState.providerCredentialSessions}
                providerSessionAssignments={providerSessionAssignments}
                pendingContextId={onboardingPendingContextId}
                error={onboardingError}
                onClassifyProviderSession={handleClassifyProviderSession}
                onCreatePersonal={
                  onCreatePersonalContext
                    ? () => void handleCreateContext("personal", onCreatePersonalContext)
                    : undefined
                }
                onCreateCompany={
                  onCreateCompanyContext
                    ? () => void handleCreateContext("company", onCreateCompanyContext)
                    : undefined
                }
              />
            </>
          }
          confidenceSummary={<SelectorConfidenceSummary context={selectedContext} />}
          rememberControl={
            <RememberProjectControl
              binding={launchState.binding}
              contexts={launchState.contexts}
              rememberProject={rememberProject}
              selectedContextId={selectedContextId}
              disabled={launchPending}
              onRememberProjectChange={setRememberProject}
            />
          }
          launchActions={
            <>
              {launchPending ? (
                <Card
                  as="p"
                  size="sm"
                  className="mb-3 border border-border bg-muted/30 p-3 text-sm text-muted-foreground"
                  role="status"
                >
                  Launching selected context...
                </Card>
              ) : null}

              {launchError ? <GuiErrorNotice error={launchError} /> : null}

              {mismatchError?.contextMismatch ? (
                <ContextMismatchDialog
                  mismatch={mismatchError.contextMismatch}
                  contexts={launchState.contexts}
                  launchPending={launchPending}
                  onCancel={() => setMismatchError(undefined)}
                  onOpenAnyway={() => void handleLaunch(true)}
                />
              ) : null}

              <SelectorActions
                launchDisabled={selectedContextId === undefined}
                launchPending={launchPending}
                onLaunch={() => void handleLaunch()}
                onCancel={() => void cancelSelector({ closeSelector: onCancel })}
              />
            </>
          }
        />
      )}
    </div>
  );
}

function MissingDefaultContextActions({
  launchState,
  providerCredentialSessions,
  providerSessionAssignments,
  pendingContextId,
  error,
  onClassifyProviderSession,
  onCreatePersonal,
  onCreateCompany,
}: {
  launchState: LaunchState;
  providerCredentialSessions: ProviderCredentialSession[];
  providerSessionAssignments: ProviderSessionAssignments;
  pendingContextId?: string;
  error?: DisplayError;
  onClassifyProviderSession: (providerId: string, contextId: "personal" | "company") => void;
  onCreatePersonal?: () => void;
  onCreateCompany?: () => void;
}) {
  const missingDefaults = missingDefaultContextIds(launchState.contexts);
  const missingPersonal = missingDefaults.includes("personal");
  const missingCompany = missingDefaults.includes("company");
  const pending = pendingContextId !== undefined;
  const classificationComplete = providerCredentialSessions.every(
    (session) => providerSessionAssignments[session.providerId] !== undefined,
  );

  if (!missingPersonal && !missingCompany) {
    return null;
  }

  return (
    <Card as="section" size="sm" className="border border-border bg-muted/30 py-0" aria-label="Add default contexts">
      <CardContent className="p-4">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <h3 className="text-sm font-semibold">Add another default context</h3>
            <p className="mt-1 text-sm text-muted-foreground">
              Create the missing Personal or Company context when this machine needs both identities.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            {missingPersonal ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={pending || !classificationComplete || onCreatePersonal === undefined}
                onClick={onCreatePersonal}
              >
                {pendingContextId === "personal" ? "Creating..." : "Add Personal"}
              </Button>
            ) : null}
            {missingCompany ? (
              <Button
                type="button"
                variant="outline"
                size="sm"
                disabled={pending || !classificationComplete || onCreateCompany === undefined}
                onClick={onCreateCompany}
              >
                {pendingContextId === "company" ? "Creating..." : "Add Company"}
              </Button>
            ) : null}
          </div>
        </div>
        <ProviderCredentialClassification
          sessions={providerCredentialSessions}
          assignments={providerSessionAssignments}
          disabled={pending}
          onClassify={onClassifyProviderSession}
        />
        {pendingContextId ? (
          <p className="mt-3 text-sm text-muted-foreground" role="status">
            Creating {pendingContextId === "personal" ? "Personal" : "Company"} context...
          </p>
        ) : null}
        {error ? <GuiErrorNotice error={error} /> : null}
      </CardContent>
    </Card>
  );
}

export { SelectorView };
