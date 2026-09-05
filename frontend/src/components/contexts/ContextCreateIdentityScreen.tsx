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

const contextPurposeMaxLength = 120;

const contextIconOptions = [
	{ id: "user", label: "Person", symbol: "●" },
	{ id: "building", label: "Building", symbol: "▦" },
	{ id: "users", label: "People", symbol: "◉" },
	{ id: "code", label: "Code", symbol: "⌘" },
	{ id: "heart", label: "Heart", symbol: "♥" },
	{ id: "sparkles", label: "Sparkles", symbol: "✦" },
] as const;

const contextAccentOptions = [
	{ id: "sage", label: "Sage", swatchClassName: "bg-accent-personal" },
	{
		id: "slate-blue",
		label: "Slate blue",
		swatchClassName: "bg-accent-company",
	},
	{ id: "amber", label: "Amber", swatchClassName: "bg-accent-freelance" },
	{ id: "custom", label: "Orchid", swatchClassName: "bg-accent-custom" },
] as const;

function contextPurposeValidation(purpose: string | undefined): string | undefined {
	if ((purpose?.length ?? 0) > contextPurposeMaxLength) {
		return `Keep the purpose to ${contextPurposeMaxLength} characters or fewer.`;
	}
	return undefined;
}

function ContextCreateIdentityScreen({
	draft,
	onDraftChange,
	onContinue,
}: ContextCreateIdentityScreenProps) {
	const purposeError = contextPurposeValidation(draft.purpose);
	const canContinue =
		(draft.name?.trim().length ?? 0) > 0 && purposeError === undefined;
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

			<label className="block space-y-2 text-sm font-medium">
				Purpose <span className="font-normal text-muted-foreground">(optional)</span>
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm"
					value={draft.purpose ?? ""}
					onChange={(event) =>
						onDraftChange({ ...draft, purpose: event.target.value })
					}
					placeholder="For example, personal projects and experiments"
					aria-invalid={purposeError !== undefined}
					aria-describedby={purposeError ? "context-purpose-error" : undefined}
				/>
			</label>
			{purposeError ? (
				<p id="context-purpose-error" role="alert" className="text-sm text-destructive">
					{purposeError}
				</p>
			) : null}

			<label className="block space-y-2 text-sm font-medium">
				Description <span className="font-normal text-muted-foreground">(optional)</span>
				<textarea
					className="min-h-24 w-full resize-none rounded-lg border border-input bg-card px-3 py-2 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 md:text-sm"
					value={draft.description ?? ""}
					onChange={(event) =>
						onDraftChange({ ...draft, description: event.target.value })
					}
					placeholder="Add any details that will help you recognize this context."
					maxLength={500}
				/>
			</label>

			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Choose an icon</legend>
				<div className="flex flex-wrap gap-2">
					{contextIconOptions.map((icon) => (
						<Button
							key={icon.id}
							type="button"
							variant={draft.icon === icon.id ? "default" : "outline"}
							size="sm"
							aria-label={`${icon.label} icon`}
							aria-pressed={draft.icon === icon.id}
							onClick={() => onDraftChange({ ...draft, icon: icon.id })}
						>
							<span aria-hidden="true" className="text-base leading-none">
								{icon.symbol}
							</span>
						</Button>
					))}
				</div>
			</fieldset>

			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Choose an accent</legend>
				<div className="flex flex-wrap gap-2">
					{contextAccentOptions.map((accent) => (
						<Button
							key={accent.id}
							type="button"
							variant={draft.accent === accent.id ? "default" : "outline"}
							size="sm"
							aria-label={`${accent.label} accent`}
							aria-pressed={draft.accent === accent.id}
							onClick={() => onDraftChange({ ...draft, accent: accent.id })}
						>
							<span
								aria-hidden="true"
								className={`size-3 rounded-full ${accent.swatchClassName}`}
							/>
							{accent.label}
						</Button>
					))}
				</div>
			</fieldset>

			<div
				aria-label="Context preview"
				className="rounded-lg border border-border bg-muted/30 p-4"
			>
				<p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
					Preview
				</p>
				<p className="mt-1 font-semibold">{draft.name || "Untitled context"}</p>
				{draft.purpose?.trim() ? (
					<p className="mt-1 text-sm text-muted-foreground">{draft.purpose}</p>
				) : null}
			</div>

			<Button type="button" disabled={!canContinue} onClick={onContinue}>
				Continue
			</Button>
		</section>
	);
}

export type { ContextCreateIdentityScreenProps };
export {
	ContextCreateIdentityScreen,
	contextAccentOptions,
	contextIconOptions,
	contextPurposeMaxLength,
	contextPurposeValidation,
};
