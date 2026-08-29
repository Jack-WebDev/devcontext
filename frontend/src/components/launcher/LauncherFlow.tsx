interface LauncherFlowProps {
	projectPath: string;
}

// LauncherFlow is intentionally separate from the management shell. Later
// launcher phases add resolution and selection states inside this focused
// surface without bringing management navigation into a project launch.
function LauncherFlow({ projectPath }: LauncherFlowProps) {
	return (
		<main
			className="flex min-h-screen items-center justify-center bg-background p-6"
			data-launcher-flow
		>
			<section
				aria-labelledby="launcher-heading"
				className="w-full max-w-xl border border-border bg-card p-6 shadow-sm"
			>
				<p className="text-sm font-semibold text-muted-foreground">Dev Context</p>
				<h1 id="launcher-heading" className="mt-2 text-2xl font-semibold">
					Open project
				</h1>
				<p
					className="mt-4 break-all font-mono text-sm text-muted-foreground"
					data-launcher-project-path
				>
					{projectPath}
				</p>
				<p className="mt-6 text-sm text-muted-foreground">
					Preparing launch options...
				</p>
			</section>
		</main>
	);
}

export { LauncherFlow };
