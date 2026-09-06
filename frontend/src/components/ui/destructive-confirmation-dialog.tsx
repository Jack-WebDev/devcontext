import type { ReactNode } from "react";
import type { DisplayError } from "../../lib/devctx-api.js";

import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "./alert-dialog.js";

interface DestructiveConfirmationDialogProps {
	title: string;
	objectName: string;
	impact: ReactNode;
	nonDeletionAssurance: ReactNode;
	confirmLabel: string;
	pending?: boolean;
	error?: DisplayError;
	onCancel: () => void;
	onConfirm: () => void;
}

function DestructiveConfirmationDialog({ title, objectName, impact, nonDeletionAssurance, confirmLabel, pending = false, error, onCancel, onConfirm }: DestructiveConfirmationDialogProps) {
	return <AlertDialog open onOpenChange={(open) => !open && !pending && onCancel()}>
		<AlertDialogContent>
			<AlertDialogHeader>
				<AlertDialogTitle>{title}</AlertDialogTitle>
				<AlertDialogDescription>Affects <strong>{objectName}</strong>.</AlertDialogDescription>
			</AlertDialogHeader>
			<div className="space-y-2 text-sm text-muted-foreground">
				<p>{impact}</p>
				<p>{nonDeletionAssurance}</p>
			</div>
			{error ? <div role="alert" className="text-sm text-destructive"><p>{error.message}</p><p className="mt-1">{error.recovery}</p></div> : null}
			<AlertDialogFooter>
				<AlertDialogCancel disabled={pending}>Cancel</AlertDialogCancel>
				<AlertDialogAction variant="destructive" disabled={pending} onClick={onConfirm}>{confirmLabel}</AlertDialogAction>
			</AlertDialogFooter>
		</AlertDialogContent>
	</AlertDialog>;
}

export type { DestructiveConfirmationDialogProps };
export { DestructiveConfirmationDialog };
