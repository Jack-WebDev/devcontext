import type {
	LaunchVerificationStep,
	LaunchVerificationStepStatus,
	PreflightCheck,
	PreflightGroup,
} from "../../lib/devctx-api";
import { Card } from "../ui/card.js";

interface LaunchVerificationProgressProps {
	projectName: string;
	contextName: string;
	groups?: PreflightGroup[];
	steps?: LaunchVerificationStep[];
}

function LaunchVerificationProgress({
	projectName,
	contextName,
	groups = [],
	steps = [],
}: LaunchVerificationProgressProps) {
	return (
		<Card
			as="section"
			size="sm"
			className="border border-border bg-muted/30 p-4 text-sm"
			aria-labelledby="launch-verification-title"
			aria-live="polite"
			role="status"
		>
			<h3 id="launch-verification-title" className="font-medium">
				Launch verification
			</h3>
			<p className="mt-1 text-muted-foreground">
				Launching {projectName} as {contextName}...
			</p>
			{groups.length > 0 ? (
				<PreflightGroups groups={groups} />
			) : steps.length === 0 ? (
				<p className="mt-3 text-muted-foreground">
					Preparing launch verification...
				</p>
			) : (
				<ol className="mt-3 space-y-3 border-t border-border pt-3">
					{steps.map((step) => (
						<VerificationStepRow key={step.id} step={step} />
					))}
				</ol>
			)}
		</Card>
	);
}

function PreflightGroups({ groups }: { groups: PreflightGroup[] }) {
	return (
		<ol className="mt-3 space-y-3 border-t border-border pt-3">
			{groups.map((group) => (
				<PreflightGroupRow key={group.id} group={group} />
			))}
		</ol>
	);
}

function PreflightGroupRow({ group }: { group: PreflightGroup }) {
	const presentation = verificationStepPresentation(group.status);

	return (
		<li className="min-w-0">
			<div className="flex items-start justify-between gap-3">
				<p className="font-medium">{group.label}</p>
				<span
					className={`shrink-0 text-xs font-medium ${presentation.className}`}
				>
					{group.blocking ? "Fix required" : presentation.label}
				</span>
			</div>
			<p className="mt-1 text-xs text-muted-foreground">{group.message}</p>
			{group.checks.length > 0 ? (
				<details className="mt-2 text-xs">
					<summary className="cursor-pointer text-muted-foreground hover:text-foreground">
						Show details
					</summary>
					<ul className="mt-2 space-y-2 border-l border-border pl-3">
						{group.checks.map((check) => (
							<PreflightCheckRow key={check.id} check={check} />
						))}
					</ul>
				</details>
			) : null}
		</li>
	);
}

function PreflightCheckRow({ check }: { check: PreflightCheck }) {
	const presentation = verificationStepPresentation(check.status);
	return (
		<li>
			<div className="flex items-start justify-between gap-3">
				<p className="font-medium">{check.label}</p>
				<span className={`shrink-0 ${presentation.className}`}>
					{check.blocking ? "Fix required" : presentation.label}
				</span>
			</div>
			<p className="mt-1 text-muted-foreground">{check.message}</p>
			{check.actionHint ? (
				<p className="mt-1 text-muted-foreground">{check.actionHint}</p>
			) : null}
		</li>
	);
}

function VerificationStepRow({ step }: { step: LaunchVerificationStep }) {
	const presentation = verificationStepPresentation(step.status);

	return (
		<li className="flex min-w-0 items-start justify-between gap-3">
			<div className="min-w-0">
				<p className="font-medium">{step.label}</p>
				<p className="mt-1 text-xs text-muted-foreground">{step.message}</p>
			</div>
			<span
				className={`shrink-0 text-xs font-medium ${presentation.className}`}
			>
				{presentation.label}
			</span>
		</li>
	);
}

function verificationStepPresentation(status: LaunchVerificationStepStatus): {
	label: string;
	className: string;
} {
	switch (status) {
		case "ready":
			return { label: "Ready", className: "text-emerald-700" };
		case "needs_attention":
			return { label: "Needs attention", className: "text-amber-700" };
		case "blocked":
			return { label: "Blocked", className: "text-destructive" };
		case "pending":
			return { label: "Pending", className: "text-muted-foreground" };
	}
}

export { LaunchVerificationProgress, verificationStepPresentation };
