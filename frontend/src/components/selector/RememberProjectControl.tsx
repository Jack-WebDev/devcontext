import type { ContextState, ProjectBindingState } from "../../lib/devctx-api";
import { Card } from "../ui/card.js";
import { Checkbox } from "../ui/checkbox.js";

interface RememberProjectControlProps {
	binding: ProjectBindingState;
	contexts: ContextState[];
	rememberProject: boolean;
	selectedContextId?: string;
	disabled?: boolean;
	onRememberProjectChange?: (rememberProject: boolean) => void;
}

function RememberProjectControl({
	binding,
	contexts,
	rememberProject,
	selectedContextId,
	disabled: disabledByParent = false,
	onRememberProjectChange,
}: RememberProjectControlProps) {
	const boundContext = boundContextName(binding, contexts);
	if (boundContext !== undefined) {
		return (
			<Card as="div" size="sm" hierarchy="inset" className="p-3 text-sm">
				<p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
					Remembered context
				</p>
				<p className="mt-1 text-muted-foreground">
					<span className="font-medium text-foreground">{boundContext}</span>{" "}
					will be suggested the next time you open this project.
				</p>
			</Card>
		);
	}
	if (binding.dangling) {
		return (
			<Card as="div" size="sm" hierarchy="inset" className="p-3 text-sm">
				<p className="text-xs font-semibold tracking-wide text-muted-foreground uppercase">
					Remembered context unavailable
				</p>
				<p className="mt-1 text-muted-foreground">
					Choose a context to launch this project without changing its remembered
					context.
				</p>
			</Card>
		);
	}

	const disabled = disabledByParent || selectedContextId === undefined;
	const selectedContext = contexts.find(
		(context) => context.id === selectedContextId,
	);
	const label = selectedContext
		? `Remember ${selectedContext.name} for this project`
		: "Remember this project";

	return (
		<Card
			as="label"
			size="sm"
			hierarchy="selection"
			className="flex-row items-start gap-3 p-3 text-sm has-checked:border-primary has-checked:ring-2 has-checked:ring-ring/30"
		>
			<Checkbox
				className="mt-0.5"
				checked={rememberProject}
				disabled={disabled}
				onCheckedChange={(checked) =>
					onRememberProjectChange?.(checked === true)
				}
			/>
			<span className="space-y-1">
				<span className="block font-medium">{label}</span>
				<span className="block text-muted-foreground">
					{disabled
						? disabledByParent
							? "Remembering is unavailable while launch is in progress."
							: "Select a context before remembering this project."
						: "Dev Context will suggest this context next time and still show its launch safety checks."}
				</span>
			</span>
		</Card>
	);
}

function boundContextName(
	binding: ProjectBindingState,
	contexts: ContextState[],
): string | undefined {
	if (!binding.bound || binding.dangling || binding.contextId === undefined) {
		return undefined;
	}

	const context = contexts.find(
		(candidate) => candidate.id === binding.contextId,
	);
	return context?.name ?? binding.contextId;
}

function canRememberProject(binding: ProjectBindingState): boolean {
	return !binding.bound && !binding.dangling;
}

export { boundContextName, canRememberProject, RememberProjectControl };
