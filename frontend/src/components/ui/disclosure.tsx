import type * as React from "react";

import { cn } from "../../lib/utils.js";

function Disclosure({
	summary,
	className,
	children,
	...props
}: React.ComponentProps<"details"> & { summary: React.ReactNode }) {
	return (
		<details
			data-slot="disclosure"
			className={cn(
				"group/disclosure rounded-lg border border-border/70 bg-(--surface-subtle) px-3 py-2.5 text-sm open:bg-card",
				className,
			)}
			{...props}
		>
			<summary className="cursor-pointer list-none font-medium text-foreground [&::-webkit-details-marker]:hidden">
				{summary}
			</summary>
			<div className="pt-3">{children}</div>
		</details>
	);
}

export { Disclosure };
