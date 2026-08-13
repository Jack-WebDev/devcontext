import { useEffect, useState } from "react";

import type { LaunchState } from "../../lib/devctx-api";
import { ContextCard } from "./ContextCard";
import { ProjectIdentity } from "./ProjectIdentity";
import { RememberProjectControl } from "./RememberProjectControl";
import { SelectorActions } from "./SelectorActions";
import { cancelSelector } from "./cancel-action";
import { initialSelectedContextId, nextSelectedContextId } from "./selection-state";

interface SelectorViewProps {
  launchState: LaunchState;
  onCancel: () => Promise<void> | void;
}

function SelectorView({ launchState, onCancel }: SelectorViewProps) {
  const [selectedContextId, setSelectedContextId] = useState<string | undefined>(() =>
    initialSelectedContextId(launchState),
  );
  const [rememberProject, setRememberProject] = useState(false);

  useEffect(() => {
    setSelectedContextId(initialSelectedContextId(launchState));
    setRememberProject(false);
  }, [launchState]);

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
              onSelect={(contextId) => setSelectedContextId(nextSelectedContextId(launchState.contexts, contextId))}
            />
          ))}
        </div>

        <RememberProjectControl
          binding={launchState.binding}
          contexts={launchState.contexts}
          rememberProject={rememberProject}
          selectedContextId={selectedContextId}
          onRememberProjectChange={setRememberProject}
        />

        <SelectorActions onCancel={() => void cancelSelector({ closeSelector: onCancel })} />
      </div>
    </div>
  );
}

export { SelectorView };
