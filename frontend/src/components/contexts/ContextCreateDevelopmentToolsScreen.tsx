import { Button } from "../ui/button.js";
import type { DevelopmentToolIntegration } from "../../lib/devctx-api.js";
import { developmentToolCategories } from "./development-tool-categories.js";
import { developmentToolStatusPresentation } from "./development-tool-status.js";

interface ContextCreateDevelopmentToolsScreenProps {
	integrations?: DevelopmentToolIntegration[];
	onBack?: () => void;
	onContinue: () => void;
}

function ContextCreateDevelopmentToolsScreen({
	integrations = [],
	onBack,
	onContinue,
}: ContextCreateDevelopmentToolsScreenProps) {
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
}: {
	integration: DevelopmentToolIntegration;
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
		</li>
	);
}

export type { ContextCreateDevelopmentToolsScreenProps };
export { ContextCreateDevelopmentToolsScreen };
