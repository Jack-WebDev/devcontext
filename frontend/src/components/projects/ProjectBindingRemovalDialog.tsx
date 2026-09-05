import type { DisplayError, ProjectListItem } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ProjectBindingRemovalDialogProps {
	project: ProjectListItem;
	pending: boolean;
	error?: DisplayError;
	onCancel: () => void;
	onConfirm: () => void;
}

function ProjectBindingRemovalDialog({
	project,
	pending,
	error,
	onCancel,
	onConfirm,
}: ProjectBindingRemovalDialogProps) {
	const contextName = project.contextName ?? project.contextId ?? "this context";
	return (
		<Card
			as="section"
			aria-labelledby="project-binding-removal-title"
			aria-modal="true"
			className="border border-border py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3 id="project-binding-removal-title" className="text-base font-semibold">
						Remove project binding?
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						{project.project.name} will no longer normally open with {contextName}.
					</p>
				</div>
				<div className="space-y-2 border-y border-border py-4 text-sm text-muted-foreground">
					<p>Project files and folders are never deleted.</p>
					<p>The next launch will ask you to choose a context.</p>
				</div>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error.message}
					</p>
				) : null}
				<div className="flex justify-end gap-3">
					<Button type="button" variant="outline" disabled={pending} onClick={onCancel}>
						Cancel
					</Button>
					<Button type="button" variant="destructive" disabled={pending} onClick={onConfirm}>
						{pending ? "Removing binding..." : "Remove binding"}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export { ProjectBindingRemovalDialog };
