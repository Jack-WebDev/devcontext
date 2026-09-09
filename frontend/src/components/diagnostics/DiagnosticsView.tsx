import { useEffect, useState } from "react";

import type {
	ApiResult,
	ContextListItem,
	DiagnosticCheck,
	DiagnosticsState,
	DisplayError,
	RepairAction,
	RepairActionsState,
	RunRepairActionResult,
} from "../../lib/devctx-api";
import { StatusIndicator } from "../status/StatusIndicator.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { Disclosure } from "../ui/disclosure.js";
import { PageHeader } from "../ui/page-header.js";

interface DiagnosticsViewProps {
	contexts: ContextListItem[];
	load: (contextId: string) => Promise<ApiResult<DiagnosticsState>>;
	loadRepairActions: (
		contextId: string,
	) => Promise<ApiResult<RepairActionsState>>;
	runRepairAction: (
		contextId: string,
		actionId: string,
		confirmDestructive: boolean,
	) => Promise<ApiResult<RunRepairActionResult>>;
}

type DiagnosticsLoad =
	| { status: "loading" }
	| { status: "loaded"; data: DiagnosticsState }
	| { status: "error"; error: DisplayError };

function DiagnosticsView({
	contexts,
	load,
	loadRepairActions,
	runRepairAction,
}: DiagnosticsViewProps) {
	const [contextID, setContextID] = useState("");
	const [diagnostics, setDiagnostics] = useState<DiagnosticsLoad>({
		status: "loading",
	});
	const [repairActions, setRepairActions] = useState<RepairAction[]>([]);
	const [repairAction, setRepairAction] = useState<RepairAction>();
	const [repairPending, setRepairPending] = useState(false);
	const [repairError, setRepairError] = useState<DisplayError>();
	const [reportStatus, setReportStatus] = useState<"copied" | "error">();

	useEffect(() => {
		if (
			contextID !== "" &&
			!contexts.some((item) => item.context.id === contextID)
		) {
			setContextID("");
		}
	}, [contextID, contexts]);

	useEffect(() => {
		let active = true;
		setDiagnostics({ status: "loading" });
		load(contextID).then((result) => {
			if (!active) {
				return;
			}
			setDiagnostics(
				result.ok
					? { status: "loaded", data: result.data }
					: { status: "error", error: result.error },
			);
		});
		return () => {
			active = false;
		};
	}, [contextID, load]);

	useEffect(() => {
		if (contextID === "") {
			setRepairActions([]);
			return;
		}
		loadRepairActions(contextID).then((result) =>
			setRepairActions(result.ok ? result.data.actions : []),
		);
	}, [contextID, loadRepairActions]);

	async function run(action: RepairAction, confirmDestructive = false) {
		if (repairPending) {
			return;
		}
		setRepairPending(true);
		setRepairError(undefined);
		try {
			const result = await runRepairAction(
				contextID,
				action.id,
				confirmDestructive,
			);
			if (!result.ok) {
				setRepairError(result.error);
				return;
			}
			setDiagnostics({ status: "loaded", data: result.data.diagnostics });
			setRepairAction(undefined);
			const actions = await loadRepairActions(contextID);
			if (actions.ok) {
				setRepairActions(actions.data.actions);
			}
		} finally {
			setRepairPending(false);
		}
	}

	async function copyReport() {
		if (diagnostics.status !== "loaded") {
			return;
		}
		try {
			await navigator.clipboard.writeText(
				createDiagnosticReport(
					diagnostics.data,
					contexts.find((item) => item.context.id === contextID)?.context.name,
				),
			);
			setReportStatus("copied");
		} catch {
			setReportStatus("error");
		}
	}

	return (
		<section
			aria-labelledby="diagnostics-heading"
			className="page-content page-section-stack"
		>
			<PageHeader
				id="diagnostics-heading"
				eyebrow="Application readiness"
				title="System Health"
				description="Review application-wide readiness, or open a context to check its isolation, files, tools, bindings, and environment."
			/>

			<label
				className="grid max-w-sm gap-2 text-sm font-medium"
				htmlFor="diagnostics-context-select"
			>
				Context
				<select
					id="diagnostics-context-select"
					className="native-control w-full"
					value={contextID}
					onChange={(event) => setContextID(event.currentTarget.value)}
				>
					<option value="">System health</option>
					{contexts.map((item) => (
						<option key={item.context.id} value={item.context.id}>
							{item.context.name}
						</option>
					))}
				</select>
			</label>

			{diagnostics.status === "loaded" && contextID === "" ? (
				<SystemHealthSummary
					groups={diagnostics.data.groups}
					contexts={contexts}
					onOpenContext={setContextID}
				/>
			) : null}
			{diagnostics.status === "loaded" ? (
				<DiagnosticReportAction
					status={reportStatus}
					onCopy={() => void copyReport()}
				/>
			) : null}
			{renderDiagnostics(diagnostics)}
			{repairActions.length > 0 ? (
				<RepairActions
					actions={repairActions}
					pending={repairPending}
					onSelect={(action) =>
						action.destructive ? setRepairAction(action) : void run(action)
					}
				/>
			) : null}
			{repairAction ? (
				<RepairConfirmation
					action={repairAction}
					pending={repairPending}
					error={repairError}
					onCancel={() => !repairPending && setRepairAction(undefined)}
					onConfirm={() => void run(repairAction, true)}
				/>
			) : null}
		</section>
	);
}

function SystemHealthSummary({
	groups,
	contexts,
	onOpenContext,
}: {
	groups: DiagnosticsState["groups"];
	contexts: ContextListItem[];
	onOpenContext: (contextID: string) => void;
}) {
	return (
		<Card
			as="section"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby="system-health-summary-heading"
		>
			<CardContent className="space-y-5 p-5">
				<div>
					<h3
						id="system-health-summary-heading"
						className="text-lg font-semibold"
					>
						Application health
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						Configuration, storage, tools, integrations, and updates for this
						app.
					</p>
				</div>
				<div className="grid gap-3 sm:grid-cols-2">
					{groups.map((group) => (
						<SystemHealthGroup key={group.id} group={group} />
					))}
				</div>
				<section
					className="border-t border-border pt-4"
					aria-labelledby="context-health-heading"
				>
					<h4 id="context-health-heading" className="font-medium">
						Context health
					</h4>
					<p className="mt-1 text-sm text-muted-foreground">
						Context-specific checks, including isolation, are kept separate from
						application health.
					</p>
					{contexts.length === 0 ? (
						<p className="mt-3 text-sm text-muted-foreground">
							No contexts are configured.
						</p>
					) : (
						<div className="mt-3 flex flex-wrap gap-2">
							{contexts.map((item) => (
								<Button
									key={item.context.id}
									type="button"
									variant="outline"
									size="sm"
									onClick={() => onOpenContext(item.context.id)}
								>
									Open {item.context.name} health
								</Button>
							))}
						</div>
					)}
				</section>
			</CardContent>
		</Card>
	);
}

function SystemHealthGroup({
	group,
}: {
	group: DiagnosticsState["groups"][number];
}) {
	const status = diagnosticGroupStatus(group);
	const issueCount = group.checks.filter(
		(check) => check.severity !== "ready",
	).length;
	return (
		<div className="border border-border p-4">
			<div className="flex items-center justify-between gap-3">
				<h4 className="font-medium">{group.label}</h4>
				<StatusIndicator status={status} />
			</div>
			<p className="mt-2 text-sm text-muted-foreground">
				{issueCount === 0
					? "Ready"
					: `${issueCount} item${issueCount === 1 ? "" : "s"} needs attention`}
			</p>
		</div>
	);
}

function diagnosticGroupStatus(
	group: DiagnosticsState["groups"][number],
): DiagnosticCheck["severity"] {
	if (group.checks.some((check) => check.severity === "blocked"))
		return "blocked";
	if (group.checks.some((check) => check.severity === "needs_attention"))
		return "needs_attention";
	return "ready";
}

function RepairActions({
	actions,
	pending,
	onSelect,
}: {
	actions: RepairAction[];
	pending: boolean;
	onSelect: (action: RepairAction) => void;
}) {
	return (
		<Card
			as="section"
			hierarchy="tertiary"
			className="py-0"
			aria-labelledby="repair-actions-heading"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3 id="repair-actions-heading" className="text-lg font-semibold">
						Repair
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						Repair actions affect only this context’s isolated storage.
					</p>
				</div>
				<div className="flex flex-wrap gap-3">
					{actions.map((action) => (
						<Button
							key={action.id}
							type="button"
							variant={action.destructive ? "destructive" : "outline"}
							size="sm"
							disabled={pending}
							onClick={() => onSelect(action)}
						>
							{action.label}
						</Button>
					))}
				</div>
			</CardContent>
		</Card>
	);
}

function RepairConfirmation({
	action,
	pending,
	error,
	onCancel,
	onConfirm,
}: {
	action: RepairAction;
	pending: boolean;
	error?: DisplayError;
	onCancel: () => void;
	onConfirm: () => void;
}) {
	return (
		<Card
			as="section"
			aria-labelledby="repair-confirmation-heading"
			aria-modal="true"
			className="border border-destructive/30 py-0"
			role="dialog"
		>
			<CardContent className="space-y-4 p-5">
				<div>
					<h3
						id="repair-confirmation-heading"
						className="text-lg font-semibold"
					>
						Confirm {action.label}
					</h3>
					<p className="mt-1 text-sm text-muted-foreground">
						{action.description}
					</p>
				</div>
				<p className="text-sm text-destructive">
					This permanently removes the following context-owned items.
				</p>
				{action.targets.length === 0 ? (
					<p className="text-sm text-muted-foreground">
						No files are currently present.
					</p>
				) : (
					<Disclosure summary={`Show ${action.targets.length} affected paths`}>
						<ul className="mt-3 max-h-48 space-y-1 overflow-y-auto border border-border p-3 font-mono text-xs">
							{action.targets.map((target) => (
								<li key={target.path}>
									{target.label} ({target.kind}): {target.path}
								</li>
							))}
						</ul>
					</Disclosure>
				)}
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error.message}
					</p>
				) : null}
				<div className="flex justify-end gap-3">
					<Button
						type="button"
						variant="outline"
						disabled={pending}
						onClick={onCancel}
					>
						Cancel
					</Button>
					<Button
						type="button"
						variant="destructive"
						disabled={pending}
						onClick={onConfirm}
					>
						{pending ? "Resetting..." : "Reset storage"}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function renderDiagnostics(diagnostics: DiagnosticsLoad) {
	if (diagnostics.status === "loading") {
		return (
			<p className="text-sm text-muted-foreground">Running diagnostics...</p>
		);
	}
	if (diagnostics.status === "error") {
		return (
			<p className="text-sm text-destructive" role="alert">
				{diagnostics.error.message}
			</p>
		);
	}
	if (diagnostics.data.groups.length === 0) {
		return (
			<EmptyDiagnostics message="No diagnostics are available for this context." />
		);
	}
	return (
		<div className="space-y-4">
			{diagnostics.data.groups.map((group) => (
				<DiagnosticGroupCard key={group.id} group={group} />
			))}
		</div>
	);
}

function DiagnosticReportAction({
	status,
	onCopy,
}: {
	status?: "copied" | "error";
	onCopy: () => void;
}) {
	return (
		<div className="flex flex-wrap items-center gap-3">
			<Button type="button" variant="outline" size="sm" onClick={onCopy}>
				Copy diagnostic report
			</Button>
			{status === "copied" ? (
				<p className="text-sm text-muted-foreground" role="status">
					Diagnostic report copied.
				</p>
			) : null}
			{status === "error" ? (
				<p className="text-sm text-destructive" role="alert">
					Could not copy the diagnostic report. Try again.
				</p>
			) : null}
		</div>
	);
}

function createDiagnosticReport(
	diagnostics: DiagnosticsState,
	contextName?: string,
): string {
	const scope = contextName ? `Context: ${contextName}` : "System health";
	const lines = ["Dev Context diagnostic report", scope, ""];
	for (const group of diagnostics.groups) {
		lines.push(group.label);
		for (const check of group.checks) {
			lines.push(`- [${check.severity}] ${check.label}: ${check.message}`);
			if (check.actionHint) {
				lines.push(`  Next step: ${check.actionHint}`);
			}
		}
		lines.push("");
	}
	return lines.join("\n").trimEnd();
}

function DiagnosticGroupCard({
	group,
}: {
	group: DiagnosticsState["groups"][number];
}) {
	return (
		<Card
			as="section"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby={`diagnostics-${group.id}-heading`}
		>
			<CardContent className="space-y-4 p-5">
				<h3
					id={`diagnostics-${group.id}-heading`}
					className="text-lg font-semibold"
				>
					{group.label}
				</h3>
				<ul className="divide-y divide-border border-y border-border">
					{group.checks.map((check) => (
						<DiagnosticCheckRow key={check.id} check={check} />
					))}
				</ul>
			</CardContent>
		</Card>
	);
}

function DiagnosticCheckRow({
	check,
}: {
	check: DiagnosticsState["groups"][number]["checks"][number];
}) {
	const visibleDetails = check.details.filter((detail) => !detail.isPath);
	const pathDetails = check.details.filter((detail) => detail.isPath);
	return (
		<li className="space-y-3 py-4">
			<div className="flex items-start justify-between gap-4">
				<div className="min-w-0">
					<h4 className="font-medium">{check.label}</h4>
					<p className="mt-1 text-sm text-muted-foreground">{check.message}</p>
				</div>
				<StatusIndicator status={check.severity} />
			</div>
			{visibleDetails.length > 0 ? (
				<DiagnosticDetails details={visibleDetails} />
			) : null}
			{pathDetails.length > 0 ? (
				<Disclosure summary="Show paths">
					<DiagnosticDetails details={pathDetails} className="mt-3" />
				</Disclosure>
			) : null}
			{check.actionHint ? (
				<p className="text-sm text-muted-foreground">
					Next step: {check.actionHint}
				</p>
			) : null}
		</li>
	);
}

function DiagnosticDetails({
	details,
	className = "",
}: {
	details: { label: string; value: string }[];
	className?: string;
}) {
	return (
		<dl className={`grid gap-2 text-sm ${className}`}>
			{details.map((detail) => (
				<div
					key={`${detail.label}:${detail.value}`}
					className="grid gap-1 sm:grid-cols-[10rem_1fr] sm:gap-3"
				>
					<dt className="text-muted-foreground">{detail.label}</dt>
					<dd className="break-all font-mono text-xs" title={detail.value}>
						{detail.value}
					</dd>
				</div>
			))}
		</dl>
	);
}

function EmptyDiagnostics({ message }: { message: string }) {
	return (
		<Card as="section" hierarchy="secondary" className="py-0">
			<CardContent className="p-5 text-sm text-muted-foreground">
				{message}
			</CardContent>
		</Card>
	);
}

export {
	DiagnosticCheckRow,
	createDiagnosticReport,
	diagnosticGroupStatus,
	DiagnosticsView,
	RepairConfirmation,
	renderDiagnostics,
	SystemHealthSummary,
};
