import type { ContextState, ProjectBindingState } from "../../lib/devctx-api";
import { Card } from "../ui/card.js";

interface RememberProjectControlProps {
  binding: ProjectBindingState;
  contexts: ContextState[];
  rememberProject: boolean;
  selectedContextId?: string;
  disabled?: boolean;
  onRememberProjectChange?: (rememberProject: boolean) => void;
}

function RememberProjectControl({
  binding,
  contexts,
  rememberProject,
  selectedContextId,
  disabled: disabledByParent = false,
  onRememberProjectChange,
}: RememberProjectControlProps) {
  const boundContext = boundContextName(binding, contexts);
  if (boundContext !== undefined) {
    return (
      <Card as="p" size="sm" className="border border-border bg-muted/30 p-3 text-sm text-muted-foreground">
        This project is remembered for <span className="font-medium text-foreground">{boundContext}</span>.
      </Card>
    );
  }

  const disabled = disabledByParent || selectedContextId === undefined;

  return (
    <Card as="label" size="sm" className="flex-row items-start gap-3 border border-border p-3 text-sm">
      <input
        type="checkbox"
        className="mt-0.5 size-4 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:cursor-not-allowed"
        checked={rememberProject}
        disabled={disabled}
        readOnly={onRememberProjectChange === undefined}
        onChange={(event) => onRememberProjectChange?.(event.currentTarget.checked)}
      />
      <span className="space-y-1">
        <span className="block font-medium">Remember this project</span>
        <span className="block text-muted-foreground">
          {disabled
            ? disabledByParent
              ? "Remembering is unavailable while launch is in progress."
              : "Select a context before remembering this project."
            : "Use this context automatically for this project next time."}
        </span>
      </span>
    </Card>
  );
}

function boundContextName(binding: ProjectBindingState, contexts: ContextState[]): string | undefined {
  if (!binding.bound || binding.dangling || binding.contextId === undefined) {
    return undefined;
  }

  const context = contexts.find((candidate) => candidate.id === binding.contextId);
  return context?.name ?? binding.contextId;
}

export { RememberProjectControl, boundContextName };
