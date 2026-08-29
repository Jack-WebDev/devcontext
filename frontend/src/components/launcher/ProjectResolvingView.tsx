function ProjectResolvingView() {
	return (
		<section aria-live="polite" aria-labelledby="project-resolving-title" role="status">
			<h2 id="project-resolving-title" className="text-lg font-semibold">
				Preparing your project
			</h2>
			<p className="mt-2 text-sm text-muted-foreground">
				Checking the project, its remembered context, available contexts, and
				readiness.
			</p>
		</section>
	);
}

export { ProjectResolvingView };
