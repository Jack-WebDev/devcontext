import type * as React from "react";

import { cn } from "../../lib/utils.js";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
	return (
		<textarea
			data-slot="textarea"
			className={cn(
				"flex field-sizing-content min-h-24 w-full resize-none rounded-[10px] border border-input bg-card/85 px-3 py-2.5 text-base shadow-xs transition-[color,border-color,background-color,box-shadow] outline-none placeholder:text-muted-foreground/80 hover:border-foreground/20 hover:bg-card focus-visible:border-ring focus-visible:bg-card focus-visible:ring-2 focus-visible:ring-ring/25 disabled:cursor-not-allowed disabled:bg-muted/50 disabled:opacity-55 aria-busy:cursor-progress aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm dark:aria-invalid:border-destructive/50",
				className,
			)}
			{...props}
		/>
	);
}

export { Textarea };
