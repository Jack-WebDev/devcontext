function ProjectResolvingView() {
	return (
		<section
			aria-live="polite"
			aria-labelledby="project-resolving-title"
			role="status"
		>
			<h2 id="project-resolving-title" className="text-section-title">
				Preparing your project
			</h2>
			<p className="text-body text-secondary mt-2">
				Checking the project, its remembered context, available contexts, and
				readiness.
			</p>
		</section>
	);
}

export { ProjectResolvingView };
