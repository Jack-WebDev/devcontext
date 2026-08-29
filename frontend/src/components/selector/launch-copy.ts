function launchActionLabel(contextName: string | undefined): string {
	return contextName ? `Launch ${contextName}` : "Launch";
}

function launchPendingLabel(
	projectName: string | undefined,
	contextName: string | undefined,
): string {
	if (projectName && contextName) {
		return `Launching ${projectName} as ${contextName}...`;
	}
	if (contextName) {
		return `Launching ${contextName}...`;
	}
	return "Launching...";
}

export { launchActionLabel, launchPendingLabel };
