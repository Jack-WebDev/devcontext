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
					<h3 className="text-base font-semibold">Your development identity</h3>
					<p className="mt-2 text-sm text-muted-foreground">
						A context is a development identity, not a provider profile. Use
						one for Personal work, Work, a Client, or Open Source projects.
					</p>
				</CardContent>
			</Card>

			<details className="group rounded-lg border border-border bg-card">
				<summary className="cursor-pointer px-5 py-4 text-sm font-semibold marker:content-none">
					What stays separate
				</summary>
				<div className="border-t border-border px-5 py-4 text-sm text-muted-foreground">
					<ul className="space-y-2">
						<li>
							Projects open with the context you choose for that launch.
						</li>
						<li>Account sessions and tool data use separate local storage.</li>
						<li>Tool settings and launch environment are prepared per context.</li>
						<li>
							A launch uses the selected context without changing another
							context.
						</li>
					</ul>
				</div>
			</details>

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
