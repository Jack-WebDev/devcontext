import { Button } from "../ui/button.js";

interface ProjectNotFoundViewProps {
	choosingFolder: boolean;
	onChooseFolder: () => void;
	onCancel: () => void;
}

function ProjectNotFoundView({
	choosingFolder,
	onChooseFolder,
	onCancel,
}: ProjectNotFoundViewProps) {
	return (
		<section aria-labelledby="project-not-found-title">
			<h2 id="project-not-found-title" className="text-lg font-semibold">
				Project not found
			</h2>
			<p className="mt-2 text-sm text-muted-foreground">
				This project folder is no longer available. Choose its new location to
				continue, or cancel this launch.
			</p>
			<div className="mt-6 flex flex-wrap gap-3">
				<Button type="button" onClick={onChooseFolder} disabled={choosingFolder}>
					{choosingFolder ? "Opening folder picker..." : "Choose folder"}
				</Button>
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
			</div>
		</section>
	);
}

export { ProjectNotFoundView };
