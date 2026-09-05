import { Button } from "../ui/button.js";

interface ContextCreateIdentityScreenProps {
	name: string;
	onNameChange: (name: string) => void;
	onContinue: () => void;
}

function ContextCreateIdentityScreen({
	name,
	onNameChange,
	onContinue,
}: ContextCreateIdentityScreenProps) {
	const canContinue = name.trim().length > 0;

	return (
		<section
			aria-labelledby="context-identity-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Create a context
				</p>
				<h2 id="context-identity-title" className="text-2xl font-semibold">
					What kind of development identity are you creating?
				</h2>
				<p className="text-sm text-muted-foreground">
					Give it a name. You can choose projects and development tools next.
				</p>
			</div>

			<label className="block space-y-2 text-sm font-medium">
				Context name
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 md:text-sm"
					value={name}
					onChange={(event) => onNameChange(event.target.value)}
					placeholder="For example, Personal"
					autoComplete="off"
				/>
			</label>

			<Button type="button" disabled={!canContinue} onClick={onContinue}>
				Continue
			</Button>
		</section>
	);
}

export type { ContextCreateIdentityScreenProps };
export { ContextCreateIdentityScreen };
