interface KeyboardShortcutEvent {
  key: string;
  ctrlKey: boolean;
  metaKey: boolean;
  altKey: boolean;
  shiftKey: boolean;
}

interface KeyboardShortcut {
  id: "command_palette" | "select_context";
  label: string;
  matches: (event: KeyboardShortcutEvent) => boolean;
}

const keyboardShortcuts: Record<KeyboardShortcut["id"], KeyboardShortcut> = {
  command_palette: {
    id: "command_palette",
    label: "Ctrl/Cmd+K",
    matches: (event) => event.key.toLowerCase() === "k" && (event.ctrlKey || event.metaKey) && !event.altKey && !event.shiftKey,
  },
  select_context: {
    id: "select_context",
    label: "1-9",
    matches: (event) => /^[1-9]$/.test(event.key) && !event.ctrlKey && !event.metaKey && !event.altKey && !event.shiftKey,
  },
};

function isCommandPaletteShortcut(event: KeyboardShortcutEvent): boolean {
  return keyboardShortcuts.command_palette.matches(event);
}

function contextPositionFromShortcut(event: KeyboardShortcutEvent): number | undefined {
  if (!keyboardShortcuts.select_context.matches(event)) {
    return undefined;
  }

  return Number(event.key) - 1;
}

export { contextPositionFromShortcut, isCommandPaletteShortcut, keyboardShortcuts };
export type { KeyboardShortcut, KeyboardShortcutEvent };
