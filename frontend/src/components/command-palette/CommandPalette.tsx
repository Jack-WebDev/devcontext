import {
	Command,
	CommandDialog,
	CommandEmpty,
	CommandGroup,
	CommandInput,
	CommandItem,
	CommandList,
} from "../ui/command";
import type { CommandPaletteAction } from "./actions";
import { keyboardShortcuts } from "./shortcut";

interface CommandPaletteProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	launchActions: CommandPaletteAction[];
	navigationActions: CommandPaletteAction[];
}

function CommandPalette({
	open,
	onOpenChange,
	launchActions,
	navigationActions,
}: CommandPaletteProps) {
	function selectAction(action: CommandPaletteAction) {
		onOpenChange(false);
		action.onSelect();
	}

	return (
		<CommandDialog
			open={open}
			onOpenChange={onOpenChange}
			title="Command palette"
			description={`Search available actions. ${keyboardShortcuts.command_palette.label} toggles this palette.`}
		>
			<Command label="Command palette">
				<CommandInput autoFocus placeholder="Search actions..." />
				<CommandList>
					<CommandEmpty>No matching actions.</CommandEmpty>
					{launchActions.length > 0 ? (
						<CommandGroup heading="Launch">
							{launchActions.map((action) => (
								<CommandItem
									key={action.id}
									value={[action.label, ...(action.keywords ?? [])].join(" ")}
									disabled={action.disabled}
									onSelect={() => selectAction(action)}
								>
									{action.label}
								</CommandItem>
							))}
						</CommandGroup>
					) : null}
					<CommandGroup heading="Navigation">
						{navigationActions.map((action) => (
							<CommandItem
								key={action.id}
								value={[action.label, ...(action.keywords ?? [])].join(" ")}
								onSelect={() => selectAction(action)}
							>
								{action.label}
							</CommandItem>
						))}
					</CommandGroup>
				</CommandList>
			</Command>
		</CommandDialog>
	);
}

export type { CommandPaletteProps };
export { CommandPalette };
