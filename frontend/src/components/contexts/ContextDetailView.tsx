import { useEffect, useState } from "react";
import type { ReactNode } from "react";

import type {
	ApiResult,
	ContextDetailsState,
	ContextState,
	DevelopmentToolIntegration,
	DiagnosticGroup,
	HistoryEntry,
	ProjectListItem,
	UpdateContextDevelopmentToolsRequest,
	UnbindProjectRequest,
	UpdateContextAppearanceRequest,
	UpdateContextDetailsRequest,
} from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";
import { ProjectBindingRemovalDialog } from "../projects/ProjectBindingRemovalDialog.js";
import {
	contextAccentOption,
	contextAccentOptions,
	contextIconOption,
	contextIconOptions,
} from "./context-identity-options.js";
import type { ContextDetailDestination } from "../shell/routes.js";

interface ContextDetailViewProps {
	contextId: string;
	destination: ContextDetailDestination;
	onBack: () => void;
	onNavigate: (destination: ContextDetailDestination) => void;
	load: (contextId: string) => Promise<ApiResult<ContextDetailsState>>;
	updateDetails: (
		request: UpdateContextDetailsRequest,
	) => Promise<ApiResult<ContextState>>;
	updateAppearance: (
		request: UpdateContextAppearanceRequest,
	) => Promise<ApiResult<ContextState>>;
	updateDevelopmentTools: (
		request: UpdateContextDevelopmentToolsRequest,
	) => Promise<ApiResult<ContextState>>;
	getProjects: () => Promise<ApiResult<{ projects: ProjectListItem[] }>>;
	unbindProject: (request: UnbindProjectRequest) => Promise<ApiResult<unknown>>;
	onOpenProjects: () => void;
	getHistory: () => Promise<ApiResult<{ entries: HistoryEntry[] }>>;
	getDiagnostics: (request: {
		contextId: string;
	}) => Promise<ApiResult<{ groups: DiagnosticGroup[] }>>;
	onContextUpdated: () => Promise<void>;
}

const contextPurposeMaxLength = 120;
const contextDescriptionMaxLength = 500;

function ContextDetailView({
	contextId,
	destination,
	onBack,
	onNavigate,
	load,
	updateDetails,
	updateAppearance,
	updateDevelopmentTools,
	getProjects,
	unbindProject,
	onOpenProjects,
	getHistory,
	getDiagnostics,
	onContextUpdated,
}: ContextDetailViewProps) {
	const [result, setResult] = useState<ApiResult<ContextDetailsState>>();
	function applyUpdatedContext(context: ContextState) {
		setResult((current) =>
			current?.ok ? { ok: true, data: { ...current.data, context } } : current,
		);
	}

	useEffect(() => {
		let active = true;
		setResult(undefined);
		void load(contextId).then((next) => {
			if (active) setResult(next);
		});
		return () => {
			active = false;
		};
	}, [contextId, load]);

	if (!result) return <ContextDetailLoading />;
	if (!result.ok)
		return (
			<ContextDetailError message={result.error.message} onBack={onBack} />
		);

	const { context } = result.data;
	return (
		<section
			aria-labelledby="context-detail-heading"
			className="page-content page-section-stack"
		>
			<nav
				aria-label="Breadcrumb"
				className="flex items-center gap-2 text-sm text-muted-foreground"
			>
				<button
					type="button"
					className="hover:text-foreground"
					onClick={onBack}
				>
					Contexts
				</button>
				<span aria-hidden="true">/</span>
				<span aria-current="page" className="text-foreground">
					{context.name}
				</span>
			</nav>
			<div className="flex items-start justify-between gap-4">
				<div>
					<p className="text-body text-secondary">Context</p>
					<h2 id="context-detail-heading" className="text-page-title">
						{context.name}
					</h2>
				</div>
				<Button type="button" variant="outline" onClick={onBack}>
					Back to contexts
				</Button>
			</div>
			{destination === "overview" ? (
				<ContextDetailOverview context={context} onNavigate={onNavigate} />
			) : null}
			{destination === "name-purpose" ? (
				<ContextNamePurposeEditor
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async (request) => {
						const saved = await updateDetails({ contextId, ...request });
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
			{destination === "appearance" ? (
				<ContextAppearanceEditor
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async (request) => {
						const saved = await updateAppearance({ contextId, ...request });
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
			{destination === "linked-projects" ? (
				<ContextLinkedProjects
					contextId={contextId}
					getProjects={getProjects}
					onRemove={async (projectPath) => {
						const removed = await unbindProject({ projectPath });
						if (removed.ok) await onContextUpdated();
						return removed;
					}}
					onOpenProjects={onOpenProjects}
				/>
			) : null}
			{destination === "development-tools" ? (
				<ContextDevelopmentToolsEditor
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async (enabledDevelopmentToolIds) => {
						const saved = await updateDevelopmentTools({
							contextId,
							enabledDevelopmentToolIds,
						});
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
			{destination === "launch-preferences" ? (
				<ContextLaunchPreferences
					context={context}
					onCancel={() => onNavigate("overview")}
					onSave={async ({ toolId, executableOverride }) => {
						const enabledDevelopmentToolIds = [
							toolId,
							...(context.developmentTools
								?.filter((item) => item.category !== "coding" && item.enabled)
								.map((item) => item.id) ?? []),
						];
						const saved = await updateDevelopmentTools({
							contextId,
							enabledDevelopmentToolIds,
							...(executableOverride === undefined
								? {}
								: { executableOverride }),
						});
						if (saved.ok) {
							applyUpdatedContext(saved.data);
							await onContextUpdated();
						}
						return saved;
					}}
				/>
			) : null}
			{destination === "environment" ? <ContextEnvironment /> : null}
			{destination === "activity" ? (
				<ContextActivity contextId={contextId} getHistory={getHistory} />
			) : null}
			{destination === "advanced" ? (
				<ContextAdvanced
					contextId={contextId}
					location={result.data.location}
					getDiagnostics={getDiagnostics}
				/>
			) : null}
		</section>
	);
}

function ContextDetailOverview({
	context,
	onNavigate,
}: {
	context: ContextState;
	onNavigate: (destination: ContextDetailDestination) => void;
}) {
	return (
		<div className="grid gap-4 lg:grid-cols-2">
			<DestinationCard
				title="Name & purpose"
				description="Set the name and description people see for this development identity."
				action="Edit name & purpose"
				onClick={() => onNavigate("name-purpose")}
			/>
			<DestinationCard
				title="Linked projects"
				description="Review the projects that normally open with this context."
				action="Manage linked projects"
				onClick={() => onNavigate("linked-projects")}
			/>
			<DestinationCard
				title="Development tools"
				description="Manage the integrations that belong to this context."
				action="Manage development tools"
				onClick={() => onNavigate("development-tools")}
			/>
			<DestinationCard
				title="Launch preferences"
				description="Choose the coding tool this context uses for launches."
				action="Edit launch preferences"
				onClick={() => onNavigate("launch-preferences")}
			/>
			<DestinationCard
				title="Appearance"
				description="Choose an icon and accent that help distinguish this context."
				action="Edit appearance"
				onClick={() => onNavigate("appearance")}
			/>
			<DestinationCard
				title="Environment"
				description="Review the safe configuration that shapes this context’s launches."
				action="View environment"
				onClick={() => onNavigate("environment")}
			/>
			<DestinationCard
				title="Activity"
				description="Review recent launches and changes for this context."
				action="View activity"
				onClick={() => onNavigate("activity")}
			/>
			<DestinationCard
				title="Advanced"
				description="Open diagnostics and implementation details when you need them."
				action="Open advanced"
				onClick={() => onNavigate("advanced")}
			/>
			<Card as="section" hierarchy="secondary" className="py-0 lg:col-span-2">
				<CardContent className="inset-group">
					<h3 className="text-section-title">Current setup</h3>
					<p className="mt-2 text-body text-secondary">
						{context.tool.name} ·{" "}
						{context.providers.filter((provider) => provider.enabled).length}{" "}
						enabled integrations
					</p>
				</CardContent>
			</Card>
		</div>
	);
}

function DestinationCard({
	title,
	description,
	action,
	onClick,
}: {
	title: string;
	description: string;
	action: string;
	onClick: () => void;
}) {
	return (
		<Card as="section" hierarchy="secondary" className="py-0">
			<CardContent className="inset-group">
				<h3 className="text-section-title">{title}</h3>
				<p className="mt-2 text-body text-secondary">{description}</p>
				<Button
					type="button"
					variant="outline"
					size="sm"
					className="mt-5"
					onClick={onClick}
				>
					{action}
				</Button>
			</CardContent>
		</Card>
	);
}

function ContextNamePurposeEditor({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (
		request: Omit<UpdateContextDetailsRequest, "contextId">,
	) => Promise<ApiResult<ContextState>>;
}) {
	const [name, setName] = useState(context.name);
	const [purpose, setPurpose] = useState(context.purpose ?? "");
	const [description, setDescription] = useState(context.description ?? "");
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	const nameError = name.trim() ? undefined : "A context name is required.";
	const purposeError =
		purpose.length > contextPurposeMaxLength
			? `Keep the purpose to ${contextPurposeMaxLength} characters or fewer.`
			: undefined;
	const descriptionError =
		description.length > contextDescriptionMaxLength
			? `Keep the description to ${contextDescriptionMaxLength} characters or fewer.`
			: undefined;
	const validationError = nameError ?? purposeError ?? descriptionError;
	async function submit() {
		if (validationError) return;
		setPending(true);
		setError(undefined);
		const result = await onSave({ name, purpose, description });
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<form
			className="max-w-2xl space-y-5"
			onSubmit={(event) => {
				event.preventDefault();
				void submit();
			}}
		>
			<div>
				<h3 className="text-section-title">Name & purpose</h3>
				<p className="mt-1 text-body text-secondary">
					Update the human-readable identity for this context. Its ID and
					configuration stay unchanged.
				</p>
			</div>
			<Field label="Context name">
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 md:text-sm"
					value={name}
					onChange={(event) => setName(event.target.value)}
					aria-invalid={nameError !== undefined}
					autoComplete="off"
				/>
			</Field>
			<Field label="Purpose" optional>
				<input
					className="h-10 w-full rounded-lg border border-input bg-card px-3 py-1 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm"
					value={purpose}
					onChange={(event) => setPurpose(event.target.value)}
					aria-invalid={purposeError !== undefined}
					maxLength={contextPurposeMaxLength + 1}
				/>
			</Field>
			<Field label="Description" optional>
				<textarea
					className="min-h-24 w-full resize-none rounded-lg border border-input bg-card px-3 py-2 text-base outline-none placeholder:text-muted-foreground hover:border-foreground/20 focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm"
					value={description}
					onChange={(event) => setDescription(event.target.value)}
					aria-invalid={descriptionError !== undefined}
					maxLength={contextDescriptionMaxLength + 1}
				/>
			</Field>
			{validationError || error ? (
				<p role="alert" className="text-sm text-destructive">
					{validationError ?? error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button
					type="submit"
					disabled={pending || validationError !== undefined}
				>
					{pending ? "Saving..." : "Save changes"}
				</Button>
			</div>
		</form>
	);
}

function ContextAppearanceEditor({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (
		request: Omit<UpdateContextAppearanceRequest, "contextId">,
	) => Promise<ApiResult<ContextState>>;
}) {
	const [icon, setIcon] = useState(context.metadata?.icon ?? "");
	const [accent, setAccent] = useState(context.metadata?.accent ?? "");
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	const iconOption = contextIconOption(icon);
	const accentOption = contextAccentOption(accent);
	async function submit() {
		setPending(true);
		setError(undefined);
		const result = await onSave({ icon, accent });
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<form
			className="max-w-2xl space-y-6"
			onSubmit={(event) => {
				event.preventDefault();
				void submit();
			}}
		>
			<div>
				<h3 className="text-section-title">Appearance</h3>
				<p className="mt-1 text-body text-secondary">
					Choose a visual identity without changing this context’s name, tools,
					or project bindings.
				</p>
			</div>
			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Icon</legend>
				<div className="flex flex-wrap gap-2">
					{contextIconOptions.map((option) => (
						<Button
							key={option.id}
							type="button"
							variant={icon === option.id ? "default" : "outline"}
							size="sm"
							aria-label={`${option.label} icon`}
							aria-pressed={icon === option.id}
							onClick={() => setIcon(option.id)}
						>
							<span aria-hidden="true" className="text-base leading-none">
								{option.symbol}
							</span>
						</Button>
					))}
				</div>
			</fieldset>
			<fieldset className="space-y-3">
				<legend className="text-sm font-medium">Accent</legend>
				<div className="flex flex-wrap gap-2">
					{contextAccentOptions.map((option) => (
						<Button
							key={option.id}
							type="button"
							variant={accent === option.id ? "default" : "outline"}
							size="sm"
							aria-pressed={accent === option.id}
							onClick={() => setAccent(option.id)}
						>
							<span
								aria-hidden="true"
								className={`size-3 rounded-full ${option.swatchClassName}`}
							/>
							{option.label}
						</Button>
					))}
				</div>
			</fieldset>
			<article
				aria-label={`Appearance preview for ${context.name}`}
				aria-live="polite"
				className="rounded-lg border border-border bg-muted/30 p-4"
			>
				<p className="text-label text-muted-foreground">Preview</p>
				<div className="mt-3 flex items-center gap-3">
					<span
						aria-hidden="true"
						className={`flex size-10 items-center justify-center rounded-full text-lg ${accentOption?.swatchClassName ?? "bg-muted"}`}
					>
						{iconOption?.symbol ?? "○"}
					</span>
					<span className="font-medium">{context.name}</span>
				</div>
			</article>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button type="submit" disabled={pending}>
					{pending ? "Saving..." : "Save appearance"}
				</Button>
			</div>
		</form>
	);
}

function ContextLinkedProjects({
	contextId,
	getProjects,
	onRemove,
	onOpenProjects,
}: {
	contextId: string;
	getProjects: () => Promise<ApiResult<{ projects: ProjectListItem[] }>>;
	onRemove: (path: string) => Promise<ApiResult<unknown>>;
	onOpenProjects: () => void;
}) {
	const [projects, setProjects] = useState<ProjectListItem[]>();
	const [error, setError] = useState<string>();
	const [projectToRemove, setProjectToRemove] = useState<ProjectListItem>();
	const [removingPath, setRemovingPath] = useState<string>();
	useEffect(() => {
		void getProjects().then((result) =>
			result.ok
				? setProjects(
						result.data.projects.filter(
							(project) => project.contextId === contextId,
						),
					)
				: setError(result.error.message),
		);
	}, [contextId, getProjects]);
	async function remove(path: string): Promise<boolean> {
		setRemovingPath(path);
		try {
			const result = await onRemove(path);
			if (!result.ok) {
				setError(result.error.message);
				return false;
			}
			setProjects((current) =>
				current?.filter((project) => project.project.path !== path),
			);
			return true;
		} finally {
			setRemovingPath(undefined);
		}
	}
	return (
		<section
			className="max-w-3xl space-y-5"
			aria-labelledby="linked-projects-heading"
		>
			<div>
				<h3 id="linked-projects-heading" className="text-section-title">
					Linked projects
				</h3>
				<p className="mt-1 text-body text-secondary">
					Manage linked projects from the Projects area, where their current
					location and context can be reviewed safely.
				</p>
			</div>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			{projects === undefined ? (
				<p className="text-body text-secondary">Loading linked projects…</p>
			) : projects.length === 0 ? (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="inset-group text-body text-secondary">
						No projects are linked to this context.
					</CardContent>
				</Card>
			) : (
				<div className="space-y-3">
					{projects.map((project) => (
						<Card
							key={project.project.path}
							as="article"
							hierarchy="secondary"
							className="py-0"
						>
							<CardContent className="inset-group">
								<h4 className="font-medium">{project.project.name}</h4>
								<p className="mt-1 break-all font-mono text-sm text-muted-foreground">
									{project.project.path}
								</p>
								<div className="mt-4 flex flex-wrap gap-2">
									<Button
										type="button"
										variant="outline"
										size="sm"
										onClick={onOpenProjects}
									>
										Manage project
									</Button>
									<Button
										type="button"
										variant="destructive"
										size="sm"
										onClick={() => setProjectToRemove(project)}
									>
										Remove binding
									</Button>
								</div>
							</CardContent>
						</Card>
					))}
				</div>
			)}
			{projectToRemove ? (
				<ProjectBindingRemovalDialog
					project={projectToRemove}
					pending={removingPath === projectToRemove.project.path}
					onCancel={() =>
						removingPath === undefined && setProjectToRemove(undefined)
					}
					onConfirm={() =>
						void remove(projectToRemove.project.path).then((removed) => {
							if (removed) setProjectToRemove(undefined);
						})
					}
				/>
			) : null}
		</section>
	);
}

function ContextDevelopmentToolsEditor({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (ids: string[]) => Promise<ApiResult<ContextState>>;
}) {
	const [selected, setSelected] = useState(
		() =>
			context.developmentTools
				?.filter((item) => item.enabled)
				.map((item) => item.id) ?? [context.tool.id],
	);
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	function toggle(integration: DevelopmentToolIntegration) {
		setSelected((current) =>
			integration.category === "coding"
				? [
						...current.filter(
							(id) =>
								context.developmentTools?.find((item) => item.id === id)
									?.category !== "coding",
						),
						integration.id,
					]
				: current.includes(integration.id)
					? current.filter((id) => id !== integration.id)
					: [...current, integration.id],
		);
	}
	async function submit() {
		if (
			!selected.some(
				(id) =>
					context.developmentTools?.find((item) => item.id === id)?.category ===
					"coding",
			)
		)
			return;
		setPending(true);
		const result = await onSave(selected);
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<section
			className="max-w-3xl space-y-5"
			aria-labelledby="development-tools-heading"
		>
			<div>
				<h3 id="development-tools-heading" className="text-section-title">
					Development tools
				</h3>
				<p className="mt-1 text-body text-secondary">
					Choose the integrations owned by this context.
				</p>
			</div>
			<div className="space-y-3">
				{(context.developmentTools ?? []).map((integration) => (
					<Card
						key={integration.id}
						as="article"
						hierarchy="secondary"
						className="py-0"
					>
						<CardContent className="inset-group">
							<div className="flex items-start justify-between gap-3">
								<div>
									<h4 className="font-medium">{integration.name}</h4>
									<p className="mt-1 text-sm text-muted-foreground">
										{integration.message}
									</p>
								</div>
								<Button
									type="button"
									variant={
										selected.includes(integration.id) ? "outline" : "default"
									}
									size="sm"
									onClick={() => toggle(integration)}
								>
									{selected.includes(integration.id)
										? "Disable"
										: integration.category === "coding"
											? "Use for launch"
											: "Enable"}
								</Button>
							</div>
							{integration.recoveryHint ? (
								<p className="mt-2 text-sm text-muted-foreground">
									{integration.recoveryHint}
								</p>
							) : null}
						</CardContent>
					</Card>
				))}
			</div>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button type="button" disabled={pending} onClick={() => void submit()}>
					{pending ? "Saving..." : "Save development tools"}
				</Button>
			</div>
		</section>
	);
}

function ContextLaunchPreferences({
	context,
	onCancel,
	onSave,
}: {
	context: ContextState;
	onCancel: () => void;
	onSave: (selection: {
		toolId: string;
		executableOverride?: string;
	}) => Promise<ApiResult<ContextState>>;
}) {
	const [toolId, setToolId] = useState(context.tool.id);
	const [useCustomExecutable, setUseCustomExecutable] = useState(false);
	const [executableOverride, setExecutableOverride] = useState("");
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string>();
	async function submit() {
		setPending(true);
		const result = await onSave({
			toolId,
			...(useCustomExecutable ? { executableOverride } : {}),
		});
		setPending(false);
		if (!result.ok) {
			setError(result.error.message);
			return;
		}
		onCancel();
	}
	return (
		<section
			className="max-w-2xl space-y-5"
			aria-labelledby="launch-preferences-heading"
		>
			<div>
				<h3 id="launch-preferences-heading" className="text-section-title">
					Launch preferences
				</h3>
				<p className="mt-1 text-body text-secondary">
					Choose the coding tool used when this context launches a project.
				</p>
			</div>
			<fieldset className="space-y-2">
				<legend className="text-sm font-medium">Coding tool</legend>
				{context.availableTools.map((tool) => (
					<Button
						key={tool.id}
						type="button"
						variant={toolId === tool.id ? "default" : "outline"}
						className="mr-2"
						onClick={() => setToolId(tool.id)}
					>
						{tool.name}
					</Button>
				))}
			</fieldset>
			<fieldset className="space-y-2 border-t border-border pt-5">
				<legend className="text-sm font-medium">Custom executable</legend>
				<label className="flex items-start gap-3 text-sm">
					<input
						type="checkbox"
						checked={useCustomExecutable}
						onChange={(event) =>
							setUseCustomExecutable(event.currentTarget.checked)
						}
					/>
					<span>
						Use a specific executable for this context.
						<span className="mt-1 block text-muted-foreground">
							Useful when the coding-tool command is not available on PATH.
						</span>
					</span>
				</label>
				{useCustomExecutable ? (
					<label
						className="grid gap-2 text-sm font-medium"
						htmlFor="context-executable-override"
					>
						Executable path
						<input
							id="context-executable-override"
							type="text"
							className="h-10 border border-input bg-background px-3 text-sm text-foreground"
							placeholder="/path/to/code"
							value={executableOverride}
							onChange={(event) =>
								setExecutableOverride(event.currentTarget.value)
							}
						/>
					</label>
				) : null}
			</fieldset>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			<div className="flex justify-end gap-3">
				<Button type="button" variant="outline" onClick={onCancel}>
					Cancel
				</Button>
				<Button type="button" disabled={pending} onClick={() => void submit()}>
					{pending ? "Saving..." : "Save launch preference"}
				</Button>
			</div>
		</section>
	);
}

function ContextEnvironment() {
	return (
		<section
			className="max-w-2xl space-y-4"
			aria-labelledby="context-environment-heading"
		>
			<div>
				<h3 id="context-environment-heading" className="text-section-title">
					Environment
				</h3>
				<p className="mt-1 text-body text-secondary">
					This context’s environment is assembled from its selected tool and
					enabled integrations at launch.
				</p>
			</div>
			<Card as="section" hierarchy="secondary" className="py-0">
				<CardContent className="inset-group">
					<h4 className="font-medium">No custom environment values</h4>
					<p className="mt-2 text-body text-secondary">
						Dev Context does not store user-defined environment values for
						contexts yet. This keeps credentials and private values out of
						context configuration.
					</p>
				</CardContent>
			</Card>
		</section>
	);
}

function ContextActivity({
	contextId,
	getHistory,
}: {
	contextId: string;
	getHistory: () => Promise<ApiResult<{ entries: HistoryEntry[] }>>;
}) {
	const [entries, setEntries] = useState<HistoryEntry[]>();
	const [error, setError] = useState<string>();
	useEffect(() => {
		void getHistory().then((result) =>
			result.ok
				? setEntries(
						result.data.entries.filter(
							(entry) => entry.contextId === contextId,
						),
					)
				: setError(result.error.message),
		);
	}, [contextId, getHistory]);
	return (
		<section
			className="max-w-3xl space-y-4"
			aria-labelledby="context-activity-heading"
		>
			<div>
				<h3 id="context-activity-heading" className="text-section-title">
					Activity
				</h3>
				<p className="mt-1 text-body text-secondary">
					Recent local launches and changes for this context.
				</p>
			</div>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			{entries === undefined ? (
				<p className="text-body text-secondary">Loading activity…</p>
			) : entries.length === 0 ? (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="inset-group text-body text-secondary">
						No activity has been recorded for this context yet.
					</CardContent>
				</Card>
			) : (
				<Card as="section" hierarchy="secondary" className="py-0">
					<CardContent className="divide-y divide-border p-0">
						{entries.map((entry, index) => (
							<article key={`${entry.timestamp}-${index}`} className="p-4">
								<p className="font-medium">{entry.message}</p>
								<p className="mt-1 text-sm text-muted-foreground">
									{new Date(entry.timestamp).toLocaleString()}
								</p>
								{entry.projectPath ? (
									<p className="mt-1 break-all font-mono text-xs text-muted-foreground">
										{entry.projectPath}
									</p>
								) : null}
							</article>
						))}
					</CardContent>
				</Card>
			)}
		</section>
	);
}

function ContextAdvanced({
	contextId,
	location,
	getDiagnostics,
}: {
	contextId: string;
	location: string;
	getDiagnostics: (request: {
		contextId: string;
	}) => Promise<ApiResult<{ groups: DiagnosticGroup[] }>>;
}) {
	const [groups, setGroups] = useState<DiagnosticGroup[]>();
	const [error, setError] = useState<string>();
	useEffect(() => {
		void getDiagnostics({ contextId }).then((result) =>
			result.ok
				? setGroups(result.data.groups)
				: setError(result.error.message),
		);
	}, [contextId, getDiagnostics]);
	return (
		<section
			className="max-w-3xl space-y-4"
			aria-labelledby="context-advanced-heading"
		>
			<div>
				<h3 id="context-advanced-heading" className="text-section-title">
					Advanced
				</h3>
				<p className="mt-1 text-body text-secondary">
					Implementation details and safe diagnostics for troubleshooting.
				</p>
			</div>
			<Card as="section" hierarchy="secondary" className="py-0">
				<CardContent className="inset-group">
					<dl className="space-y-3 text-sm">
						<div>
							<dt className="text-muted-foreground">Context ID</dt>
							<dd className="mt-1 break-all font-mono">{contextId}</dd>
						</div>
						<div>
							<dt className="text-muted-foreground">Local location</dt>
							<dd className="mt-1 break-all font-mono">{location}</dd>
						</div>
					</dl>
				</CardContent>
			</Card>
			{error ? (
				<p role="alert" className="text-sm text-destructive">
					{error}
				</p>
			) : null}
			{groups === undefined ? (
				<p className="text-body text-secondary">Loading diagnostics…</p>
			) : (
				groups.map((group) => (
					<Card
						key={group.id}
						as="section"
						hierarchy="secondary"
						className="py-0"
					>
						<CardContent className="inset-group">
							<h4 className="font-medium">{group.label}</h4>
							<ul className="mt-3 space-y-2 text-sm">
								{group.checks.map((check) => (
									<li key={check.id}>
										<span className="font-medium">{check.label}</span>
										<p className="text-muted-foreground">{check.message}</p>
									</li>
								))}
							</ul>
						</CardContent>
					</Card>
				))
			)}
		</section>
	);
}

function Field({
	label,
	optional,
	children,
}: {
	label: string;
	optional?: boolean;
	children: ReactNode;
}) {
	return (
		<label className="block space-y-2 text-sm font-medium">
			{label}
			{optional ? (
				<span className="font-normal text-muted-foreground"> (optional)</span>
			) : null}
			{children}
		</label>
	);
}
function ContextDetailLoading() {
	return (
		<section className="page-content">
			<p className="text-body text-secondary">Loading context…</p>
		</section>
	);
}
function ContextDetailError({
	message,
	onBack,
}: {
	message: string;
	onBack: () => void;
}) {
	return (
		<section className="page-content space-y-4">
			<p role="alert" className="text-destructive">
				{message}
			</p>
			<Button type="button" variant="outline" onClick={onBack}>
				Back to contexts
			</Button>
		</section>
	);
}

export type { ContextDetailDestination, ContextDetailViewProps };
export {
	ContextAppearanceEditor,
	ContextDetailView,
	ContextNamePurposeEditor,
	contextDescriptionMaxLength,
	contextPurposeMaxLength,
};
