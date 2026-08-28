import type { HistoryEntry } from "../../lib/devctx-api";
import { Card, CardContent } from "../ui/card.js";

interface HistoryViewProps {
  entries: HistoryEntry[];
}

interface HistoryDateGroup {
  date: string;
  entries: HistoryEntry[];
}

function HistoryView({entries}: HistoryViewProps) {
  const groups = groupHistoryEntriesByDate(entries);

  return (
    <section aria-labelledby="history-heading" className="space-y-6">
      <div>
        <p className="text-sm text-muted-foreground">Local activity</p>
        <h2 id="history-heading" className="text-2xl font-semibold">History</h2>
        <p className="mt-1 text-sm text-muted-foreground">Review launches and changes recorded on this device.</p>
      </div>

      {groups.length === 0 ? (
        <Card as="section" hierarchy="secondary" className="py-0">
          <CardContent className="p-5 text-sm text-muted-foreground">
            No activity has been recorded yet. Launches and context changes will appear here.
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-6" aria-label="Activity history">
          {groups.map((group) => <HistoryDateGroupCard key={group.date} group={group} />)}
        </div>
      )}
    </section>
  );
}

function HistoryDateGroupCard({group}: {group: HistoryDateGroup}) {
  return (
    <section aria-labelledby={`history-date-${group.date}`} className="space-y-3">
      <h3 id={`history-date-${group.date}`} className="text-sm font-semibold text-muted-foreground">{formatHistoryDate(group.date)}</h3>
      <Card hierarchy="secondary" className="py-0">
        <CardContent className="divide-y divide-border p-0">
          {group.entries.map((entry, index) => <HistoryEntryRow key={historyEntryKey(entry, index)} entry={entry} />)}
        </CardContent>
      </Card>
    </section>
  );
}

function HistoryEntryRow({entry}: {entry: HistoryEntry}) {
  return (
    <article className="space-y-3 p-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h4 className="font-medium">{entry.message}</h4>
          <p className="mt-1 text-sm text-muted-foreground">{formatHistoryEvent(entry.event)}</p>
        </div>
        <time className="shrink-0 text-sm text-muted-foreground" dateTime={entry.timestamp}>{formatHistoryTime(entry.timestamp)}</time>
      </div>
      <dl className="grid gap-3 text-sm sm:grid-cols-3">
        <HistoryDetail label="Project" value={entry.projectPath ?? "Not associated with a project"} mono={entry.projectPath !== undefined} />
        <HistoryDetail label="Context" value={entry.contextId ?? "Not associated with a context"} />
        <HistoryDetail label="Event" value={formatHistoryEvent(entry.event)} />
      </dl>
    </article>
  );
}

function HistoryDetail({label, value, mono = false}: {label: string; value: string; mono?: boolean}) {
  return (
    <div className="min-w-0">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className={`mt-1 truncate font-medium${mono ? " font-mono text-xs" : ""}`} title={value}>{value}</dd>
    </div>
  );
}

function groupHistoryEntriesByDate(entries: HistoryEntry[]): HistoryDateGroup[] {
  const groups = new Map<string, HistoryEntry[]>();
  const sortedEntries = [...entries].sort((left, right) => historyTimestamp(right.timestamp) - historyTimestamp(left.timestamp));

  for (const entry of sortedEntries) {
    const date = historyDateKey(entry.timestamp);
    const group = groups.get(date);
    if (group === undefined) {
      groups.set(date, [entry]);
    } else {
      group.push(entry);
    }
  }

  return [...groups.entries()].map(([date, groupedEntries]) => ({date, entries: groupedEntries}));
}

function historyDateKey(timestamp: string): string {
  const date = new Date(timestamp);
  if (Number.isNaN(date.getTime())) {
    return "unknown";
  }
  return [date.getFullYear(), String(date.getMonth() + 1).padStart(2, "0"), String(date.getDate()).padStart(2, "0")].join("-");
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
  return value.toLocaleDateString(undefined, {weekday: "long", year: "numeric", month: "long", day: "numeric"});
}

function formatHistoryTime(timestamp: string): string {
  const value = new Date(timestamp);
  return Number.isNaN(value.getTime()) ? "Time unavailable" : value.toLocaleTimeString(undefined, {hour: "numeric", minute: "2-digit"});
}

function formatHistoryEvent(event: string): string {
  if (event === "") {
    return "Activity recorded";
  }
  return event.split("_").map((word) => word.charAt(0).toUpperCase() + word.slice(1)).join(" ");
}

function historyEntryKey(entry: HistoryEntry, index: number): string {
  return `${entry.timestamp}:${entry.event}:${entry.projectPath ?? ""}:${entry.contextId ?? ""}:${index}`;
}

export { HistoryView, formatHistoryDate, formatHistoryEvent, formatHistoryTime, groupHistoryEntriesByDate };
export type { HistoryDateGroup, HistoryViewProps };
