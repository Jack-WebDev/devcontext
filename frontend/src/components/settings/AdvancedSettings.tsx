import { Button } from "../ui/button.js";

function AdvancedSettings({ onOpenDiagnostics }: { onOpenDiagnostics: () => void }) {
	return (
		<section className="border-b border-border pb-6" aria-labelledby="settings-advanced">
			<h3 id="settings-advanced" className="font-semibold">Advanced</h3>
			<p className="mt-1 text-sm text-muted-foreground">
				Open technical diagnostics when you need to investigate local configuration, storage, tools, or integrations.
			</p>
			<div className="mt-4 flex items-center justify-between gap-6">
				<span className="text-sm text-muted-foreground">CLI version: <code>devctx --version</code></span>
				<Button type="button" variant="outline" onClick={onOpenDiagnostics}>
					Open diagnostics
				</Button>
			</div>
		</section>
	);
}

export { AdvancedSettings };
