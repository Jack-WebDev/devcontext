import type { ReactNode } from "react";

import { Button } from "./button.js";
import { Card, CardContent } from "./card.js";
import { Skeleton } from "./skeleton.js";

interface EmptyStateProps {
	title: string;
	description: string;
	actionLabel?: string;
	onAction?: () => void;
}

// Collection states intentionally keep one next action. This avoids a dead-end
// without turning an empty screen into a second navigation surface.
function EmptyState({ title, description, actionLabel, onAction }: EmptyStateProps) {
	return (
		<Card as="section" hierarchy="secondary" className="py-0">
			<CardContent className="space-y-4 p-5">
				<div>
					<h3 className="font-medium text-foreground">{title}</h3>
					<p className="mt-1 text-sm text-muted-foreground">{description}</p>
				</div>
				{actionLabel && onAction ? <Button type="button" size="sm" onClick={onAction}>{actionLabel}</Button> : null}
			</CardContent>
		</Card>
	);
}

function CollectionSkeleton({ label, rows = 3 }: { label: string; rows?: number }) {
	return (
		<section aria-label={label} aria-busy="true" className="space-y-4">
			<span className="sr-only">{label}</span>
			{Array.from({ length: rows }, (_, index) => (
				<Card key={index} hierarchy="secondary" className="py-0" aria-hidden="true">
					<CardContent className="space-y-3 p-5">
						<Skeleton className="h-5 w-1/3" />
						<Skeleton className="h-4 w-2/3" />
						<Skeleton className="h-9 w-28" />
					</CardContent>
				</Card>
			))}
		</section>
	);
}

export type { EmptyStateProps };
export { CollectionSkeleton, EmptyState };
