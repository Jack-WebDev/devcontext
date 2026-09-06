import { useState } from "react";
import type { HistoryCategory, HistoryEntry } from "../../lib/devctx-api";
import { ProjectSafetyLabel } from "../projects/ProjectSafetyLabel.js";
import { Card, CardContent } from "../ui/card.js";
import { Disclosure } from "../ui/disclosure.js";
import { Button } from "../ui/button.js";
import { EmptyState } from "../ui/collection-state.js";

interface HistoryViewProps {
	entries: HistoryEntry[];
	onOpenProjects?: () => void;
}

interface HistoryDateGroup {
	date: string;
	entries: HistoryEntry[];
}

type HistoryFilter =
	| "all"
	| "launches"
	| "context"
	| "project"
	| "repair"
	| "authentication";

function HistoryView({ entries, onOpenProjects }: HistoryViewProps) {
	const [filter, setFilter] = useState<HistoryFilter>("all");
	const [search, setSearch] = useState("");
	const [selectedEntry, setSelectedEntry] = useState<HistoryEntry>();
	const groups = groupHistoryEntriesByDate(
		filterHistoryEntries(entries, filter, search),
	);

	return (
		<section aria-labelledby="history-heading" className="space-y-6">
			<div>
				<p className="text-sm text-muted-foreground">Local activity</p>
				<h2 id="history-heading" className="text-2xl font-semibold">
					History
				</h2>
				<p className="mt-1 text-sm text-muted-foreground">
					Review launches and changes recorded on this device.
				</p>
			</div>

			<div className="grid gap-4 sm:grid-cols-[12rem_minmax(0,1fr)]">
				<label
					className="grid gap-2 text-sm font-medium"
					htmlFor="history-filter"
				>
					Filter history
					<select
						id="history-filter"
						className="h-10 border border-input bg-background px-3 text-sm text-foreground"
						value={filter}
						onChange={(event) =>
							setFilter(event.currentTarget.value as HistoryFilter)
						}
					>
						<option value="all">All activity</option>
						<option value="launches">Launches</option>
						<option value="context">Context changes</option>
						<option value="project">Project changes</option>
						<option value="repair">Repairs</option>
						<option value="authentication">Authentication</option>
					</select>
				</label>
				<label
					className="grid gap-2 text-sm font-medium"
					htmlFor="history-search"
				>
					Search project or context
					<input
						id="history-search"
						type="search"
						className="h-10 border border-input bg-background px-3 text-sm text-foreground"
						value={search}
						onChange={(event) => setSearch(event.currentTarget.value)}
						placeholder="Search by project path or context"
					/>
				</label>
			</div>

			{groups.length === 0 ? (
				<EmptyState
					title={entries.length === 0 ? "No activity yet" : "No matching activity"}
					description={entries.length === 0 ? "No activity has been recorded yet. Launches and context changes will appear here." : "No activity matches the selected filter or search."}
					{...(entries.length === 0 ? { actionLabel: "View projects", onAction: onOpenProjects } : {})}
				/>
			) : (
				<div className="space-y-6">
					{groups.map((group) => (
						<HistoryDateGroupCard
							key={group.date}
							group={group}
							onSelectEntry={setSelectedEntry}
						/>
					))}
				</div>
			)}
			{selectedEntry ? (
				<HistoryEventDetails
					entry={selectedEntry}
					onClose={() => setSelectedEntry(undefined)}
				/>
			) : null}
		</section>
	);
}

function HistoryDateGroupCard({
	group,
	onSelectEntry,
}: {
	group: HistoryDateGroup;
	onSelectEntry: (entry: HistoryEntry) => void;
}) {
	return (
		<section
			aria-labelledby={`history-date-${group.date}`}
			className="space-y-3"
		>
			<h3
				id={`history-date-${group.date}`}
				className="text-sm font-semibold text-muted-foreground"
			>
				{formatHistoryDate(group.date)}
			</h3>
			<Card hierarchy="secondary" className="py-0">
				<CardContent className="divide-y divide-border p-0">
					{group.entries.map((entry, index) => (
						<HistoryEntryRow
							key={historyEntryKey(entry, index)}
							entry={entry}
							onSelect={() => onSelectEntry(entry)}
						/>
					))}
				</CardContent>
			</Card>
		</section>
	);
}

function HistoryEntryRow({
	entry,
	onSelect,
}: {
	entry: HistoryEntry;
	onSelect: () => void;
}) {
	return (
		<article className="space-y-3 p-5">
			<div className="flex flex-wrap items-start justify-between gap-3">
				<div>
					<h4 className="font-medium">{entry.message}</h4>
					<p className="mt-1 text-sm text-muted-foreground">
						{formatHistoryCategory(entry.category)}
					</p>
				</div>
				<time
					className="shrink-0 text-sm text-muted-foreground"
					dateTime={entry.timestamp}
				>
					{formatHistoryTime(entry.timestamp)}
				</time>
			</div>
			<dl className="grid gap-3 text-sm sm:grid-cols-3">
				<HistoryDetail
					label="Project"
					value={entry.projectPath ?? "Not associated with a project"}
					mono={entry.projectPath !== undefined}
				/>
				<div>
					<dt className="text-muted-foreground">Context</dt>
					<dd className="mt-1">
						<ProjectSafetyLabel contextName={entry.contextId} />
					</dd>
				</div>
				<HistoryDetail
					label="Activity"
					value={formatHistoryCategory(entry.category)}
				/>
			</dl>
			<Button type="button" variant="ghost" size="sm" onClick={onSelect}>
				View details
			</Button>
		</article>
	);
}

function HistoryEventDetails({
	entry,
	onClose,
}: {
	entry: HistoryEntry;
	onClose: () => void;
}) {
	const timestamp = formatHistoryTimestamp(entry.timestamp);
	return (
		<Card
			as="section"
			aria-labelledby="history-event-details-title"
			className="border-primary/30 py-0"
		>
			<CardContent className="space-y-5 p-5">
				<div>
					<h3 id="history-event-details-title" className="text-base font-semibold">
						Activity details
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">{entry.message}</p>
				</div>
				<dl className="grid gap-x-4 gap-y-3 text-sm sm:grid-cols-[8rem_minmax(0,1fr)]">
					<HistoryDetail label="Result" value={entry.message} />
					<HistoryDetail label="When" value={timestamp} />
					<HistoryDetail
						label="Project"
						value={entry.projectPath ?? "Not associated with a project"}
						mono={entry.projectPath !== undefined}
					/>
					<HistoryDetail
						label="Context"
						value={entry.contextId ?? "Not associated with a context"}
					/>
					<HistoryDetail
						label="Activity"
						value={formatHistoryCategory(entry.category)}
					/>
				</dl>
				{entry.toolId ? (
					<Disclosure summary="Technical details">
						<HistoryDetail label="Tool" value={entry.toolId} mono />
					</Disclosure>
				) : null}
				<div className="flex justify-end">
					<Button type="button" onClick={onClose}>
						Close
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function HistoryDetail({
	label,
	value,
	mono = false,
}: {
	label: string;
	value: string;
	mono?: boolean;
}) {
	return (
		<div className="min-w-0">
			<dt className="text-muted-foreground">{label}</dt>
			<dd
				className={`mt-1 truncate font-medium${mono ? " font-mono text-xs" : ""}`}
				title={value}
			>
				{value}
			</dd>
		</div>
	);
}

function groupHistoryEntriesByDate(
	entries: HistoryEntry[],
): HistoryDateGroup[] {
	const groups = new Map<string, HistoryEntry[]>();
	const sortedEntries = [...entries].sort(
		(left, right) =>
			historyTimestamp(right.timestamp) - historyTimestamp(left.timestamp),
	);

	for (const entry of sortedEntries) {
		const date = historyDateKey(entry.timestamp);
		const group = groups.get(date);
		if (group === undefined) {
			groups.set(date, [entry]);
		} else {
			group.push(entry);
		}
	}

	return [...groups.entries()].map(([date, groupedEntries]) => ({
		date,
		entries: groupedEntries,
	}));
}

function filterHistoryEntries(
	entries: HistoryEntry[],
	filter: HistoryFilter,
	search: string,
): HistoryEntry[] {
	const query = search.trim().toLocaleLowerCase();
	return entries.filter((entry) => {
		if (!historyFilterIncludes(filter, entry.category)) {
			return false;
		}
		if (query === "") {
			return true;
		}
		return [entry.projectPath, entry.contextId].some((value) =>
			value?.toLocaleLowerCase().includes(query),
		);
	});
}

function historyFilterIncludes(
	filter: HistoryFilter,
	category: HistoryCategory,
): boolean {
	if (filter === "all") {
		return true;
	}
	if (filter === "launches") {
		return (
			category === "launch" ||
			category === "warning" ||
			category === "workspace" ||
			category === "override"
		);
	}
	if (filter === "project") {
		return category === "binding";
	}
	return category === filter;
}

function historyDateKey(timestamp: string): string {
	const date = new Date(timestamp);
	if (Number.isNaN(date.getTime())) {
		return "unknown";
	}
	return [
		date.getFullYear(),
		String(date.getMonth() + 1).padStart(2, "0"),
		String(date.getDate()).padStart(2, "0"),
	].join("-");
}

function historyTimestamp(timestamp: string): number {
	const value = new Date(timestamp).getTime();
	return Number.isNaN(value) ? 0 : value;
}

function formatHistoryDate(date: string): string {
	if (date === "unknown") {
		return "Date unavailable";
	}
	const value = new Date(`${date}T00:00:00`);
	return value.toLocaleDateString(undefined, {
		weekday: "long",
		year: "numeric",
		month: "long",
		day: "numeric",
	});
}

function formatHistoryTime(timestamp: string): string {
	const value = new Date(timestamp);
	return Number.isNaN(value.getTime())
		? "Time unavailable"
		: value.toLocaleTimeString(undefined, {
				hour: "numeric",
				minute: "2-digit",
			});
}

function formatHistoryTimestamp(timestamp: string): string {
	const value = new Date(timestamp);
	return Number.isNaN(value.getTime())
		? "Time unavailable"
		: value.toLocaleString(undefined, {
				year: "numeric",
				month: "long",
				day: "numeric",
				hour: "numeric",
				minute: "2-digit",
			});
}

function formatHistoryCategory(category: HistoryCategory): string {
	const labels: Record<HistoryCategory, string> = {
		launch: "Launch",
		context: "Context change",
		binding: "Project binding",
		repair: "Repair",
		authentication: "Authentication",
		workspace: "Workspace",
		override: "Context override",
		warning: "Warning",
	};
	return labels[category];
}

function historyEntryKey(entry: HistoryEntry, index: number): string {
	return `${entry.timestamp}:${entry.projectPath ?? ""}:${entry.contextId ?? ""}:${index}`;
}

export type { HistoryDateGroup, HistoryFilter, HistoryViewProps };
export {
	filterHistoryEntries,
	formatHistoryDate,
	formatHistoryCategory,
	formatHistoryTime,
	formatHistoryTimestamp,
	groupHistoryEntriesByDate,
	historyFilterIncludes,
	HistoryEventDetails,
	HistoryView,
};
