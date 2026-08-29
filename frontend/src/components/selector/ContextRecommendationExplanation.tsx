import type { ContextRecommendation } from "./recommendation.js";

function ContextRecommendationExplanation({
	recommendation,
}: {
	recommendation: ContextRecommendation;
}) {
	return (
		<details className="pointer-events-auto relative z-20 mt-2 text-xs text-muted-foreground">
			<summary className="w-fit cursor-pointer font-medium text-foreground">
				Why this context
			</summary>
			<ul className="mt-2 space-y-1" aria-label={`Why ${recommendation.label}`}>
				{recommendation.reasons.map((reason) => (
					<li key={reason}>{reason}</li>
				))}
			</ul>
		</details>
	);
}

export { ContextRecommendationExplanation };
