import {
	Building2,
	CheckCircle2,
	Copy,
	Folder,
	MoreVertical,
	Play,
	Plus,
	UserRound,
} from "lucide-react";
import { type ReactNode, useId, useState } from "react";

import type {
	DisplayError,
	HomeDashboardState,
	RecentProjectState,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";

interface HomeViewProps {
	dashboard: HomeDashboardState;
	isFirstRun: boolean;
	launchPending: boolean;
	launchError?: DisplayError;
	onQuickLaunch: () => void;
	onReviewLaunchOptions: () => void;
	onRecentProjectSelect?: (project: RecentProjectState) => void;
}

export function HomeView(props: HomeViewProps) {
	if (props.isFirstRun) {
		return <FirstRunHome {...props} />;
	}

	return (
		<div className="home-content">
			<header className="home-header">
				<div>
					<h2 id="home-heading" className="home-title">
						Welcome back
					</h2>
					<p className="home-subtitle">
						Choose a context and launch a trusted, isolated development
						environment.
					</p>
				</div>
				<div className="home-toolbar">
					<button
						type="button"
						className="home-toolbar-button inline-flex items-center gap-2"
						onClick={props.onReviewLaunchOptions}
					>
						<Plus className="size-4" />
						Manage contexts
					</button>
				</div>
			</header>
			<Overview
				dashboard={props.dashboard}
				onReviewLaunchOptions={props.onReviewLaunchOptions}
			/>
			<div className="home-dashboard">
				<div className="home-column">
					<QuickLaunch {...props} />
					<RunningSummary running={props.dashboard.running} />
				</div>
				<RecentProjects
					projects={props.dashboard.recentProjects}
					onSelect={props.onRecentProjectSelect}
				/>
				<RecentActivity projects={props.dashboard.recentProjects} />
			</div>
		</div>
	);
}

function FirstRunHome({
	dashboard,
	onReviewLaunchOptions,
}: Pick<HomeViewProps, "dashboard" | "onReviewLaunchOptions">) {
	return (
		<div className="home-content w-full max-w-[820px] pt-14">
			<header className="max-w-[680px]">
				<h2 id="home-heading" className="home-title">
					Welcome to Dev Context
				</h2>
				<p className="mt-2 text-[15px] font-medium text-foreground">
					Keep different development identities separate.
				</p>
				<p className="mt-2 max-w-[620px] text-[13px] leading-[1.5] text-muted-foreground">
					A context keeps its own tools, AI accounts, and settings. Create as
					many as you need.
				</p>
			</header>

			<p className="mt-6 text-[12px] font-semibold text-muted-foreground">
				Common examples
			</p>
			<section
				className="mt-[10px] grid grid-cols-2 gap-[12px]"
				aria-label="Examples of separate development identities"
			>
				<ContextExplanation
					icon={UserRound}
					title="Personal"
					description="Your personal projects, tools, and accounts."
					tone="personal"
				/>
				<ContextExplanation
					icon={Building2}
					title="Work"
					description="Employer or client projects, tools, and accounts."
					tone="work"
				/>
			</section>

			<div className="mt-6">
				<Button
					type="button"
					className="h-[42px] rounded-[10px] bg-[var(--green-strong)] px-[18px] text-[13px] font-semibold normal-case tracking-normal text-white shadow-[0_1px_2px_rgb(0_0_0/0.12)] transition-[background-color,transform,box-shadow] duration-150 ease-out hover:bg-[var(--green)] active:scale-[0.99] focus-visible:ring-2 focus-visible:ring-[var(--green-strong)]/30 focus-visible:ring-offset-2 motion-reduce:transform-none motion-reduce:transition-none"
					onClick={onReviewLaunchOptions}
				>
					Set up your first context
				</Button>
				<p className="mt-[7px] text-[11px] text-muted-foreground">
					Takes about a minute
				</p>
			</div>

			<section
				className="mt-7 max-w-[620px] rounded-[11px] border border-black/[0.045] bg-white/60 px-[15px] py-[13px] shadow-[0_1px_2px_rgb(0_0_0/0.018)]"
				aria-labelledby="home-current-project-heading"
			>
				<p
					id="home-current-project-heading"
					className="text-xs font-medium text-muted-foreground"
				>
					Current project
				</p>
				<div className="mt-3 flex min-w-0 items-center gap-3">
					<Folder
						className="size-5 shrink-0 text-muted-foreground"
						strokeWidth={1.6}
					/>
					<div className="min-w-0">
						<h3 className="truncate text-base font-semibold">
							{dashboard.project.name}
						</h3>
						<p
							className="mt-1 truncate font-mono text-xs text-muted-foreground"
							title={dashboard.project.path}
						>
							{dashboard.project.path}
						</p>
					</div>
				</div>
			</section>
		</div>
	);
}

function ContextExplanation({
	icon: Icon,
	title,
	description,
	tone,
}: {
	icon: typeof UserRound;
	title: string;
	description: string;
	tone: "personal" | "work";
}) {
	return (
		<div
			className="flex min-h-[84px] gap-3 rounded-[11px] border border-black/[0.05] bg-white/70 px-[14px] py-[13px] shadow-[0_1px_2px_rgb(0_0_0/0.02)]"
		>
			<span
				className={`grid size-8 shrink-0 place-items-center rounded-[8px] ${tone === "personal" ? "bg-[var(--company-soft)] text-[var(--company-accent)]" : "bg-[var(--green-soft)] text-[var(--green-strong)]"}`}
			>
				<Icon className="size-4" />
			</span>
			<div>
				<h3 className="text-sm font-semibold">{title}</h3>
				<p className="mt-[5px] text-[12px] leading-[1.5] text-muted-foreground">
					{description}
				</p>
			</div>
		</div>
	);
}

function Overview({
	dashboard,
	onReviewLaunchOptions,
}: Pick<HomeViewProps, "dashboard" | "onReviewLaunchOptions">) {
	const id = useId();
	const context = dashboard.currentContext;
	const checks = context?.confidence.checks ?? [];
	const statusEntries =
		checks.length > 0
			? checks
			: context
				? [
						{
							label: context.tool.name,
							message: context.tool.message,
							severity: context.tool.status,
						},
					]
				: [];
	return (
		<section
			className="desktop-card home-overview"
			aria-label="Project and current context"
		>
			<div className="home-overview-project">
				<p className="text-xs font-medium text-muted-foreground">
					Selected project
				</p>
				<div className="mt-3 flex items-center gap-3">
					<Folder className="size-7 text-muted-foreground" strokeWidth={1.6} />
					<h3 className="truncate text-[27px] font-semibold tracking-tight">
						{dashboard.project.name}
					</h3>
				</div>
				<div className="mt-2 flex items-center gap-2 text-sm text-muted-foreground">
					<span className="truncate font-mono">{dashboard.project.path}</span>
					<CopyPathButton path={dashboard.project.path} />
				</div>
				<span className="sr-only">Git branch Last opened</span>
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="mt-5 h-8.75 rounded-md px-3 normal-case tracking-normal"
					onClick={onReviewLaunchOptions}
				>
					<Folder className="size-4" />
					Review launch options
				</Button>
			</div>
			<div className="home-overview-context">
				<p className="text-xs font-medium text-muted-foreground">
					Current context
				</p>
				{context ? (
					<>
						<div className="mt-3 flex items-center gap-5">
							<ContextBadge name={context.name} />
							<span
								className={`flex items-center gap-2 text-xs font-medium ${context.confidence.status === "ready" ? "text-success" : context.confidence.status === "blocked" ? "text-destructive" : "text-warning"}`}
							>
								<span className="size-2 rounded-full bg-current" />
								{context.confidence.status === "ready"
									? "All systems ready"
									: context.confidence.status === "blocked"
										? "Setup blocked"
										: "Review setup"}
							</span>
						</div>
						<div className="mt-5 grid grid-cols-4">
							{statusEntries.map((check, index) => (
								<ProviderStatus
									key={id}
									label={check.label}
									detail={check.message}
									status={check.severity}
									bordered={index > 0}
								/>
							))}
						</div>
					</>
				) : (
					<p className="mt-4 text-sm text-muted-foreground">
						<span className="block font-medium text-foreground">
							Choose a context
						</span>
						Select the identity you want to use for this project.
						<button
							type="button"
							onClick={onReviewLaunchOptions}
							className="mt-3 block text-xs font-medium text-(--green-strong) transition-colors duration-150 hover:text-success"
						>
							Choose context
						</button>
					</p>
				)}
			</div>
		</section>
	);
}

function ContextBadge({ name }: { name: string }) {
	const company = name.toLowerCase().includes("company");
	const Icon = company ? Building2 : UserRound;
	return (
		<span
			className={`inline-flex h-10 items-center gap-2 rounded-md px-3 text-[17px] font-semibold ${company ? "home-context-badge-company" : "home-context-badge-personal"}`}
		>
			<Icon className="size-4" />
			{name}
		</span>
	);
}

function CopyPathButton({ path }: { path: string }) {
	const [copied, setCopied] = useState(false);
	async function copyPath() {
		await navigator.clipboard?.writeText(path);
		setCopied(true);
		window.setTimeout(() => setCopied(false), 1_500);
	}
	return (
		<button
			type="button"
			onClick={() => void copyPath()}
			className="shrink-0 text-muted-foreground transition-colors duration-150 hover:text-foreground"
			aria-label="Copy project path"
			title={copied ? "Copied" : "Copy project path"}
		>
			<Copy className="size-3.5" />
		</button>
	);
}
function ProviderStatus({
	label,
	detail,
	bordered,
	status,
}: {
	label: string;
	detail: string;
	bordered?: boolean;
	status: "ready" | "needs_attention" | "blocked";
}) {
	return (
		<div
			className={`min-w-0 px-4 first:pl-0 ${bordered ? "border-l border-border" : ""}`}
		>
			<p className="truncate text-[13px] font-semibold">{label}</p>
			<p
				className="mt-1 truncate text-[11px] text-muted-foreground"
				title={detail}
			>
				{detail}
			</p>
			<p
				className={`mt-5 flex items-center gap-1.5 text-xs font-medium ${status === "ready" ? "text-success" : status === "blocked" ? "text-destructive" : "text-warning"}`}
			>
				<CheckCircle2 className="size-4" />
				{status === "ready"
					? "Ready"
					: status === "blocked"
						? "Blocked"
						: "Needs attention"}
			</p>
		</div>
	);
}

function QuickLaunch({
	dashboard,
	launchPending,
	launchError,
	onQuickLaunch,
	onReviewLaunchOptions,
}: HomeViewProps) {
	const context = dashboard.currentContext;
	const canLaunch =
		context !== undefined && context.confidence.status !== "blocked";
	return (
		<Panel title="Quick launch" labelledBy="home-quick-launch-heading">
			{context ? (
				<LaunchRow
					contextName={context.name}
					launchPending={launchPending}
					disabled={!canLaunch}
					onClick={onQuickLaunch}
				/>
			) : (
				<button
					type="button"
					className="home-row-button flex w-full items-center gap-3 text-left"
					onClick={onReviewLaunchOptions}
				>
					<span className="grid size-8.5 place-items-center rounded-full home-context-badge-personal">
						<UserRound className="size-4" />
					</span>
					<span className="min-w-0 flex-1">
						<span className="block text-xs font-medium">
							Set up your first context
						</span>
						<span className="mt-0.5 block text-[10px] text-muted-foreground">
							Keep your work and personal tools separate.
						</span>
					</span>
				</button>
			)}
			{launchError ? (
				<p className="mt-2 text-xs text-destructive" role="alert">
					{launchError.message}
				</p>
			) : null}
		</Panel>
	);
}

function LaunchRow({
	contextName,
	launchPending,
	disabled,
	onClick,
}: {
	contextName: string;
	launchPending: boolean;
	disabled: boolean;
	onClick: () => void;
}) {
	const company = contextName.toLowerCase().includes("company");
	const Icon = company ? Building2 : UserRound;
	return (
		<button
			type="button"
			className="home-row-button flex w-full items-center gap-3 text-left disabled:opacity-50"
			disabled={disabled || launchPending}
			onClick={onClick}
		>
			<span
				className={`grid size-8.5 place-items-center rounded-full ${company ? "home-context-badge-company" : "home-context-badge-personal"}`}
			>
				<Icon className="size-4" />
			</span>
			<span className="min-w-0 flex-1">
				<span className="block text-xs font-medium">
					{launchPending
						? `Launching ${contextName}...`
						: `Launch ${contextName}`}
				</span>
				<span className="mt-0.5 block text-[10px] text-muted-foreground">
					Uses your {contextName.toLowerCase()} tools and accounts
				</span>
			</span>
		</button>
	);
}

function RunningSummary({
	running,
}: {
	running: HomeDashboardState["running"];
}) {
	return (
		<Panel
			title="Running environments"
			action="View all"
			labelledBy="home-running-heading"
			fill
		>
			{running.count === 0 ? (
				<p className="text-xs text-muted-foreground">No active environments.</p>
			) : (
				<ul className="divide-y divide-border">
					{running.contextCounts.map((item) => (
						<li
							key={item.contextId}
							className="flex min-h-15.25 items-center gap-2 py-2"
						>
							<span className="size-2 rounded-full bg-success" />
							<span className="grid size-8 place-items-center rounded-md bg-[#e7f0fa] text-[#2674c7]">
								<Play className="size-4 fill-current" />
							</span>
							<span className="min-w-0 flex-1">
								<span className="block truncate text-xs font-medium">
									{item.contextName}
								</span>
								<span className="block text-[10px] text-muted-foreground">
									{item.count} active environment{item.count === 1 ? "" : "s"}
								</span>
							</span>
							<span className="text-right text-[10px] text-muted-foreground">
								VS Code
								<br />
								Running
							</span>
							<MoreVertical className="size-4 text-muted-foreground" />
						</li>
					))}
				</ul>
			)}
		</Panel>
	);
}

function RecentProjects({
	projects,
	onSelect,
}: {
	projects: RecentProjectState[];
	onSelect?: (project: RecentProjectState) => void;
}) {
	return (
		<Panel
			title="Recent projects"
			action="View all"
			labelledBy="home-recent-projects-heading"
			fill
		>
			{projects.length === 0 ? (
				<p className="text-xs text-muted-foreground">
					No recent projects yet. Projects opened through Dev Context will
					appear here.
				</p>
			) : (
				<ul aria-label="Recent projects">
					{projects.slice(0, 5).map((project) => (
						<li key={`${project.project.path}:${project.contextId}`}>
							<button
								type="button"
								disabled={!onSelect}
								onClick={() => onSelect?.(project)}
								className="flex min-h-15.5 w-full items-center gap-3 text-left hover:bg-(--hover) disabled:hover:bg-transparent"
							>
								<Folder
									className="size-5 shrink-0 text-muted-foreground"
									strokeWidth={1.6}
								/>
								<span className="min-w-0 flex-1">
									<span className="block truncate text-xs font-medium">
										{project.project.name}
									</span>
									<span className="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
										{project.project.path}
									</span>
								</span>
								<ContextLabel name={project.contextName ?? project.contextId} />
								<span className="w-9 text-right text-[10px] text-muted-foreground">
									{formatRecentProjectTime(project.lastLaunchedAt)}
								</span>
							</button>
						</li>
					))}
				</ul>
			)}
			<p className="sr-only">Review a project before launching it.</p>
		</Panel>
	);
}

function ContextLabel({ name }: { name: string }) {
	const company = name.toLowerCase().includes("company");
	const Icon = company ? Building2 : UserRound;
	return (
		<span
			className={`inline-flex h-5.75 items-center gap-1 rounded-md px-2 text-[9px] font-medium ${company ? "home-context-badge-company" : "home-context-badge-personal"}`}
		>
			<Icon className="size-3" />
			{name}
		</span>
	);
}

function RecentActivity({
	projects: _projects,
}: {
	projects: RecentProjectState[];
}) {
	return (
		<Panel title="Recent activity" labelledBy="home-activity-heading" fill>
			<p className="text-xs text-muted-foreground">No activity yet.</p>
		</Panel>
	);
}
function Panel({
	title,
	action,
	labelledBy,
	fill,
	children,
}: {
	title: string;
	action?: string;
	labelledBy: string;
	fill?: boolean;
	children: ReactNode;
}) {
	return (
		<section
			className={`desktop-card home-panel ${fill ? "flex flex-1 flex-col" : ""}`}
			aria-labelledby={labelledBy}
		>
			<div className="mb-3 flex items-center justify-between">
				<h3 id={labelledBy} className="home-panel-title">
					{title}
				</h3>
				{action ? (
					<button
						type="button"
						className="text-[11px] text-muted-foreground hover:text-foreground"
					>
						{action}
					</button>
				) : null}
			</div>
			{children}
		</section>
	);
}

export function formatRecentProjectTime(value: string): string {
	const time = new Date(value);
	if (Number.isNaN(time.getTime())) return "Unknown";
	const minutes = Math.round((Date.now() - time.getTime()) / 60_000);
	if (minutes >= 0 && minutes < 60) return `${minutes || 1}m ago`;
	if (minutes >= 60 && minutes < 1440)
		return `${Math.round(minutes / 60)}h ago`;
	return `${Math.round(minutes / 1440)}d ago`;
}
export function homeConfidenceSummary(
	status: "ready" | "needs_attention" | "blocked",
) {
	return status === "blocked"
		? "This context is blocked until its required setup is resolved."
		: status === "needs_attention"
			? "This context can launch, but review its setup before continuing."
			: "All systems required for this context are ready.";
}
