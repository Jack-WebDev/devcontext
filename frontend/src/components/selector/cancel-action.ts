interface CancelSelectorDependencies {
	closeSelector: () => Promise<void> | void;
	canCancel?: boolean;
}

async function cancelSelector(
	dependencies: CancelSelectorDependencies,
): Promise<boolean> {
	if (dependencies.canCancel === false) {
		return false;
	}
	await dependencies.closeSelector();
	return true;
}

export { cancelSelector };
