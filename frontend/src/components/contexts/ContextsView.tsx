import type { ContextListItem } from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface ContextsViewProps {
	contexts: ContextListItem[];
	onSelect?: (id: string) => void;
	onNew?: () => void;
}

function ContextsView({ contexts, onSelect, onNew }: ContextsViewProps) {
	return (
		<section
			aria-labelledby="contexts-heading"
			className="page-content page-section-stack"
		>
			<div className="flex items-start justify-between gap-4">
				<div>
					<p className="text-sm text-muted-foreground">
						Development identities
					</p>
					<h2 id="contexts-heading" className="text-2xl font-semibold">
						Contexts
					</h2>
				</div>
				<Button type="button" onClick={onNew}>
					New context
				</Button>
			</div>

			{contexts.length === 0 ? (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="inset-group text-sm text-muted-foreground">
						No contexts are configured yet. Create a context to set up an
						isolated development identity.
					</CardContent>
				</Card>
			) : (
				<div className="grid gap-4 lg:grid-cols-2">
					{contexts.map((item) => (
						<ContextListCard
							key={item.context.id}
							item={item}
							onSelect={onSelect}
						/>
					))}
				</div>
			)}
		</section>
	);
}

function ContextListCard({
	item,
	onSelect,
}: {
	item: ContextListItem;
	onSelect?: (id: string) => void;
}) {
	const { context } = item;
	return (
		<Card
			as="article"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby={`context-${context.id}-heading`}
		>
			<CardContent className="inset-group space-y-4">
				<div className="flex items-start justify-between gap-3">
					<div className="min-w-0">
						<h3
							id={`context-${context.id}-heading`}
							className="truncate text-lg font-semibold"
							title={context.name}
						>
							{context.name}
						</h3>
						{context.description ? (
							<p className="mt-1 text-sm text-muted-foreground">
								{context.description}
							</p>
						) : null}
					</div>
					<StatusIndicator status={context.confidence?.status ?? "blocked"} />
				</div>

				<dl className="grid gap-3 border-t border-border pt-4 text-sm sm:grid-cols-2">
					<ContextListDetail label="Coding tool" value={context.tool.name} />
					<ContextListDetail
						label="Projects"
						value={projectCountLabel(item.projectCount)}
					/>
					<ContextListDetail label="Providers" value={providerSummary(item)} />
					<ContextListDetail
						label="Last used"
						value={formatContextTime(item.lastUsedAt)}
					/>
				</dl>
				{onSelect ? (
					<Button
						type="button"
						variant="outline"
						size="sm"
						onClick={() => onSelect(context.id)}
					>
						View details
					</Button>
				) : null}
			</CardContent>
		</Card>
	);
}

function ContextListDetail({ label, value }: { label: string; value: string }) {
	return (
		<div className="min-w-0">
			<dt className="text-muted-foreground">{label}</dt>
			<dd className="mt-1 truncate font-medium" title={value}>
				{value}
			</dd>
		</div>
	);
}

function providerSummary(item: ContextListItem): string {
	if (item.enabledProviders.length === 0) {
		return "No providers enabled";
	}
	return item.enabledProviders.map((provider) => provider.name).join(", ");
}

function projectCountLabel(count: number): string {
	return `${count} ${count === 1 ? "project" : "projects"}`;
}

function formatContextTime(value: string | undefined): string {
	if (value === undefined) {
		return "Never launched";
	}
	const time = new Date(value);
	return Number.isNaN(time.getTime()) ? "Unavailable" : time.toLocaleString();
}

export { ContextsView, formatContextTime, projectCountLabel, providerSummary };
