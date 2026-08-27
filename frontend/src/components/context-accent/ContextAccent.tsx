type ContextAccent = "sage" | "slate-blue" | "amber" | "custom" | "neutral";

interface ContextAccentIndicatorProps {
  accent: ContextAccent;
  className?: string;
}

function ContextAccentIndicator({ accent, className = "" }: ContextAccentIndicatorProps) {
  return <span className={`size-2 shrink-0 ${contextAccentClassName(accent)} ${className}`} aria-hidden="true" />;
}

function contextAccentFromMetadata(value: string | undefined): ContextAccent {
  switch (value) {
    case "sage":
    case "slate-blue":
    case "amber":
    case "custom":
      return value;
    default:
      return "neutral";
  }
}

function contextAccentClassName(accent: ContextAccent): string {
  switch (accent) {
    case "sage":
      return "bg-accent-personal";
    case "slate-blue":
      return "bg-accent-company";
    case "amber":
      return "bg-accent-freelance";
    case "custom":
      return "bg-accent-custom";
    case "neutral":
      return "bg-muted-foreground";
  }
}

export { ContextAccentIndicator, contextAccentClassName, contextAccentFromMetadata };
export type { ContextAccent, ContextAccentIndicatorProps };
