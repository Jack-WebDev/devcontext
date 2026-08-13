import { type KeyboardEvent, useEffect, useRef, useState } from "react";

import type {
  ApiResult,
  BindProjectRequest,
  CreateContextResult,
  DisplayError,
  LaunchProjectRequest,
  LaunchProjectResult,
  LaunchState,
  ProjectBindingState,
} from "../../lib/devctx-api";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
import { ContextCard } from "./ContextCard";
import { FirstRunWelcome, shouldRenderFirstRunWelcome } from "./FirstRunWelcome";
import { GuiErrorNotice } from "./GuiErrorNotice";
import { ProjectIdentity } from "./ProjectIdentity";
import { RememberProjectControl } from "./RememberProjectControl";
import { SelectorActions } from "./SelectorActions";
import { cancelSelector } from "./cancel-action";
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
  onLaunchProject: (request: LaunchProjectRequest) => Promise<ApiResult<LaunchProjectResult>>;
  onCancel: () => Promise<void> | void;
  onCreatePersonalContext?: () => Promise<ApiResult<CreateContextResult>>;
  onCreateCompanyContext?: () => Promise<ApiResult<CreateContextResult>>;
}

function SelectorView({
  launchState,
  onBindProject,
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
  const contextButtonRefs = useRef(new Map<string, HTMLButtonElement>());
  const launchGuard = useRef(createLaunchRequestGuard());
  const mismatchDialogOpen = mismatchError?.contextMismatch !== undefined;
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

  async function handleCreateContext(contextId: string, createContext: () => Promise<ApiResult<CreateContextResult>>) {
    setOnboardingPendingContextId(contextId);
    setOnboardingError(undefined);

    try {
      const result = await createContext();
      if (!result.ok) {
        setOnboardingError(result.error);
      }
    } finally {
      setOnboardingPendingContextId(undefined);
    }
  }

  return (
    <div className="space-y-8" onKeyDown={handleSelectorKeyDown}>
      <ProjectIdentity project={launchState.project} />

      {shouldRenderFirstRunWelcome(launchState) ? (
        <FirstRunWelcome
          launchState={launchState}
          pendingContextId={onboardingPendingContextId}
          error={onboardingError}
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
      ) : (
        <div className="space-y-3">
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

          {launchPending ? (
            <p className="border border-border bg-muted/30 p-3 text-sm text-muted-foreground" role="status">
              Launching selected context...
            </p>
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

          <RememberProjectControl
            binding={launchState.binding}
            contexts={launchState.contexts}
            rememberProject={rememberProject}
            selectedContextId={selectedContextId}
            disabled={launchPending}
            onRememberProjectChange={setRememberProject}
          />

          <SelectorActions
            launchDisabled={selectedContextId === undefined}
            launchPending={launchPending}
            onLaunch={() => void handleLaunch()}
            onCancel={() => void cancelSelector({ closeSelector: onCancel })}
          />
        </div>
      )}
    </div>
  );
}

export { SelectorView };
