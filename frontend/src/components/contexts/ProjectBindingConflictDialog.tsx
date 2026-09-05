import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ProjectBindingConflictDialogProps {
	projectName: string;
	boundContextName: string;
	onKeepExisting: () => void;
	onMoveToNewContext: () => void;
	onCancel: () => void;
}

function ProjectBindingConflictDialog({
	projectName,
	boundContextName,
	onKeepExisting,
	onMoveToNewContext,
	onCancel,
}: ProjectBindingConflictDialogProps) {
	return (
		<Card
			as="section"
			aria-labelledby="project-binding-conflict-title"
			aria-modal="true"
			className="border border-border py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3
						id="project-binding-conflict-title"
						className="text-base font-semibold"
					>
						This project already belongs to {boundContextName}
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						{projectName} is currently remembered for {boundContextName}. Moving
						it will be applied only when you create this new context.
					</p>
				</div>

				<div className="flex flex-wrap justify-end gap-3">
					<Button type="button" variant="outline" onClick={onCancel}>
						Cancel
					</Button>
					<Button type="button" variant="outline" onClick={onKeepExisting}>
						Keep existing
					</Button>
					<Button type="button" onClick={onMoveToNewContext}>
						Move to new context
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export type { ProjectBindingConflictDialogProps };
export { ProjectBindingConflictDialog };
