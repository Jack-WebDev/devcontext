import type { DisplayError, PreflightLaunchProjectResult } from "../../lib/devctx-api";
import { Card, CardContent } from "../ui/card.js";
import { PreflightReviewView } from "./PreflightReviewView.js";

interface PreflightReviewDialogProps {
	preflight: PreflightLaunchProjectResult;
	pending: boolean;
	error?: DisplayError;
	onCancel: () => void;
	onContinue: () => void;
}

// PreflightReviewDialog gives management launch actions the same deliberate
// warning-continuation step as the focused launcher.
function PreflightReviewDialog({
	preflight,
	pending,
	error,
	onCancel,
	onContinue,
}: PreflightReviewDialogProps) {
	return (
		<Card
			as="section"
			role="dialog"
			aria-modal="true"
			aria-label="Preflight review"
			className="border border-border py-0"
		>
			<CardContent className="p-5">
				<PreflightReviewView
					projectName={preflight.project.name}
					contextName={preflight.context.name}
					preflight={preflight}
					pending={pending}
					onFixFirst={onCancel}
					onLaunchWithoutIt={
						preflight.groups.some((group) => group.blocking) || pending
							? undefined
							: onContinue
					}
				/>
				{error ? (
					<p className="mt-3 text-sm text-destructive" role="alert">
						{error.message}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}

export type { PreflightReviewDialogProps };
export { PreflightReviewDialog };
