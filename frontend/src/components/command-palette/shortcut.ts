interface CommandPaletteShortcutEvent {
  key: string;
  ctrlKey: boolean;
  metaKey: boolean;
  altKey: boolean;
  shiftKey: boolean;
}

function isCommandPaletteShortcut(event: CommandPaletteShortcutEvent): boolean {
  return event.key.toLowerCase() === "k" && (event.ctrlKey || event.metaKey) && !event.altKey && !event.shiftKey;
}

export { isCommandPaletteShortcut };
export type { CommandPaletteShortcutEvent };
