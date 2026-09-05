import { Button } from "../ui/button.js";
import { developmentToolCategories } from "./development-tool-categories.js";

interface ContextCreateDevelopmentToolsScreenProps {
	onBack?: () => void;
	onContinue: () => void;
}

function ContextCreateDevelopmentToolsScreen({
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

			<ul
				className="grid gap-3 sm:grid-cols-2"
				aria-label="Development tool categories"
			>
				{developmentToolCategories.map((category) => (
					<li key={category.id} className="rounded-lg border bg-card p-4">
						<h3 className="text-sm font-medium">{category.name}</h3>
						<p className="mt-1 text-sm text-muted-foreground">
							{category.description}
						</p>
					</li>
				))}
			</ul>

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

export type { ContextCreateDevelopmentToolsScreenProps };
export { ContextCreateDevelopmentToolsScreen };
