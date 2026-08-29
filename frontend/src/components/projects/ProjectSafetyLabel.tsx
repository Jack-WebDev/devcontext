function ProjectSafetyLabel({ contextName }: { contextName?: string }) {
	const label = contextName ?? "No remembered context";
	return (
		<span className="inline-flex border border-border bg-muted/30 px-2 py-1 text-xs font-medium text-foreground">
			Context: {label}
		</span>
	);
}

export { ProjectSafetyLabel };
