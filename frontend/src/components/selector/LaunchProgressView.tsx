import type {
	LaunchVerificationStep,
	PreflightGroup,
} from "../../lib/devctx-api";
import { LaunchVerificationProgress } from "./LaunchVerificationProgress.js";

interface LaunchProgressViewProps {
	projectName: string;
	contextName: string;
	groups?: PreflightGroup[];
	steps?: LaunchVerificationStep[];
	showVerification: boolean;
}

// This intentionally contains no selection or recovery controls. While a
// launch request is in flight, changing context or canceling would create an
// ambiguous result for the process that is being started.
function LaunchProgressView({
	projectName,
	contextName,
	groups,
	steps,
	showVerification,
}: LaunchProgressViewProps) {
	return (
		<section className="space-y-4" aria-label="Launching project" aria-live="polite">
			<div>
				<p className="text-sm text-muted-foreground">Launching project</p>
				<h2 className="mt-1 text-xl font-semibold">Opening {projectName}</h2>
				<p className="mt-1 text-sm text-muted-foreground">
					Using {contextName}.
				</p>
			</div>
			{showVerification ? (
				<LaunchVerificationProgress
					projectName={projectName}
					contextName={contextName}
					heading="Launch progress"
					groups={groups}
					steps={steps}
				/>
			) : null}
		</section>
	);
}

export { LaunchProgressView };
