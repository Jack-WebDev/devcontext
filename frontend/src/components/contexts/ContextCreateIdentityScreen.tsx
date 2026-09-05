import { Button } from "../ui/button.js";
import type { CreateContextRequest } from "../../lib/devctx-api";
import {
	contextIdentityTemplates,
	draftFromContextIdentityTemplate,
} from "./context-identity-templates.js";

interface ContextCreateIdentityScreenProps {
	draft: CreateContextRequest;
	onDraftChange: (draft: CreateContextRequest) => void;
	onContinue: () => void;
}

function ContextCreateIdentityScreen({
	draft,
	onDraftChange,
	onContinue,
}: ContextCreateIdentityScreenProps) {
	const canContinue = (draft.name?.trim().length ?? 0) > 0;
	const selectedTemplate = contextIdentityTemplates.find(
		(template) => template.backendTemplateId === draft.templateId,
	);

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
					Start from a familiar identity, then make its name your own.
				</p>
			</div>

			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Start with</legend>
				<div className="grid gap-2 sm:grid-cols-2">
					{contextIdentityTemplates.map((template) => (
						<Button
							key={template.id}
							type="button"
							variant={
								selectedTemplate?.id === template.id ? "default" : "outline"
							}
							className="h-auto justify-start whitespace-normal px-4 py-3 text-left"
							onClick={() =>
								onDraftChange(draftFromContextIdentityTemplate(template))
							}
						>
							<span>
								<span className="block">{template.name}</span>
								<span className="mt-1 block text-xs font-normal opacity-80">
									{template.description}
								</span>
							</span>
						</Button>
					))}
				</div>
			</fieldset>

			<label className="block space-y-2 text-sm font-medium">
				Context name
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 md:text-sm"
					value={draft.name ?? ""}
					onChange={(event) =>
						onDraftChange({ ...draft, name: event.target.value })
					}
					placeholder="For example, Personal"
					autoComplete="off"
				/>
			</label>

			<div
				aria-label="Context preview"
				className="rounded-lg border border-border bg-muted/30 p-4"
			>
				<p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
					Preview
				</p>
				<p className="mt-1 font-semibold">{draft.name || "Untitled context"}</p>
			</div>

			<Button type="button" disabled={!canContinue} onClick={onContinue}>
				Continue
			</Button>
		</section>
	);
}

export type { ContextCreateIdentityScreenProps };
export { ContextCreateIdentityScreen };
