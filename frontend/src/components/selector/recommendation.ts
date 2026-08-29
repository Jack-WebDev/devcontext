function recommendationReason(
	resolutionSource: string | undefined,
): string | undefined {
	switch (resolutionSource) {
		case "project_binding":
			return "Remembered for this project";
		case "remembered_context":
			return "Remembered context";
		case "last_launch":
			return "Used for the last launch";
		default:
			return undefined;
	}
}

export { recommendationReason };
