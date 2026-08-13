interface SelectorActionsProps {
  onCancel: () => void;
}

function SelectorActions({ onCancel }: SelectorActionsProps) {
  return (
    <div className="flex justify-end border-t border-border pt-4">
      <button
        type="button"
        className="border border-border px-4 py-2 text-sm font-medium text-foreground transition-colors hover:border-foreground/30"
        onClick={onCancel}
      >
        Cancel
      </button>
    </div>
  );
}

export { SelectorActions };
