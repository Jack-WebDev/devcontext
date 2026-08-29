import type { ReactNode } from "react";

type Status =
	| "ready"
	| "needs_attention"
	| "not_configured"
	| "blocked"
	| "running"
	| "protected"
	| "failed";

interface StatusIndicatorProps {
	status: Status;
	label?: string;
	children?: ReactNode;
}

function StatusIndicator({ status, label, children }: StatusIndicatorProps) {
	const presentation = statusPresentation(status);
	const text = children ?? label ?? presentation.label;

	return (
		<span
			className={`inline-flex items-center gap-1.5 text-xs font-medium ${presentation.className}`}
		>
			<span
				className={`size-2 rounded-full ${presentation.indicatorClassName}`}
				aria-hidden="true"
			/>
			<span>{text}</span>
		</span>
	);
}

function statusPresentation(status: Status): {
	label: string;
	className: string;
	indicatorClassName: string;
} {
	switch (status) {
		case "ready":
			return {
				label: "Ready",
				className: "text-success",
				indicatorClassName: "bg-success",
			};
		case "needs_attention":
			return {
				label: "Needs attention",
				className: "text-warning",
				indicatorClassName: "bg-warning",
			};
		case "not_configured":
			return {
				label: "Not configured",
				className: "text-warning",
				indicatorClassName: "bg-warning",
			};
		case "blocked":
		case "failed":
			return {
				label: status === "blocked" ? "Blocked" : "Failed",
				className: "text-destructive",
				indicatorClassName: "bg-destructive",
			};
		case "running":
			return {
				label: "Running",
				className: "text-accent-company",
				indicatorClassName: "bg-accent-company",
			};
		case "protected":
			return {
				label: "Protected",
				className: "text-success",
				indicatorClassName: "bg-success",
			};
	}
}

export type { Status, StatusIndicatorProps };
export { StatusIndicator, statusPresentation };
