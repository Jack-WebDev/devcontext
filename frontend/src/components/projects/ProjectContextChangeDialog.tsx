import { useEffect, useState } from "react";

import type {
	ContextState,
	DisplayError,
	ProjectListItem,
} from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ProjectContextChangeDialogProps {
	project: ProjectListItem;
	contexts: ContextState[];
	pending: boolean;
	error?: DisplayError;
	onCancel: () => void;
	onConfirm: (contextId: string) => void;
}

function ProjectContextChangeDialog({
	project,
	contexts,
	pending,
	error,
	onCancel,
	onConfirm,
}: ProjectContextChangeDialogProps) {
	const initialContextID = project.contextId ?? contexts[0]?.id ?? "";
	const [contextID, setContextID] = useState(initialContextID);
	const selectedContext = contexts.find((context) => context.id === contextID);

	useEffect(() => {
		setContextID(project.contextId ?? contexts[0]?.id ?? "");
	}, [project.contextId, contexts]);

	return (
		<Card
			as="section"
			aria-labelledby="project-context-change-title"
			aria-modal="true"
			className="border border-border py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3
						id="project-context-change-title"
						className="text-base font-semibold"
					>
						Change project context
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						This changes the development identity Dev Context remembers for{" "}
						{project.project.name}.
					</p>
				</div>

				<dl className="grid gap-2 border-y border-border py-4 text-sm">
					<ProjectContextDetail label="Project" value={project.project.path} />
					<ProjectContextDetail
						label="Current context"
						value={project.contextName ?? project.contextId ?? "None"}
					/>
				</dl>

				<label
					className="grid gap-2 text-sm font-medium"
					htmlFor="project-context-select"
				>
					New context
					<select
						id="project-context-select"
						className="h-10 border border-input bg-background px-3 text-sm text-foreground"
						disabled={pending || contexts.length === 0}
						value={contextID}
						onChange={(event) => setContextID(event.currentTarget.value)}
					>
						{contexts.map((context) => (
							<option key={context.id} value={context.id}>
								{context.name}
							</option>
						))}
					</select>
				</label>

				{selectedContext ? (
					<ContextSafetySummary context={selectedContext} />
				) : null}
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error.message}
					</p>
				) : null}

				<div className="flex justify-end gap-3">
					<Button
						type="button"
						variant="outline"
						disabled={pending}
						onClick={onCancel}
					>
						Cancel
					</Button>
					<Button
						type="button"
						disabled={pending || selectedContext === undefined}
						onClick={() => onConfirm(contextID)}
					>
						{pending
							? "Saving context..."
							: `Use ${selectedContext?.name ?? "context"}`}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function ContextSafetySummary({ context }: { context: ContextState }) {
	const status = context.confidence?.status ?? "blocked";
	const implication = safetyImplication(status, context.name);
	return (
		<section
			className="space-y-2 border border-border bg-surface-muted p-3"
			aria-labelledby="project-context-safety-heading"
		>
			<div className="flex items-center justify-between gap-3">
				<h4 id="project-context-safety-heading" className="font-medium">
					Safety implications
				</h4>
				<StatusIndicator status={status} />
			</div>
			<p className="text-sm text-muted-foreground">{implication}</p>
		</section>
	);
}

function safetyImplication(
	status: "ready" | "needs_attention" | "blocked",
	contextName: string,
): string {
	switch (status) {
		case "ready":
			return `${contextName} is ready for an isolated launch.`;
		case "needs_attention":
			return `${contextName} can launch, but its setup needs attention before you continue.`;
		case "blocked":
			return `${contextName} is blocked and cannot launch until its required setup is resolved.`;
	}
}

function ProjectContextDetail({
	label,
	value,
}: {
	label: string;
	value: string;
}) {
	return (
		<div className="grid gap-1">
			<dt className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
				{label}
			</dt>
			<dd className="truncate font-mono text-sm" title={value}>
				{value}
			</dd>
		</div>
	);
}

export { ProjectContextChangeDialog, safetyImplication };
