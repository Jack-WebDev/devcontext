import type { PreflightLaunchProjectResult } from "../../lib/devctx-api";
import { Button } from "../ui/button.js";
import { LaunchVerificationProgress } from "./LaunchVerificationProgress.js";

interface PreflightReviewViewProps {
	projectName: string;
	contextName: string;
	preflight: PreflightLaunchProjectResult;
	onFixFirst: () => void;
	onLaunchWithoutIt?: () => void;
}

function PreflightReviewView({
	projectName,
	contextName,
	preflight,
	onFixFirst,
	onLaunchWithoutIt,
}: PreflightReviewViewProps) {
	const blocking = preflight.groups.some((group) => group.blocking);
	const description = blocking
		? `${projectName} cannot launch as ${contextName} until the required checks are fixed.`
		: `${projectName} will open as ${contextName}. Review the checks or continue intentionally.`;

	return (
		<section className="space-y-3" aria-label="Preflight review">
			<LaunchVerificationProgress
				projectName={projectName}
				contextName={contextName}
				heading={blocking ? "Launch is blocked" : "Ready to launch"}
				description={description}
				groups={preflight.groups}
			/>
			<div className="flex flex-wrap justify-end gap-3">
				<Button type="button" variant="outline" onClick={onFixFirst}>
					Fix first
				</Button>
				{onLaunchWithoutIt ? (
					<Button type="button" onClick={onLaunchWithoutIt}>
						Launch without it
					</Button>
				) : null}
			</div>
		</section>
	);
}

export { PreflightReviewView };
