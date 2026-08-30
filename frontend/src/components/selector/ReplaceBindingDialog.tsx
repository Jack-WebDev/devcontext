import type { DisplayError } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ReplaceBindingDialogProps {
	boundContextName: string;
	replacementContextName: string;
	pending: boolean;
	error?: DisplayError;
	onKeepCurrent: () => void;
	onReplace: () => void;
}

function ReplaceBindingDialog({
	boundContextName,
	replacementContextName,
	pending,
	error,
	onKeepCurrent,
	onReplace,
}: ReplaceBindingDialogProps) {
	return (
		<Card
			as="section"
			aria-labelledby="replace-binding-title"
			aria-modal="true"
			className="border border-border py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3 id="replace-binding-title" className="text-base font-semibold">
						Remember this context?
					</h3>
					<p className="mt-1 text-muted-foreground">
						This project is still remembered for {boundContextName}. Remembering
						{` ${replacementContextName}`} will use it by default next time.
					</p>
				</div>

				{error ? (
					<p className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
						{error.message} {error.recovery}
					</p>
				) : null}

				<div className="flex justify-end gap-3">
					<Button type="button" disabled={pending} onClick={onKeepCurrent}>
						Keep {boundContextName}
					</Button>
					<Button
						type="button"
						variant="outline"
						disabled={pending}
						onClick={onReplace}
					>
						{pending
							? "Remembering..."
							: `Remember ${replacementContextName}`}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export { ReplaceBindingDialog };
