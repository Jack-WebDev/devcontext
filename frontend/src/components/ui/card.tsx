import * as React from "react";

import { cn } from "../../lib/utils.js";

type CardElement = "article" | "div" | "label" | "p" | "section";
/**
 * Surface purpose: primary is a standalone decision, inset groups related
 * content, secondary supports a primary surface, tertiary stays in page flow,
 * and selection is an interactive choice.
 */
type CardHierarchy =
	| "primary"
	| "inset"
	| "secondary"
	| "tertiary"
	| "selection";

function Card({
	as: Component = "div",
	className,
	size = "default",
	hierarchy = "primary",
	...props
}: React.HTMLAttributes<HTMLElement> & {
	as?: CardElement;
	size?: "default" | "sm";
	hierarchy?: CardHierarchy;
}) {
	return React.createElement(Component, {
		...props,
		"data-slot": "card",
		"data-size": size,
		"data-hierarchy": hierarchy,
		className: cn(
			"group/card flex flex-col gap-4 overflow-hidden rounded-xl py-[var(--card-spacing)] text-sm text-card-foreground [--card-spacing:var(--layout-card-padding)] has-[>img:first-child]:pt-0 data-[size=sm]:[--card-spacing:1rem] *:[img:first-child]:rounded-none *:[img:last-child]:rounded-none",
			cardHierarchyClassName(hierarchy),
			className,
		),
	});
}

function cardHierarchyClassName(hierarchy: CardHierarchy): string {
	switch (hierarchy) {
		case "primary":
			return "border border-border/70 bg-card shadow-sm";
		case "inset":
			return "border border-border bg-[var(--surface-subtle)] shadow-none";
		case "secondary":
			return "border border-border bg-surface-muted shadow-none";
		case "tertiary":
			return "bg-transparent shadow-none";
		case "selection":
			return "border border-border bg-card shadow-sm transition-[border-color,box-shadow,background-color] duration-150 hover:border-foreground/20 data-[selected=true]:border-primary data-[selected=true]:ring-2 data-[selected=true]:ring-ring/40";
	}
}

function CardHeader({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-header"
			className={cn(
				"group/card-header @container/card-header grid auto-rows-min items-start gap-1.5 rounded-none px-(--card-spacing) has-data-[slot=card-action]:grid-cols-[1fr_auto] has-data-[slot=card-description]:grid-rows-[auto_auto] [.border-b]:pb-(--card-spacing)",
				className,
			)}
			{...props}
		/>
	);
}

function CardTitle({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-title"
			className={cn("text-section-title", className)}
			{...props}
		/>
	);
}

function CardDescription({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-description"
			className={cn("text-body text-secondary", className)}
			{...props}
		/>
	);
}

function CardAction({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-action"
			className={cn(
				"col-start-2 row-span-2 row-start-1 self-start justify-self-end",
				className,
			)}
			{...props}
		/>
	);
}

function CardContent({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-content"
			className={cn("px-(--card-spacing)", className)}
			{...props}
		/>
	);
}

function CardFooter({ className, ...props }: React.ComponentProps<"div">) {
	return (
		<div
			data-slot="card-footer"
			className={cn(
				"flex items-center px-(--card-spacing) [.border-t]:pt-(--card-spacing)",
				className,
			)}
			{...props}
		/>
	);
}

export type { CardHierarchy };
export {
	Card,
	CardAction,
	CardContent,
	CardDescription,
	CardFooter,
	CardHeader,
	CardTitle,
	cardHierarchyClassName,
};
