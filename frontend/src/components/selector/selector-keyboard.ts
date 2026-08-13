interface SelectorKeyboardState {
  selectedContextId?: string;
  launchPending: boolean;
  mismatchDialogOpen: boolean;
}

type EscapeKeyboardAction = "close-dialog" | "close-selector" | "none";

function canLaunchSelectedContextFromKeyboard(state: SelectorKeyboardState): boolean {
  return state.selectedContextId !== undefined && !state.launchPending && !state.mismatchDialogOpen;
}

function escapeKeyboardAction(state: SelectorKeyboardState): EscapeKeyboardAction {
  if (state.launchPending) {
    return "none";
  }

  return state.mismatchDialogOpen ? "close-dialog" : "close-selector";
}

export { canLaunchSelectedContextFromKeyboard, escapeKeyboardAction };
export type { EscapeKeyboardAction, SelectorKeyboardState };
