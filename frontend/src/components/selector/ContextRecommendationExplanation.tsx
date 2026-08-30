import { Disclosure } from "../ui/disclosure.js";
import type { ContextRecommendation } from "./recommendation.js";

function ContextRecommendationExplanation({
	recommendation,
}: {
	recommendation: ContextRecommendation;
}) {
	return (
		<Disclosure
			summary="Why this context"
			className="pointer-events-auto relative z-20 mt-2 text-xs text-muted-foreground"
		>
			<ul className="mt-2 space-y-1" aria-label={`Why ${recommendation.label}`}>
				{recommendation.reasons.map((reason) => (
					<li key={reason}>{reason}</li>
				))}
			</ul>
		</Disclosure>
	);
}

export { ContextRecommendationExplanation };
