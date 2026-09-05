import type { ProviderCredentialSession } from "../../lib/devctx-api";
import { Badge } from "../ui/badge.js";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

type ProviderSessionAssignment = string | "not_sure" | undefined;
type ProviderSessionAssignments = Record<string, ProviderSessionAssignment>;
interface ProviderSessionAssignmentOption {
	id: string;
	name: string;
}

interface ProviderCredentialClassificationProps {
	sessions: ProviderCredentialSession[];
	assignments: ProviderSessionAssignments;
	assignmentOptions?: ProviderSessionAssignmentOption[];
	disabled?: boolean;
	onAssign: (providerId: string, assignment: ProviderSessionAssignment) => void;
}

function ProviderCredentialClassification({
	sessions,
	assignments,
	assignmentOptions = defaultAssignmentOptions,
	disabled = false,
	onAssign,
}: ProviderCredentialClassificationProps) {
	if (sessions.length === 0) {
		return null;
	}

	return (
		<section
			className="space-y-3"
			aria-label="Assign detected integration sessions"
		>
			<div>
				<h4 className="text-sm font-semibold">Assign detected sessions</h4>
				<p className="mt-1 text-sm text-muted-foreground">
					Dev Context found signed-in local sessions. Assign a verified session
					to the context that should use it, or leave it unassigned.
				</p>
			</div>
			<div className="grid gap-3 sm:grid-cols-2">
				{sessions.map((session) => (
					<ProviderCredentialSessionCard
						key={session.providerId}
						session={session}
						assignment={assignments[session.providerId]}
						assignmentOptions={assignmentOptions}
						disabled={disabled}
						onAssign={onAssign}
					/>
				))}
			</div>
		</section>
	);
}

function ProviderCredentialSessionCard({
	session,
	assignment,
	assignmentOptions,
	disabled,
	onAssign,
}: {
	session: ProviderCredentialSession;
	assignment: ProviderSessionAssignment;
	assignmentOptions: ProviderSessionAssignmentOption[];
	disabled: boolean;
	onAssign: (providerId: string, assignment: ProviderSessionAssignment) => void;
}) {
	return (
		<Card as="article" size="sm" className="border border-border py-0">
			<CardContent className="p-4">
				<div className="flex items-start justify-between gap-3">
					<div>
						<h5 className="text-sm font-semibold">{session.name}</h5>
						<p className="mt-1 text-xs text-muted-foreground">
							{providerSessionSourceLabel(session)}
						</p>
						<ProviderCredentialMetadata session={session} />
					</div>
					<Badge
						variant={assignment === undefined ? "secondary" : "default"}
						className="text-xs font-medium"
					>
						{assignmentLabel(assignment, assignmentOptions)}
					</Badge>
				</div>
				{session.discovered !== false ? (
					<div className="mt-4 flex flex-wrap gap-2">
						{assignmentOptions.map((option) => (
							<Button
								key={option.id}
								type="button"
								variant={assignment === option.id ? "default" : "outline"}
								size="sm"
								disabled={disabled}
								onClick={() => onAssign(session.providerId, option.id)}
							>
								Use with {option.name}
							</Button>
						))}
						<Button
							type="button"
							variant={assignment === "not_sure" ? "default" : "outline"}
							size="sm"
							disabled={disabled}
							onClick={() => onAssign(session.providerId, "not_sure")}
						>
							Not sure
						</Button>
					</div>
				) : (
					<p className="mt-4 text-sm text-muted-foreground">
						This session could not be verified, so it cannot be assigned.
					</p>
				)}
			</CardContent>
		</Card>
	);
}

function ProviderCredentialMetadata({
	session,
}: {
	session: ProviderCredentialSession;
}) {
	const rows = providerCredentialRows(session);

	if (!session.metadataAvailable || rows.length === 0) {
		return (
			<p className="mt-2 text-sm text-muted-foreground">
				Account metadata unavailable. Refresh this provider sign-in, then reopen
				Dev Context to identify the session.
			</p>
		);
	}

	return (
		<dl className="mt-2 space-y-1 text-sm">
			{rows.map((row) => (
				<div key={row.label} className="grid grid-cols-[auto_1fr] gap-x-2">
					<dt className="text-muted-foreground">{row.label}:</dt>
					<dd
						className="min-w-0 truncate font-medium text-foreground"
						title={row.value}
					>
						{row.value}
					</dd>
				</div>
			))}
		</dl>
	);
}

function providerCredentialRows(
	session: ProviderCredentialSession,
): Array<{ label: string; value: string }> {
	return session.fields.filter(rowHasValue);
}

function providerSessionSourceLabel(
	session: ProviderCredentialSession,
): string {
	return `Current global ${session.name} session`;
}

function assignmentLabel(
	assignment: ProviderSessionAssignment,
	options: ProviderSessionAssignmentOption[],
): string {
	if (assignment === undefined) return "Unassigned";
	if (assignment === "not_sure") return "Not sure — unassigned";
	return `Will use with ${options.find((option) => option.id === assignment)?.name ?? "selected context"}`;
}

const defaultAssignmentOptions: ProviderSessionAssignmentOption[] = [
	{ id: "personal", name: "Personal" },
	{ id: "company", name: "Company" },
];

function rowHasValue(row: {
	label: string;
	value?: string;
}): row is { label: string; value: string } {
	return row.value !== undefined && row.value !== "";
}

export type {
	ProviderSessionAssignment,
	ProviderSessionAssignments,
	ProviderSessionAssignmentOption,
};
export { ProviderCredentialClassification };
