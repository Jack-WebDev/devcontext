import type { DisplayError } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { Disclosure } from "../ui/disclosure.js";
import { GuiErrorNotice } from "./GuiErrorNotice.js";

interface LaunchFailureViewProps {
	error: DisplayError;
	onRetry: () => void;
	onCancel: () => void;
	onRunDiagnostics?: () => void;
	onChooseAnotherContext: () => void;
}

function LaunchFailureView({
	error,
	onRetry,
	onCancel,
	onRunDiagnostics,
	onChooseAnotherContext,
}: LaunchFailureViewProps) {
	return (
		<section
			className="space-y-4"
			aria-labelledby="launch-failure-actions-title"
		>
			<GuiErrorNotice error={error} />
			<div>
				<h3 id="launch-failure-actions-title" className="text-sm font-medium">
					Next steps
				</h3>
				<p className="mt-1 text-sm text-muted-foreground">
					Dev Context is still open so you can resolve the issue or try the
					launch again.
				</p>
			</div>
			<div className="flex flex-wrap gap-3">
				<Button type="button" onClick={onRetry}>
					Retry
				</Button>
				{onRunDiagnostics ? (
					<Button type="button" variant="outline" onClick={onRunDiagnostics}>
						Run diagnostics
					</Button>
				) : null}
				<Button type="button" variant="outline" onClick={onChooseAnotherContext}>
					Choose another context
				</Button>
				<Button type="button" variant="ghost" onClick={onCancel}>
					Cancel
				</Button>
			</div>
			{error.launchFailureDetails || error.technicalDetails ? (
				<Disclosure summary="Technical details">
					{error.launchFailureDetails ? (
						<dl className="mt-3 grid gap-x-4 gap-y-2 text-xs sm:grid-cols-[auto_1fr]">
							<dt className="font-medium">Executable</dt>
							<dd className="wrap-break-word text-muted-foreground">
								{error.launchFailureDetails.executable}
							</dd>
							{error.launchFailureDetails.exitCode !== undefined ? (
								<>
									<dt className="font-medium">Exit code</dt>
									<dd className="text-muted-foreground">
										{error.launchFailureDetails.exitCode}
									</dd>
								</>
							) : null}
							<dt className="font-medium">Timestamp</dt>
							<dd className="text-muted-foreground">
								{error.launchFailureDetails.timestamp}
							</dd>
						</dl>
					) : null}
					<pre className="mt-3 overflow-x-auto whitespace-pre-wrap wrap-break-word bg-muted/40 p-3 text-xs text-muted-foreground">
						{error.launchFailureDetails?.logs ?? error.technicalDetails}
					</pre>
				</Disclosure>
			) : null}
		</section>
	);
}

export type { LaunchFailureViewProps };
export { LaunchFailureView };
