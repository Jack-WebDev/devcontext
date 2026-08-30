import type { DisplayError } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface DanglingBindingDialogProps {
	missingContextId?: string;
	pending: boolean;
	error?: DisplayError;
	onChooseContext: () => void;
	onRemoveBinding: () => void;
	onCancel: () => void;
}

function DanglingBindingDialog({
	missingContextId,
	pending,
	error,
	onChooseContext,
	onRemoveBinding,
	onCancel,
}: DanglingBindingDialogProps) {
	const missingContext = missingContextId ?? "the remembered context";

	return (
		<Card
			as="section"
			aria-labelledby="dangling-binding-title"
			aria-modal="true"
			className="border border-amber-500/30 py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3 id="dangling-binding-title" className="text-base font-semibold">
						Remembered context unavailable
					</h3>
					<p className="mt-1 text-muted-foreground">
						This project is remembered for {missingContext}, which is no longer
						available. Choose another context for this launch, or remove the
						outdated remembered context.
					</p>
				</div>

				{error ? (
					<p className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
						{error.message} {error.recovery}
					</p>
				) : null}

				<div className="flex flex-wrap justify-end gap-3">
					<Button type="button" variant="outline" disabled={pending} onClick={onCancel}>
						Cancel
					</Button>
					<Button type="button" disabled={pending} onClick={onChooseContext}>
						Choose a context
					</Button>
					<Button
						type="button"
						variant="destructive"
						disabled={pending}
						onClick={onRemoveBinding}
					>
						{pending ? "Removing..." : "Remove remembered context"}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export { DanglingBindingDialog };
