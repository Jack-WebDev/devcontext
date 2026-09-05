import type { ContextState } from "../../lib/devctx-api.js";
import { Button } from "../ui/button.js";

type CreationStepStatus =
	| "pending"
	| "running"
	| "complete"
	| "skipped"
	| "failed";
interface CreationStep {
	id: "create" | "bind" | "initialize" | "verify";
	label: string;
	status: CreationStepStatus;
	detail?: string;
}

function ContextCreationProgress({
	steps,
	error,
	onRetry,
	onBack,
}: {
	steps: CreationStep[];
	error?: string;
	onRetry?: () => void;
	onBack?: () => void;
}) {
	return (
		<section
			aria-labelledby="context-creation-progress-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Create a context
				</p>
				<h2
					id="context-creation-progress-title"
					className="text-2xl font-semibold"
				>
					Setting up your context
				</h2>
				<p className="text-sm text-muted-foreground">
					Dev Context is completing the selected local setup.
				</p>
			</div>
			<ol className="space-y-3">
				{steps.map((step) => (
					<li key={step.id} className="rounded-lg border bg-card p-4">
						<p className="font-medium">
							{step.label}{" "}
							<span className="text-sm font-normal text-muted-foreground">
								{creationStepLabel(step.status)}
							</span>
						</p>
						{step.detail ? (
							<p className="mt-1 text-sm text-muted-foreground">
								{step.detail}
							</p>
						) : null}
					</li>
				))}
			</ol>
			{error ? (
				<>
					<p role="alert" className="text-sm text-destructive">
						{error}
					</p>
					<div className="flex justify-between gap-3">
						{onBack ? (
							<Button type="button" variant="outline" onClick={onBack}>
								Back to review
							</Button>
						) : (
							<span />
						)}
						{onRetry ? (
							<Button type="button" onClick={onRetry}>
								Retry
							</Button>
						) : null}
					</div>
				</>
			) : null}
		</section>
	);
}

function ContextCreateSuccessScreen({
	context,
	projectName,
	onOpenProject,
	onViewContext,
	onCreateAnother,
}: {
	context: ContextState;
	projectName?: string;
	onOpenProject?: () => void;
	onViewContext?: () => void;
	onCreateAnother: () => void;
}) {
	return (
		<section
			aria-labelledby="context-created-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Context ready
				</p>
				<h2 id="context-created-title" className="text-2xl font-semibold">
					{context.name} is ready
				</h2>
				<p className="text-sm text-muted-foreground">
					Its local tool settings and account storage are isolated from your
					other contexts.
				</p>
			</div>
			<div className="rounded-xl border bg-card p-4 text-sm">
				<p className="font-medium">Launch tool: {context.tool.name}</p>
				{projectName ? (
					<p className="mt-1 text-muted-foreground">
						{projectName} is ready to open with this context.
					</p>
				) : null}
			</div>
			<div className="flex flex-wrap gap-3">
				{onOpenProject ? (
					<Button type="button" onClick={onOpenProject}>
						{projectName ? `Open ${projectName}` : "Open a Project"}
					</Button>
				) : null}
				{onViewContext ? (
					<Button type="button" variant="outline" onClick={onViewContext}>
						View Context
					</Button>
				) : null}
				<Button type="button" variant="outline" onClick={onCreateAnother}>
					Create Another Context
				</Button>
			</div>
		</section>
	);
}

function creationStepLabel(status: CreationStepStatus) {
	return status === "running"
		? "In progress"
		: status === "complete"
			? "Complete"
			: status === "skipped"
				? "Not needed"
				: status === "failed"
					? "Needs attention"
					: "Waiting";
}

export type { CreationStep, CreationStepStatus };
export {
	ContextCreationProgress,
	ContextCreateSuccessScreen,
	creationStepLabel,
};
