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
				"group/disclosure rounded-xl bg-muted/40 px-4 py-3 text-sm transition-colors open:bg-muted/55",
				className,
			)}
			{...props}
		>
			<summary className="cursor-pointer list-none font-semibold text-foreground transition-colors hover:text-primary [&::-webkit-details-marker]:hidden">
				{summary}
			</summary>
			<div className="pt-3">{children}</div>
		</details>
	);
}

export { Disclosure };
