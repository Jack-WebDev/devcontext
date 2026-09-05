import { useState } from "react";
import { Button } from "../ui/button.js";
import type { DevelopmentToolIntegration } from "../../lib/devctx-api.js";
import { developmentToolCategories } from "./development-tool-categories.js";
import { developmentToolStatusPresentation } from "./development-tool-status.js";

interface ContextCreateDevelopmentToolsScreenProps {
	integrations?: DevelopmentToolIntegration[];
	enabledIntegrationIds?: string[];
	onEnabledIntegrationIdsChange?: (ids: string[]) => void;
	onBack?: () => void;
	onContinue: () => void;
}

function ContextCreateDevelopmentToolsScreen({
	integrations = [],
	enabledIntegrationIds,
	onEnabledIntegrationIdsChange,
	onBack,
	onContinue,
}: ContextCreateDevelopmentToolsScreenProps) {
	const [localEnabledIDs, setLocalEnabledIDs] = useState(() =>
		integrations
			.filter((integration) => integration.enabled)
			.map((integration) => integration.id),
	);
	const [setupIntegrationID, setSetupIntegrationID] = useState<string>();
	const selectedIDs = enabledIntegrationIds ?? localEnabledIDs;
	function updateEnabled(id: string, enabled: boolean) {
		const integration = integrations.find((item) => item.id === id);
		let next = selectedIDs.filter((selectedID) => selectedID !== id);
		if (enabled) {
			// A context launches one coding tool. Selecting another replaces the
			// previous coding-tool choice while other integration types compose.
			if (integration?.category === "coding") {
				next = next.filter(
					(selectedID) =>
						integrations.find((item) => item.id === selectedID)?.category !==
						"coding",
				);
			}
			next.push(id);
		}
		setLocalEnabledIDs(next);
		onEnabledIntegrationIdsChange?.(next);
	}
	const setupIntegration = integrations.find(
		(item) => item.id === setupIntegrationID,
	);
	if (setupIntegration) {
		return (
			<DevelopmentToolSetupScreen
				integration={setupIntegration}
				onBack={() => setSetupIntegrationID(undefined)}
				onSkip={() => setSetupIntegrationID(undefined)}
			/>
		);
	}
	return (
		<section
			aria-labelledby="context-development-tools-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Create a context
				</p>
				<h2
					id="context-development-tools-title"
					className="text-2xl font-semibold"
				>
					Development tools
				</h2>
				<p className="text-sm text-muted-foreground">
					Choose the development tools that belong to this context. You can
					always adjust them later.
				</p>
			</div>

			<div className="space-y-5" aria-label="Development tool categories">
				{developmentToolCategories.map((category) => {
					const categoryIntegrations = integrations.filter(
						(integration) => integration.category === category.id,
					);
					return (
						<section
							key={category.id}
							aria-labelledby={`tool-category-${category.id}`}
						>
							<h3
								id={`tool-category-${category.id}`}
								className="text-sm font-medium"
							>
								{category.name}
							</h3>
							<p className="mt-1 text-sm text-muted-foreground">
								{category.description}
							</p>
							{categoryIntegrations.length > 0 ? (
								<ul className="mt-3 space-y-2">
									{categoryIntegrations.map((integration) => (
										<DevelopmentToolCard
											key={integration.id}
											integration={integration}
											enabled={selectedIDs.includes(integration.id)}
											onEnabledChange={(enabled) =>
												updateEnabled(integration.id, enabled)
											}
											onSetup={() => {
												updateEnabled(integration.id, true);
												setSetupIntegrationID(integration.id);
											}}
										/>
									))}
								</ul>
							) : null}
						</section>
					);
				})}
			</div>

			<div className="flex items-center justify-between gap-3">
				{onBack ? (
					<Button type="button" variant="outline" onClick={onBack}>
						Back
					</Button>
				) : (
					<span />
				)}
				<Button type="button" onClick={onContinue}>
					Continue
				</Button>
			</div>
		</section>
	);
}

function DevelopmentToolCard({
	integration,
	enabled,
	onEnabledChange,
	onSetup,
}: {
	integration: DevelopmentToolIntegration;
	enabled: boolean;
	onEnabledChange: (enabled: boolean) => void;
	onSetup: () => void;
}) {
	const presentation = developmentToolStatusPresentation[integration.status];
	return (
		<li className="rounded-lg border bg-card p-4">
			<div className="flex items-start justify-between gap-3">
				<div>
					<h4 className="text-sm font-medium">{integration.name}</h4>
					<p className="mt-1 text-sm text-muted-foreground">
						{integration.message || presentation.message}
					</p>
				</div>
				<span className="shrink-0 rounded-full bg-muted px-2 py-1 text-xs font-medium">
					{presentation.label}
				</span>
			</div>
			<p className="mt-3 text-xs text-muted-foreground">
				{integration.recoveryHint || presentation.recovery}
			</p>
			<div className="mt-3 flex flex-wrap gap-2">
				<Button
					type="button"
					variant={enabled ? "outline" : "default"}
					size="sm"
					onClick={() => onEnabledChange(!enabled)}
				>
					{enabled
						? "Disable"
						: integration.category === "coding"
							? "Use for launch"
							: "Enable"}
				</Button>
				{enabled && integration.status !== "connected" ? (
					<Button type="button" variant="outline" size="sm" onClick={onSetup}>
						Set up
					</Button>
				) : null}
			</div>
		</li>
	);
}

function DevelopmentToolSetupScreen({
	integration,
	onBack,
	onSkip,
}: {
	integration: DevelopmentToolIntegration;
	onBack: () => void;
	onSkip: () => void;
}) {
	const presentation = developmentToolStatusPresentation[integration.status];
	return (
		<section
			aria-labelledby="development-tool-setup-title"
			className="mx-auto max-w-xl space-y-6"
		>
			<nav
				aria-label="Context setup path"
				className="text-sm text-muted-foreground"
			>
				Context / Development tools / {integration.name}
			</nav>
			<div className="space-y-2">
				<p className="text-sm font-medium text-muted-foreground">
					Set up development tool
				</p>
				<h2
					id="development-tool-setup-title"
					className="text-2xl font-semibold"
				>
					Set up {integration.name}
				</h2>
				<p className="text-sm text-muted-foreground">
					{integration.message || presentation.message}
				</p>
				<p className="text-sm text-muted-foreground">
					{integration.recoveryHint || presentation.recovery}
				</p>
			</div>
			<p className="rounded-lg border bg-muted/30 p-4 text-sm">
				This integration will remain enabled for this context. You can finish
				its setup now or continue creating the context and return later.
			</p>
			<div className="flex items-center justify-between gap-3">
				<Button type="button" variant="outline" onClick={onBack}>
					Back to development tools
				</Button>
				<Button type="button" variant="outline" onClick={onSkip}>
					Skip setup for now
				</Button>
			</div>
		</section>
	);
}

export type { ContextCreateDevelopmentToolsScreenProps };
export { ContextCreateDevelopmentToolsScreen };
