import { ShieldCheck } from "lucide-react";
import type { LaunchState } from "../../lib/devctx-api";

function AppStatusBar({ launchState }: { launchState?: LaunchState }) {
	const checkingIsolation = launchState?.confidence === undefined;
	const needsAttention =
		!checkingIsolation &&
		(launchState?.confidence?.status !== "ready" ||
			(launchState?.warnings.length ?? 0) > 0);
	const isolation =
		checkingIsolation
			? "Checking isolation"
			: launchState?.confidence?.status !== "ready"
				? "Isolation needs attention"
				: "Isolation ready";
	return (
		<footer className="h-12 border-t border-border bg-[#faf9f7]">
			<div className="flex h-full items-center text-[11px] text-muted-foreground">
				<div className="flex h-full w-59.5 items-center gap-2 border-r border-border px-5">
					<ShieldCheck className="size-4" />
					{isolation}
				</div>
				<div className="flex items-center gap-2 px-5">
					<span
						className={`size-2 rounded-full ${checkingIsolation ? "bg-muted-foreground" : needsAttention ? "bg-warning" : "bg-success"}`}
					/>
					{needsAttention
						? "System needs attention"
						: checkingIsolation
							? "Checking system status"
							: "All systems operational"}
				</div>
			</div>
		</footer>
	);
}

export { AppStatusBar };
