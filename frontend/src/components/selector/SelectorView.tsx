import { useEffect, useState } from "react";

import type { LaunchState } from "../../lib/devctx-api";
import { ContextCard } from "./ContextCard";
import { ProjectIdentity } from "./ProjectIdentity";
import { initialSelectedContextId, nextSelectedContextId } from "./selection-state";

interface SelectorViewProps {
  launchState: LaunchState;
}

function SelectorView({ launchState }: SelectorViewProps) {
  const [selectedContextId, setSelectedContextId] = useState<string | undefined>(() =>
    initialSelectedContextId(launchState),
  );

  useEffect(() => {
    setSelectedContextId(initialSelectedContextId(launchState));
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
      </div>
    </div>
  );
}

export { SelectorView };
