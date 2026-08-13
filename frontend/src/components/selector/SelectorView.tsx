import { useEffect, useRef, useState } from "react";

import type {
  ApiResult,
  BindProjectRequest,
  DisplayError,
  LaunchProjectRequest,
  LaunchProjectResult,
  LaunchState,
  ProjectBindingState,
} from "../../lib/devctx-api";
import { ContextMismatchDialog } from "./ContextMismatchDialog";
import { ContextCard } from "./ContextCard";
import { GuiErrorNotice } from "./GuiErrorNotice";
import { ProjectIdentity } from "./ProjectIdentity";
import { RememberProjectControl } from "./RememberProjectControl";
import { SelectorActions } from "./SelectorActions";
import { cancelSelector } from "./cancel-action";
import { createLaunchRequestGuard, launchSelectedContext } from "./launch-action";
import { initialSelectedContextId, nextSelectedContextId } from "./selection-state";

interface SelectorViewProps {
  launchState: LaunchState;
  onBindProject: (request: BindProjectRequest) => Promise<ApiResult<ProjectBindingState>>;
  onLaunchProject: (request: LaunchProjectRequest) => Promise<ApiResult<LaunchProjectResult>>;
  onCancel: () => Promise<void> | void;
}

function SelectorView({ launchState, onBindProject, onLaunchProject, onCancel }: SelectorViewProps) {
  const [selectedContextId, setSelectedContextId] = useState<string | undefined>(() =>
    initialSelectedContextId(launchState),
  );
  const [rememberProject, setRememberProject] = useState(false);
  const [launchPending, setLaunchPending] = useState(false);
  const [launchError, setLaunchError] = useState<DisplayError | undefined>(undefined);
  const [mismatchError, setMismatchError] = useState<DisplayError | undefined>(undefined);
  const launchGuard = useRef(createLaunchRequestGuard());

  useEffect(() => {
    setSelectedContextId(initialSelectedContextId(launchState));
    setRememberProject(false);
    setLaunchPending(false);
    setLaunchError(undefined);
    setMismatchError(undefined);
    launchGuard.current = createLaunchRequestGuard();
  }, [launchState]);

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

  return (
    <div className="space-y-8">
      <ProjectIdentity project={launchState.project} />

      <div className="space-y-3">
        {selectedContextId === undefined ? (
          <p className="text-sm text-muted-foreground">No context selected</p>
        ) : null}

        <div className="grid gap-4 sm:grid-cols-2">
          {launchState.contexts.map((context) => (
            <ContextCard
              key={context.id}
              context={context}
              selected={selectedContextId === context.id}
              disabled={launchPending}
              onSelect={(contextId) => setSelectedContextId(nextSelectedContextId(launchState.contexts, contextId))}
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
    </div>
  );
}

export { SelectorView };
