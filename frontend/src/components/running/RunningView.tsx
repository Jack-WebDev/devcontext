import type { RunningEnvironmentState } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface RunningViewProps {
	environments: RunningEnvironmentState[];
	onReveal?: (environment: RunningEnvironmentState) => void;
	onSwitchTo?: (environment: RunningEnvironmentState) => void;
	onStop?: (environment: RunningEnvironmentState) => void;
}

function RunningView({
	environments,
	onReveal,
	onSwitchTo,
	onStop,
}: RunningViewProps) {
	return (
		<section aria-labelledby="running-heading" className="space-y-6">
			<div>
				<p className="text-sm text-muted-foreground">Active coding tools</p>
				<h2 id="running-heading" className="text-2xl font-semibold">
					Running
				</h2>
				<p className="mt-1 text-sm text-muted-foreground">
					Each environment keeps the context selected when it was launched.
				</p>
			</div>

			{environments.length === 0 ? (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="p-5 text-sm text-muted-foreground">
						No active environments are recorded. Launch a project to create an
						isolated coding-tool environment.
					</CardContent>
				</Card>
			) : (
				<div className="space-y-4">
					{environments.map((environment) => (
						<RunningEnvironmentCard
							key={environment.id}
							environment={environment}
							onReveal={onReveal}
							onSwitchTo={onSwitchTo}
							onStop={onStop}
						/>
					))}
				</div>
			)}
		</section>
	);
}

function RunningEnvironmentCard({
	environment,
	onReveal,
	onSwitchTo,
	onStop,
}: {
	environment: RunningEnvironmentState;
	onReveal?: (environment: RunningEnvironmentState) => void;
	onSwitchTo?: (environment: RunningEnvironmentState) => void;
	onStop?: (environment: RunningEnvironmentState) => void;
}) {
	return (
		<Card
			as="article"
			hierarchy="secondary"
			className="py-0"
			aria-labelledby={`running-${environment.id}-heading`}
		>
			<CardContent className="space-y-4 p-5">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<h3
							id={`running-${environment.id}-heading`}
							className="truncate text-lg font-semibold"
							title={environment.project.name}
						>
							{environment.project.name}
						</h3>
						<p
							className="mt-1 truncate font-mono text-sm text-muted-foreground"
							title={environment.project.path}
						>
							{environment.project.path}
						</p>
					</div>
					<span className="shrink-0 text-sm font-medium text-accent-company">
						Running
					</span>
				</div>
				<dl className="grid gap-3 border-t border-border pt-4 text-sm sm:grid-cols-3">
					<RunningDetail label="Context" value={environment.context.name} />
					<RunningDetail label="Coding tool" value={environment.tool.name} />
					<RunningDetail
						label="Started"
						value={formatRunningTime(environment.startedAt)}
					/>
				</dl>
				<div className="flex flex-wrap gap-3 border-t border-border pt-4">
					<Button
						type="button"
						variant="outline"
						size="sm"
						disabled={onReveal === undefined}
						title={
							onReveal === undefined
								? "Revealing an environment is not available for this coding tool yet."
								: undefined
						}
						onClick={() => onReveal?.(environment)}
					>
						Reveal
					</Button>
					<Button
						type="button"
						variant="outline"
						size="sm"
						disabled={onSwitchTo === undefined}
						title={
							onSwitchTo === undefined
								? "Switching to an environment is not available for this coding tool yet."
								: undefined
						}
						onClick={() => onSwitchTo?.(environment)}
					>
						Switch to
					</Button>
					<Button
						type="button"
						variant="destructive"
						size="sm"
						disabled={onStop === undefined}
						title={
							onStop === undefined
								? "Stopping an environment is not available for this coding tool yet."
								: undefined
						}
						onClick={() => onStop?.(environment)}
					>
						Stop
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

function RunningDetail({ label, value }: { label: string; value: string }) {
	return (
		<div className="min-w-0">
			<dt className="text-muted-foreground">{label}</dt>
			<dd className="mt-1 truncate font-medium" title={value}>
				{value}
			</dd>
		</div>
	);
}

function formatRunningTime(value: string): string {
	const time = new Date(value);
	return Number.isNaN(time.getTime()) ? "Unavailable" : time.toLocaleString();
}

export type { RunningViewProps };
export { formatRunningTime, RunningView };
