import { EllipsisIcon } from "lucide-react";
import type { ContextListItem } from "../../lib/devctx-api";
import {
	ContextAccentIndicator,
	contextAccentFromMetadata,
} from "../context-accent/ContextAccent.js";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuItem,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "../ui/dropdown-menu.js";
import { contextIconOption } from "./context-identity-options.js";

type ContextListAction =
	| "open"
	| "edit"
	| "duplicate"
	| "export"
	| "archive"
	| "delete";

type ContextHealth =
	| "healthy"
	| "needs_attention"
	| "setup_incomplete"
	| "unavailable"
	| "archived";

interface ContextsViewProps {
	contexts: ContextListItem[];
	onSelect?: (id: string) => void;
	onNew?: () => void;
	onAction?: (id: string, action: ContextListAction) => void;
}

function ContextsView({
	contexts,
	onSelect,
	onNew,
	onAction,
}: ContextsViewProps) {
	return (
		<section
			aria-labelledby="contexts-heading"
			className="page-content page-section-stack"
		>
			<div className="flex items-start justify-between gap-4">
				<div>
					<p className="text-body text-secondary">Development identities</p>
					<h2 id="contexts-heading" className="text-page-title">
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
							onAction={onAction}
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
	onAction,
}: {
	item: ContextListItem;
	onSelect?: (id: string) => void;
	onAction?: (id: string, action: ContextListAction) => void;
}) {
	const { context } = item;
	const accent = contextAccentFromMetadata(context.metadata?.accent);
	const icon = context.metadata?.icon;
	const health = contextHealth(context);
	return (
		<Card
			as="article"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby={`context-${context.id}-heading`}
		>
			<CardContent className="inset-group space-y-5">
				<div className="flex items-start gap-4">
					<ContextIdentityMark accent={accent} icon={icon} />
					<div className="min-w-0 flex-1">
						<h3
							id={`context-${context.id}-heading`}
							className="truncate text-section-title"
							title={context.name}
						>
							{context.name}
						</h3>
						{context.purpose ? (
							<p className="mt-1 text-sm font-medium text-foreground">
								{context.purpose}
							</p>
						) : null}
						{context.description ? (
							<p className="mt-1 text-sm text-muted-foreground">
								{context.description}
							</p>
						) : null}
					</div>
					{onAction ? (
						<ContextActionsMenu
							context={context}
							onAction={(action) => onAction(context.id, action)}
						/>
					) : null}
				</div>

				<div className="flex flex-wrap items-center gap-x-5 gap-y-2 border-y border-border py-3">
					<div>
						<p className="text-label text-muted-foreground">Health</p>
						<StatusIndicator status={health.status}>
							{health.label}
						</StatusIndicator>
					</div>
					<ContextListDetail
						label="Linked projects"
						value={projectCountLabel(item.projectCount)}
					/>
					<ContextListDetail
						label="Last used"
						value={formatContextTime(item.lastUsedAt)}
					/>
				</div>

				<p className="text-caption text-muted-foreground">
					Integrations: {integrationSummary(item)}
				</p>
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

function ContextActionsMenu({
	context,
	onAction,
}: {
	context: ContextListItem["context"];
	onAction: (action: ContextListAction) => void;
}) {
	return (
		<DropdownMenu>
			<DropdownMenuTrigger
				render={
					<Button
						type="button"
						variant="ghost"
						size="icon-sm"
						aria-label={`Actions for ${context.name}`}
					/>
				}
			>
				<EllipsisIcon />
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end">
				<DropdownMenuItem onClick={() => onAction("open")}>
					Open
				</DropdownMenuItem>
				<DropdownMenuItem onClick={() => onAction("edit")}>
					Edit
				</DropdownMenuItem>
				<DropdownMenuItem onClick={() => onAction("duplicate")}>
					Duplicate
				</DropdownMenuItem>
				<DropdownMenuItem onClick={() => onAction("export")}>
					Export
				</DropdownMenuItem>
				{context.archivedAt ? null : (
					<DropdownMenuItem onClick={() => onAction("archive")}>
						Archive
					</DropdownMenuItem>
				)}
				<DropdownMenuSeparator />
				<DropdownMenuItem
					variant="destructive"
					onClick={() => onAction("delete")}
				>
					Delete
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}

function ContextIdentityMark({
	accent,
	icon,
}: {
	accent: ReturnType<typeof contextAccentFromMetadata>;
	icon?: string;
}) {
	return (
		<span
			className="flex size-11 shrink-0 items-center justify-center rounded-full bg-muted text-lg text-foreground"
			aria-hidden="true"
		>
			<ContextAccentIndicator
				accent={accent}
				className="mr-1 size-2 rounded-full"
			/>
			{contextIconOption(icon)?.symbol ?? "○"}
		</span>
	);
}

function ContextListDetail({ label, value }: { label: string; value: string }) {
	return (
		<div className="min-w-0">
			<p className="text-label text-muted-foreground">{label}</p>
			<p className="mt-1 truncate text-status" title={value}>
				{value}
			</p>
		</div>
	);
}

function integrationSummary(item: ContextListItem): string {
	const providerCount = item.enabledProviders.length;
	const providers = `${providerCount} ${providerCount === 1 ? "provider" : "providers"}`;
	return `${item.context.tool.name} · ${providers} enabled`;
}

function contextHealth(context: ContextListItem["context"]): {
	label: string;
	status: "ready" | "needs_attention" | "not_configured" | "blocked";
	type: ContextHealth;
} {
	if (context.archivedAt) {
		return { type: "archived", label: "Archived", status: "not_configured" };
	}
	if (context.confidence === undefined) {
		return {
			type: "setup_incomplete",
			label: "Setup incomplete",
			status: "not_configured",
		};
	}
	switch (context.confidence.status) {
		case "ready":
			return { type: "healthy", label: "Healthy", status: "ready" };
		case "needs_attention":
			return {
				type: "needs_attention",
				label: "Needs attention",
				status: "needs_attention",
			};
		case "blocked":
			return { type: "unavailable", label: "Unavailable", status: "blocked" };
	}
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

export type { ContextHealth, ContextListAction };
export {
	ContextsView,
	contextHealth,
	formatContextTime,
	integrationSummary,
	projectCountLabel,
};
