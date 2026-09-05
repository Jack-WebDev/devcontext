import type { CreateContextRequest } from "../../lib/devctx-api";
import {
	contextAccentOption,
	contextIconOption,
} from "./context-identity-options.js";

function ContextPreview({ draft }: { draft: CreateContextRequest }) {
	const icon = contextIconOption(draft.icon);
	const accent = contextAccentOption(draft.accent);
	const name = draft.name?.trim() || "Untitled context";
	const purpose = draft.purpose?.trim();
	const description = draft.description?.trim();
	const identity = [icon?.label, accent?.label].filter(Boolean).join(", ");

	return (
		<article
			aria-label={`Context preview: ${name}${identity ? `, ${identity}` : ""}`}
			aria-live="polite"
			className="rounded-lg border border-border bg-muted/30 p-4"
		>
			<p className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
				Preview
			</p>
			<div className="mt-3 flex items-start gap-3">
				<span
					aria-hidden="true"
					className={`flex size-9 shrink-0 items-center justify-center rounded-full text-base ${accent?.swatchClassName ?? "bg-muted"}`}
				>
					{icon?.symbol ?? "○"}
				</span>
				<div className="min-w-0">
					<h3 className="font-semibold">{name}</h3>
					{purpose ? (
						<p className="mt-1 text-sm text-muted-foreground">{purpose}</p>
					) : null}
					{description ? (
						<p className="mt-1 text-sm text-muted-foreground">{description}</p>
					) : null}
				</div>
			</div>
		</article>
	);
}

export { ContextPreview };
