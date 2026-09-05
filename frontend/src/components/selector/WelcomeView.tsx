import { Button } from "../ui/button.js";
import { Card, CardContent } from "../ui/card.js";

interface WelcomeViewProps {
	onCreateFirstContext: () => void;
}

// WelcomeView intentionally contains no setup choices. It is the calm entry
// point for a new local installation; the existing setup surface follows only
// after the user chooses to create a context.
function WelcomeView({ onCreateFirstContext }: WelcomeViewProps) {
	return (
		<section
			aria-labelledby="welcome-title"
			className="mx-auto max-w-xl space-y-6 py-8"
		>
			<div className="space-y-3">
				<p className="text-sm font-medium text-muted-foreground">Dev Context</p>
				<h2 id="welcome-title" className="text-2xl font-semibold">
					Welcome to Dev Context
				</h2>
				<p className="text-base text-muted-foreground">
					Set up a development context to keep the work you do on this
					computer organized from the start.
				</p>
			</div>

			<Card size="sm" className="border border-border py-0">
				<CardContent className="p-5">
					<h3 className="text-base font-semibold">Start with one context</h3>
					<p className="mt-2 text-sm text-muted-foreground">
						You can add more contexts whenever your work needs them.
					</p>
					<Button type="button" className="mt-5" onClick={onCreateFirstContext}>
						Create Your First Context
					</Button>
				</CardContent>
			</Card>
		</section>
	);
}

export { WelcomeView };
