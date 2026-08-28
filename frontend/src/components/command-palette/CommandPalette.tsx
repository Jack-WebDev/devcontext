import {
  Command,
  CommandEmpty,
  CommandInput,
  CommandList,
} from "../ui/command";
import { CommandDialog } from "../ui/command";
import { isCommandPaletteShortcut } from "./shortcut";

interface CommandPaletteProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

function CommandPalette({ open, onOpenChange }: CommandPaletteProps) {
  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Command palette"
      description="Search available actions."
    >
      <Command label="Command palette">
        <CommandInput autoFocus placeholder="Search actions..." />
        <CommandList>
          <CommandEmpty>No actions are available yet.</CommandEmpty>
        </CommandList>
      </Command>
    </CommandDialog>
  );
}

export { CommandPalette };
export type { CommandPaletteProps };
