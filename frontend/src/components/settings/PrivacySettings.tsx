import { privacyStatements } from "./privacy-statements.js";

function PrivacySettings() {
	return (
		<section className="p-6" aria-labelledby="settings-privacy">
			<h3 id="settings-privacy" className="font-semibold">
				Privacy &amp; Local Data
			</h3>
			<p className="mt-1 text-sm text-muted-foreground">
				Understand what Dev Context keeps locally and what portable exports
				omit.
			</p>
			<div className="mt-4 grid gap-4 sm:grid-cols-2">
				{privacyStatements.map((statement) => (
					<div key={statement.title} className="rounded-xl bg-muted/30 p-4">
						<h4 className="text-sm font-medium">{statement.title}</h4>
						<p className="mt-1 text-sm text-muted-foreground">
							{statement.description}
						</p>
					</div>
				))}
			</div>
		</section>
	);
}

export { PrivacySettings };
