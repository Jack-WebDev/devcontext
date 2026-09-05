import { useEffect, useState } from "react";
import type { ReactNode } from "react";

import type {
	ApiResult,
	ContextDetailsState,
	ContextState,
	UpdateContextAppearanceRequest,
	UpdateContextDetailsRequest,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import {
	contextAccentOption,
	contextAccentOptions,
	contextIconOption,
	contextIconOptions,
} from "./context-identity-options.js";

type ContextDetailDestination = "overview" | "name-purpose" | "appearance";

interface ContextDetailViewProps {
	contextId: string;
	destination: ContextDetailDestination;
	onBack: () => void;
	onNavigate: (destination: ContextDetailDestination) => void;
	load: (contextId: string) => Promise<ApiResult<ContextDetailsState>>;
	updateDetails: (
		request: UpdateContextDetailsRequest,
	) => Promise<ApiResult<ContextState>>;
	updateAppearance: (
		request: UpdateContextAppearanceRequest,
	) => Promise<ApiResult<ContextState>>;
	onContextUpdated: () => Promise<void>;
}

const contextPurposeMaxLength = 120;
const contextDescriptionMaxLength = 500;

function ContextDetailView({
	contextId,
	destination,
	onBack,
	onNavigate,
	load,
	updateDetails,
	updateAppearance,
	onContextUpdated,
}: ContextDetailViewProps) {
	const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
	function applyUpdatedContext(context: ContextState) {
		setResult((current) =>
			current?.ok ? { ok: true, data: { ...current.data, context } } : current,
		);
	}

	useEffect(() => {
		let active = true;
		setResult(undefined);
		void load(contextId).then((next) => {
			if (active) setResult(next);
		});
		return () => {
			active = false;
		};
	}, [contextId, load]);

	if (!result) return <ContextDetailLoading />;
	if (!result.ok)
		return (
			<ContextDetailError message={result.error.message} onBack={onBack} />
		);

	const { context } = result.data;
	return (
		<section
			aria-labelledby="context-detail-heading"
			className="page-content page-section-stack"
		>
			<nav
				aria-label="Breadcrumb"
				className="flex items-center gap-2 text-sm text-muted-foreground"
			>
				<button
					type="button"
					className="hover:text-foreground"
					onClick={onBack}
				>
					Contexts
				</button>
				<span aria-hidden="true">/</span>
				<span aria-current="page" className="text-foreground">
					{context.name}
				</span>
			</nav>
			<div className="flex items-start justify-between gap-4">
				<div>
					<p className="text-body text-secondary">Context</p>
					<h2 id="context-detail-heading" className="text-page-title">
						{context.name}
					</h2>
				</div>
				<Button type="button" variant="outline" onClick={onBack}>
					Back to contexts
				</Button>
			</div>
			{destination === "overview" ? (
				<ContextDetailOverview context={context} onNavigate={onNavigate} />
			) : null}
			{destination === "name-purpose" ? (
				<ContextNamePurposeEditor
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async (request) => {
						const saved = await updateDetails({ contextId, ...request });
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
			{destination === "appearance" ? (
				<ContextAppearanceEditor
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async (request) => {
						const saved = await updateAppearance({ contextId, ...request });
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
		</section>
	);
}

function ContextDetailOverview({
	context,
	onNavigate,
}: {
	context: ContextState;
	onNavigate: (destination: ContextDetailDestination) => void;
}) {
	return (
		<div className="grid gap-4 lg:grid-cols-2">
			<DestinationCard
				title="Name & purpose"
				description="Set the name and description people see for this development identity."
				action="Edit name & purpose"
				onClick={() => onNavigate("name-purpose")}
			/>
			<DestinationCard
				title="Appearance"
				description="Choose an icon and accent that help distinguish this context."
				action="Edit appearance"
				onClick={() => onNavigate("appearance")}
			/>
			<Card as="section" hierarchy="secondary" className="py-0 lg:col-span-2">
				<CardContent className="inset-group">
					<h3 className="text-section-title">Current setup</h3>
					<p className="mt-2 text-body text-secondary">
						{context.tool.name} ·{" "}
						{context.providers.filter((provider) => provider.enabled).length}{" "}
						enabled integrations
					</p>
				</CardContent>
			</Card>
		</div>
	);
}

function DestinationCard({
	title,
	description,
	action,
	onClick,
}: {
	title: string;
	description: string;
	action: string;
	onClick: () => void;
}) {
	return (
		<Card as="section" hierarchy="secondary" className="py-0">
			<CardContent className="inset-group">
				<h3 className="text-section-title">{title}</h3>
				<p className="mt-2 text-body text-secondary">{description}</p>
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="mt-5"
					onClick={onClick}
				>
					{action}
				</Button>
			</CardContent>
		</Card>
	);
}

function ContextNamePurposeEditor({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (
		request: Omit<UpdateContextDetailsRequest, "contextId">,
	) => Promise<ApiResult<ContextState>>;
}) {
	const [name, setName] = useState(context.name);
	const [purpose, setPurpose] = useState(context.purpose ?? "");
	const [description, setDescription] = useState(context.description ?? "");
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	const nameError = name.trim() ? undefined : "A context name is required.";
	const purposeError =
		purpose.length > contextPurposeMaxLength
			? `Keep the purpose to ${contextPurposeMaxLength} characters or fewer.`
			: undefined;
	const descriptionError =
		description.length > contextDescriptionMaxLength
			? `Keep the description to ${contextDescriptionMaxLength} characters or fewer.`
			: undefined;
	const validationError = nameError ?? purposeError ?? descriptionError;
	async function submit() {
		if (validationError) return;
		setPending(true);
		setError(undefined);
		const result = await onSave({ name, purpose, description });
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<form
			className="max-w-2xl space-y-5"
			onSubmit={(event) => {
				event.preventDefault();
				void submit();
			}}
		>
			<div>
				<h3 className="text-section-title">Name & purpose</h3>
				<p className="mt-1 text-body text-secondary">
					Update the human-readable identity for this context. Its ID and
					configuration stay unchanged.
				</p>
			</div>
			<Field label="Context name">
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 md:text-sm"
					value={name}
					onChange={(event) => setName(event.target.value)}
					aria-invalid={nameError !== undefined}
					autoComplete="off"
				/>
			</Field>
			<Field label="Purpose" optional>
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm"
					value={purpose}
					onChange={(event) => setPurpose(event.target.value)}
					aria-invalid={purposeError !== undefined}
					maxLength={contextPurposeMaxLength + 1}
				/>
			</Field>
			<Field label="Description" optional>
				<textarea
					className="min-h-24 w-full resize-none rounded-lg border border-input bg-card px-3 py-2 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm"
					value={description}
					onChange={(event) => setDescription(event.target.value)}
					aria-invalid={descriptionError !== undefined}
					maxLength={contextDescriptionMaxLength + 1}
				/>
			</Field>
			{validationError || error ? (
				<p role="alert" className="text-sm text-destructive">
					{validationError ?? error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={pending || validationError !== undefined}
				>
					{pending ? "Saving..." : "Save changes"}
				</Button>
			</div>
		</form>
	);
}

function ContextAppearanceEditor({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (
		request: Omit<UpdateContextAppearanceRequest, "contextId">,
	) => Promise<ApiResult<ContextState>>;
}) {
	const [icon, setIcon] = useState(context.metadata?.icon ?? "");
	const [accent, setAccent] = useState(context.metadata?.accent ?? "");
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	const iconOption = contextIconOption(icon);
	const accentOption = contextAccentOption(accent);
	async function submit() {
		setPending(true);
		setError(undefined);
		const result = await onSave({ icon, accent });
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<form
			className="max-w-2xl space-y-6"
			onSubmit={(event) => {
				event.preventDefault();
				void submit();
			}}
		>
			<div>
				<h3 className="text-section-title">Appearance</h3>
				<p className="mt-1 text-body text-secondary">
					Choose a visual identity without changing this context’s name, tools,
					or project bindings.
				</p>
			</div>
			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Icon</legend>
				<div className="flex flex-wrap gap-2">
					{contextIconOptions.map((option) => (
						<Button
							key={option.id}
							type="button"
							variant={icon === option.id ? "default" : "outline"}
							size="sm"
							aria-label={`${option.label} icon`}
							aria-pressed={icon === option.id}
							onClick={() => setIcon(option.id)}
						>
							<span aria-hidden="true" className="text-base leading-none">
								{option.symbol}
							</span>
						</Button>
					))}
				</div>
			</fieldset>
			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Accent</legend>
				<div className="flex flex-wrap gap-2">
					{contextAccentOptions.map((option) => (
						<Button
							key={option.id}
							type="button"
							variant={accent === option.id ? "default" : "outline"}
							size="sm"
							aria-pressed={accent === option.id}
							onClick={() => setAccent(option.id)}
						>
							<span
								aria-hidden="true"
								className={`size-3 rounded-full ${option.swatchClassName}`}
							/>
							{option.label}
						</Button>
					))}
				</div>
			</fieldset>
			<article
				aria-label={`Appearance preview for ${context.name}`}
				aria-live="polite"
				className="rounded-lg border border-border bg-muted/30 p-4"
			>
				<p className="text-label text-muted-foreground">Preview</p>
				<div className="mt-3 flex items-center gap-3">
					<span
						aria-hidden="true"
						className={`flex size-10 items-center justify-center rounded-full text-lg ${accentOption?.swatchClassName ?? "bg-muted"}`}
					>
						{iconOption?.symbol ?? "○"}
					</span>
					<span className="font-medium">{context.name}</span>
				</div>
			</article>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button type="submit" disabled={pending}>
					{pending ? "Saving..." : "Save appearance"}
				</Button>
			</div>
		</form>
	);
}

function Field({
	label,
	optional,
	children,
}: {
	label: string;
	optional?: boolean;
	children: ReactNode;
}) {
	return (
		<label className="block space-y-2 text-sm font-medium">
			{label}
			{optional ? (
				<span className="font-normal text-muted-foreground"> (optional)</span>
			) : null}
			{children}
		</label>
	);
}
function ContextDetailLoading() {
	return (
		<section className="page-content">
			<p className="text-body text-secondary">Loading context…</p>
		</section>
	);
}
function ContextDetailError({
	message,
	onBack,
}: {
	message: string;
	onBack: () => void;
}) {
	return (
		<section className="page-content space-y-4">
			<p role="alert" className="text-destructive">
				{message}
			</p>
			<Button type="button" variant="outline" onClick={onBack}>
				Back to contexts
			</Button>
		</section>
	);
}

export type { ContextDetailDestination, ContextDetailViewProps };
export {
	ContextAppearanceEditor,
	ContextDetailView,
	ContextNamePurposeEditor,
	contextDescriptionMaxLength,
	contextPurposeMaxLength,
};
