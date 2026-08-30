import { useEffect } from "react";
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
	useEffect(() => {
		function handleKeyDown(event: KeyboardEvent) {
			if (
				event.key !== "Escape" ||
				projectRecoveryEscapeAction(choosingFolder) === "none"
			) {
				return;
			}
			event.preventDefault();
			onCancel();
		}

		window.addEventListener("keydown", handleKeyDown);
		return () => window.removeEventListener("keydown", handleKeyDown);
	}, [choosingFolder, onCancel]);

	return (
		<section aria-labelledby="project-not-found-title">
			<h2 id="project-not-found-title" className="text-section-title">
				Project not found
			</h2>
			<p className="text-body text-secondary mt-2">
				This project folder is no longer available. Choose its new location to
				continue, or cancel this launch.
			</p>
			<div className="mt-6 flex flex-wrap gap-3">
				<Button
					type="button"
					onClick={onChooseFolder}
					disabled={choosingFolder}
				>
					{choosingFolder ? "Opening folder picker..." : "Choose folder"}
				</Button>
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
			</div>
		</section>
	);
}

function projectRecoveryEscapeAction(
	choosingFolder: boolean,
): "close-selector" | "none" {
	return choosingFolder ? "none" : "close-selector";
}

export { ProjectNotFoundView, projectRecoveryEscapeAction };
