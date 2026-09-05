import { useEffect, useState } from "react";

import type {
	ApiResult,
	DeleteContextPreview,
	DeleteContextResult,
	DisplayError,
} from "../../lib/devctx-api";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "../ui/alert-dialog.js";

interface ContextDeleteDialogProps {
	contextId: string;
	contextName: string;
	onClose: () => void;
	onDeleted: () => void;
	preview: (contextId: string) => Promise<ApiResult<DeleteContextPreview>>;
	deleteContext: (contextId: string) => Promise<ApiResult<DeleteContextResult>>;
}

function ContextDeleteDialog({
	contextId,
	contextName,
	onClose,
	onDeleted,
	preview,
	deleteContext,
}: ContextDeleteDialogProps) {
	const [result, setResult] = useState<ApiResult<DeleteContextPreview>>();
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<DisplayError>();

	useEffect(() => {
		void preview(contextId).then(setResult);
	}, [contextId, preview]);

	async function confirmDelete() {
		setPending(true);
		setError(undefined);
		const deleted = await deleteContext(contextId);
		setPending(false);
		if (!deleted.ok) {
			setError(deleted.error);
			return;
		}
		onDeleted();
	}

	const previewError = result && !result.ok ? result.error : undefined;
	const contextCanBeDeleted = result?.ok && result.data.context.archivedAt;

	return (
		<AlertDialog open onOpenChange={(open) => !open && onClose()}>
			<AlertDialogContent>
				<AlertDialogHeader>
					<AlertDialogTitle>Delete {contextName}?</AlertDialogTitle>
					<AlertDialogDescription>
						This removes the context and its Dev Context-owned isolated state.
						Project folders are never deleted.
					</AlertDialogDescription>
				</AlertDialogHeader>
				{result?.ok ? (
					<p className="text-body text-secondary">
						{deleteImpact(result.data)}
					</p>
				) : null}
				{!result ? (
					<p className="text-body text-secondary">Checking impact...</p>
				) : null}
				{previewError || error ? (
					<p className="text-sm text-destructive">
						{(previewError ?? error)?.message}
					</p>
				) : null}
				{result?.ok && !contextCanBeDeleted ? (
					<p className="text-sm text-warning">
						Archive this context before permanently deleting it.
					</p>
				) : null}
				<AlertDialogFooter>
					<AlertDialogCancel disabled={pending}>Cancel</AlertDialogCancel>
					<AlertDialogAction
						variant="destructive"
						disabled={!contextCanBeDeleted || pending}
						onClick={() => void confirmDelete()}
					>
						{pending ? "Deleting..." : "Delete context"}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}

function deleteImpact(preview: DeleteContextPreview): string {
	const count = preview.projectBindings.length;
	return count === 0
		? "No project bindings will be removed."
		: `${count} ${count === 1 ? "project binding" : "project bindings"} will be removed.`;
}

export { ContextDeleteDialog, deleteImpact };
