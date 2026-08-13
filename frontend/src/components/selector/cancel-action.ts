interface CancelSelectorDependencies {
  closeSelector: () => Promise<void> | void;
}

async function cancelSelector(dependencies: CancelSelectorDependencies): Promise<void> {
  await dependencies.closeSelector();
}

export { cancelSelector };
