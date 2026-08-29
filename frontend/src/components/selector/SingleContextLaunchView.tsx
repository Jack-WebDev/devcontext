import type { ContextState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";

interface SingleContextLaunchViewProps {
	context: ContextState;
	projectName: string;
	onChooseAnother: () => void;
}

function SingleContextLaunchView({
	context,
	projectName,
	onChooseAnother,
}: SingleContextLaunchViewProps) {
	return (
		<section
			aria-labelledby="single-context-launch-title"
			className="space-y-3"
		>
			<div>
				<h3 id="single-context-launch-title" className="text-lg font-semibold">
					Open {projectName} with {context.name}
				</h3>
				<p className="mt-1 text-sm text-muted-foreground">
					{context.name} is ready to open this project.
				</p>
			</div>
			<Button
				type="button"
				variant="link"
				className="px-0"
				onClick={onChooseAnother}
			>
				Choose another context
			</Button>
		</section>
	);
}

export { SingleContextLaunchView };
