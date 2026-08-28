import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface AccountIdentityMismatchDialogProps {
  contextName: string;
  launchPending: boolean;
  onCancel: () => void;
  onReviewConfiguration: () => void;
  onLaunchAnyway: () => void;
}

function AccountIdentityMismatchDialog({
  contextName,
  launchPending,
  onCancel,
  onReviewConfiguration,
  onLaunchAnyway,
}: AccountIdentityMismatchDialogProps) {
  return (
    <Card
      as="section"
      aria-labelledby="account-identity-mismatch-title"
      aria-modal="true"
      className="border border-amber-500/40 py-0"
      role="dialog"
    >
      <CardContent className="space-y-4 p-5">
        <div>
          <h3 id="account-identity-mismatch-title" className="text-base font-semibold text-foreground">
            Review provider accounts
          </h3>
          <p className="mt-1 text-muted-foreground">
            Verified provider email identities do not match for {contextName}. This may be intentional, but review the accounts before launching this context.
          </p>
        </div>
        <div className="flex flex-wrap justify-end gap-3">
          <Button type="button" variant="outline" disabled={launchPending} onClick={onCancel}>
            Cancel
          </Button>
          <Button type="button" variant="outline" disabled={launchPending} onClick={onReviewConfiguration}>
            Review configuration
          </Button>
          <Button type="button" disabled={launchPending} onClick={onLaunchAnyway}>
            {launchPending ? "Launching..." : "Launch anyway"}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

export { AccountIdentityMismatchDialog };
