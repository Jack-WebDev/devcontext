import { ShieldCheck } from "lucide-react";
import type { LaunchState } from "../../lib/devctx-api";

function AppStatusBar({
	launchState,
	onOpenSystemHealth,
}: {
	launchState?: LaunchState;
	onOpenSystemHealth: () => void;
}) {
	const setupRequired = launchState?.firstRun === true;
	const checkingIsolation =
		launchState === undefined ||
		(!setupRequired && launchState.confidence === undefined);
	const needsAttention =
		!checkingIsolation &&
		!setupRequired &&
		(launchState?.confidence?.status !== "ready" ||
			(launchState?.warnings.length ?? 0) > 0);
	const isolation = checkingIsolation
		? "Checking isolation"
		: setupRequired
			? "Isolation will be checked after setup"
			: launchState?.confidence?.status !== "ready"
				? "Isolation needs attention"
				: "Isolation ready";
	return (
		<footer className="h-12 border-t border-border/60 bg-[var(--sidebar-background)]">
			<div className="flex h-full items-center text-[11px] text-muted-foreground">
				<div className="flex h-full w-[var(--layout-sidebar-width)] items-center gap-2 border-r border-border/60 px-5">
					<ShieldCheck className="size-4" aria-hidden="true" />
					{isolation}
				</div>
				<button
					type="button"
					className="flex h-full items-center gap-2 px-5 transition-colors hover:bg-muted/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset"
					onClick={onOpenSystemHealth}
				>
					<span
						className={`size-2 rounded-full ${checkingIsolation ? "bg-muted-foreground" : setupRequired || needsAttention ? "bg-warning" : "bg-success"}`}
						aria-hidden="true"
					/>
					{setupRequired
						? "Setup required"
						: needsAttention
							? "System needs attention"
							: checkingIsolation
								? "Checking system status"
								: "All systems operational"}
				</button>
			</div>
		</footer>
	);
}

export { AppStatusBar };
