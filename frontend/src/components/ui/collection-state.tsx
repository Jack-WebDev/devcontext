import { Inbox } from "lucide-react";

import { Button } from "./button.js";
import { Skeleton } from "./skeleton.js";

interface EmptyStateProps {
	title: string;
	description: string;
	actionLabel?: string;
	onAction?: () => void;
}

// Collection states intentionally keep one next action. This avoids a dead-end
// without turning an empty screen into a second navigation surface.
function EmptyState({
	title,
	description,
	actionLabel,
	onAction,
}: EmptyStateProps) {
	return (
		<section className="flex min-h-64 flex-col items-center justify-center rounded-2xl bg-muted/30 px-8 py-12 text-center">
			<span
				className="mb-4 grid size-11 place-items-center rounded-xl bg-card text-muted-foreground shadow-sm"
				aria-hidden="true"
			>
				<Inbox className="size-5" strokeWidth={1.7} />
			</span>
			<h3 className="text-base font-semibold text-foreground">{title}</h3>
			<p className="mt-1.5 max-w-md text-sm leading-6 text-muted-foreground">
				{description}
			</p>
			{actionLabel && onAction ? (
				<Button type="button" size="sm" className="mt-5" onClick={onAction}>
					{actionLabel}
				</Button>
			) : null}
		</section>
	);
}

function CollectionSkeleton({
	label,
	rows = 3,
}: {
	label: string;
	rows?: number;
}) {
	return (
		<section
			aria-label={label}
			aria-busy="true"
			className="page-content space-y-6"
		>
			<span className="sr-only">{label}</span>
			<div className="space-y-2" aria-hidden="true">
				<Skeleton className="h-3 w-24 rounded-full" />
				<Skeleton className="h-9 w-52 rounded-lg" />
			</div>
			<div
				className="collection-surface divide-y divide-border/50"
				aria-hidden="true"
			>
				{Array.from({ length: rows }, (_, index) => (
					<div key={index} className="space-y-3 p-5">
						<Skeleton className="h-5 w-1/3 rounded-md" />
						<Skeleton className="h-4 w-2/3 rounded-md" />
					</div>
				))}
			</div>
		</section>
	);
}

export type { EmptyStateProps };
export { CollectionSkeleton, EmptyState };
