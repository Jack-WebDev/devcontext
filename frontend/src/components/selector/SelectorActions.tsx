import { Button } from "../ui/button.js";
import { Separator } from "../ui/separator.js";

interface SelectorActionsProps {
  launchDisabled: boolean;
  launchPending: boolean;
  onLaunch: () => void;
  onCancel: () => void;
}

function SelectorActions({ launchDisabled, launchPending, onLaunch, onCancel }: SelectorActionsProps) {
  return (
    <div className="space-y-4">
      <Separator />
      <div className="flex justify-end gap-3">
        <Button type="button" variant="outline" disabled={launchPending} onClick={onCancel}>
          Cancel
        </Button>
        <Button type="button" disabled={launchDisabled || launchPending} onClick={onLaunch}>
          {launchPending ? "Launching..." : "Launch"}
        </Button>
      </div>
    </div>
  );
}

export { SelectorActions };
