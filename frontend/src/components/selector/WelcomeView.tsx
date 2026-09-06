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
			className="page-reading-column page-section-stack mx-auto py-8"
		>
			<div className="space-y-3">
				<p className="text-label text-secondary">Dev Context</p>
				<h2 id="welcome-title" className="text-page-title">
					Welcome to Dev Context
				</h2>
				<p className="text-body text-secondary">
					Set up a development context to keep the work you do on this
					computer organized from the start.
				</p>
			</div>

			<Card size="sm" hierarchy="secondary" className="py-0">
				<CardContent className="inset-group">
					<h3 className="text-section-title">Your development identity</h3>
					<p className="mt-2 text-detail text-secondary">
						A context is a development identity, not a provider profile. Use
						one for Personal work, Work, a Client, or Open Source projects.
					</p>
				</CardContent>
			</Card>

			<details className="group rounded-xl border border-border bg-[var(--surface-subtle)]">
				<summary className="cursor-pointer px-[var(--layout-card-padding)] py-4 text-detail font-semibold marker:content-none">
					What stays separate
				</summary>
				<div className="border-t border-border px-[var(--layout-card-padding)] py-4 text-detail text-secondary">
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

			<Card size="sm" hierarchy="primary" className="py-0">
				<CardContent className="inset-group">
					<h3 className="text-section-title">Start with one context</h3>
					<p className="mt-2 text-detail text-secondary">
						You can add more contexts whenever your work needs them.
					</p>
					<Button
						type="button"
						variant="primary"
						className="mt-5"
						onClick={onCreateFirstContext}
					>
						Create Your First Context
					</Button>
				</CardContent>
			</Card>
		</section>
	);
}

export { WelcomeView };
