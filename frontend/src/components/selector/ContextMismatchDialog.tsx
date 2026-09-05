import type { KeyboardEvent } from "react";

import type { ContextMismatch, ContextState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ContextMismatchDialogProps {
	mismatch: ContextMismatch;
	contexts: ContextState[];
	launchPending: boolean;
	onCancel: () => void;
	onUseRememberedContext: () => void;
	onOpenAnyway: () => void;
}

function ContextMismatchDialog({
	mismatch,
	contexts,
	launchPending,
	onCancel,
	onUseRememberedContext,
	onOpenAnyway,
}: ContextMismatchDialogProps) {
	const rememberedContextName = contextName(contexts, mismatch.boundContextId);
	const requestedContextName = contextName(
		contexts,
		mismatch.requestedContextId,
	);

	function handleKeyDown(event: KeyboardEvent<HTMLElement>) {
		if (mismatchDialogKeyboardAction(event.key, launchPending) !== "cancel") {
			return;
		}

		event.preventDefault();
		event.stopPropagation();
		onCancel();
	}

	return (
		<Card
			as="section"
			aria-labelledby="context-mismatch-title"
			aria-modal="true"
			className="border border-destructive/30 py-0"
			onKeyDown={handleKeyDown}
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3
						id="context-mismatch-title"
						className="text-base font-semibold text-foreground"
					>
						This project normally opens in {rememberedContextName}
					</h3>
					<p className="mt-1 text-muted-foreground">
						Dev Context remembers {rememberedContextName} for this project.
						Opening it with {requestedContextName} instead can use that
						context's signed-in accounts, tools, and environment. It will not
						change which context this project normally opens in.
					</p>
				</div>

				<div className="grid gap-2 text-sm">
					<MismatchDetail label="Project" value={mismatch.projectPath} />
					<div
						aria-label="Context comparison"
						className="grid gap-px overflow-hidden rounded-lg border border-border bg-border sm:grid-cols-2"
						role="group"
					>
						<ContextComparison
							label="Remembered context"
							name={rememberedContextName}
							detail="Normal choice"
						/>
						<ContextComparison
							label="Requested context"
							name={requestedContextName}
							detail="Open only if intended"
						/>
					</div>
				</div>

				<div className="flex flex-wrap justify-end gap-3">
					<Button
						type="button"
						variant="outline"
						disabled={launchPending}
						onClick={onCancel}
					>
						Cancel
					</Button>
					<Button
						type="button"
						variant="destructive"
						disabled={launchPending}
						onClick={onOpenAnyway}
					>
						Open {requestedContextName} anyway
					</Button>
					<Button
						type="button"
						autoFocus
						disabled={launchPending}
						onClick={onUseRememberedContext}
					>
						{launchPending ? "Opening..." : "Use remembered context"}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function ContextComparison({
	label,
	name,
	detail,
}: {
	label: string;
	name: string;
	detail: string;
}) {
	return (
		<div className="bg-card p-3">
			<p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
				{label}
			</p>
			<p className="mt-1 font-medium text-foreground">{name}</p>
			<p className="mt-1 text-xs text-muted-foreground">{detail}</p>
		</div>
	);
}

function MismatchDetail({ label, value }: { label: string; value: string }) {
	return (
		<div className="grid gap-1">
			<p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">
				{label}
			</p>
			<p className="truncate font-mono text-sm text-foreground" title={value}>
				{value}
			</p>
		</div>
	);
}

function contextName(contexts: ContextState[], contextId: string): string {
	const context = contexts.find((candidate) => candidate.id === contextId);
	return context?.name ?? contextId;
}

type MismatchDialogKeyboardAction = "cancel" | "none";

function mismatchDialogKeyboardAction(
	key: string,
	launchPending: boolean,
): MismatchDialogKeyboardAction {
	return key === "Escape" && !launchPending ? "cancel" : "none";
}

export { ContextMismatchDialog, contextName, mismatchDialogKeyboardAction };
