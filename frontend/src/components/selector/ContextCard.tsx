import type { ContextState } from "../../lib/devctx-api";

interface ContextCardProps {
  context: ContextState;
  selected?: boolean;
  onSelect?: (contextId: string) => void;
}

function ContextCard({ context, selected = false, onSelect }: ContextCardProps) {
  const unselectedClassName = onSelect ? "border-border hover:border-foreground/20" : "border-border";
  const selectedClassName = selected
    ? "border-primary ring-2 ring-ring/40"
    : unselectedClassName;
  const className = `min-w-0 border bg-card p-5 text-left text-card-foreground shadow-sm transition-colors ${selectedClassName}`;
  const content = (
    <div className="min-w-0 space-y-2">
      <h3 id={`context-${context.id}-name`} className="truncate text-base font-semibold" title={context.name}>
        {context.name}
      </h3>
      <p className="truncate font-mono text-xs text-muted-foreground" title={context.id}>
        {context.id}
      </p>
    </div>
  );

  if (onSelect) {
    return (
      <button
        type="button"
        className={className}
        aria-labelledby={`context-${context.id}-name`}
        aria-pressed={selected}
        onClick={() => onSelect(context.id)}
      >
        {content}
      </button>
    );
  }

  return (
    <article
      className={className}
      aria-labelledby={`context-${context.id}-name`}
      data-selected={selected ? "true" : undefined}
    >
      {content}
    </article>
  );
}

export { ContextCard };
