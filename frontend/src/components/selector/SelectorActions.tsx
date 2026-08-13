interface SelectorActionsProps {
  launchDisabled: boolean;
  launchPending: boolean;
  onLaunch: () => void;
  onCancel: () => void;
}

function SelectorActions({ launchDisabled, launchPending, onLaunch, onCancel }: SelectorActionsProps) {
  return (
    <div className="flex justify-end gap-3 border-t border-border pt-4">
      <button
        type="button"
        className="border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-foreground/30"
        disabled={launchPending}
        onClick={onCancel}
      >
        Cancel
      </button>
      <button
        type="button"
        className="bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
        disabled={launchDisabled || launchPending}
        onClick={onLaunch}
      >
        {launchPending ? "Launching..." : "Launch"}
      </button>
    </div>
  );
}

export { SelectorActions };
