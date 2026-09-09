import { Input as InputPrimitive } from "@base-ui/react/input";
import type * as React from "react";

import { cn } from "../../lib/utils.js";

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
	return (
		<InputPrimitive
			type={type}
			data-slot="input"
			className={cn(
				"h-[var(--control-height)] w-full min-w-0 rounded-[10px] border border-input bg-card/85 px-3 py-1 text-base shadow-xs transition-[color,border-color,background-color,box-shadow] outline-none file:inline-flex file:h-7 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground/80 hover:border-foreground/20 hover:bg-card focus-visible:border-ring focus-visible:bg-card focus-visible:ring-2 focus-visible:ring-ring/25 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-muted/50 disabled:opacity-55 aria-busy:cursor-progress aria-invalid:border-destructive aria-invalid:ring-2 aria-invalid:ring-destructive/20 md:text-sm dark:aria-invalid:border-destructive/50",
				className,
			)}
			{...props}
		/>
	);
}

export { Input };
