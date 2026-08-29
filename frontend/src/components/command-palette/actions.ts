import type { ContextState } from "../../lib/devctx-api";
import type { AppRouteDefinition } from "../shell/routes";

interface CommandPaletteAction {
	id: string;
	label: string;
	keywords?: string[];
	disabled?: boolean;
	onSelect: () => void;
}

function launchContextActions(
	contexts: ContextState[],
	onLaunch: (contextId: string) => void,
): CommandPaletteAction[] {
	return contexts.map((context) => ({
		id: `launch-context-${context.id}`,
		label: `Launch ${context.name}`,
		keywords: [context.name, context.tool.name],
		disabled: context.confidence?.status === "blocked",
		onSelect: () => onLaunch(context.id),
	}));
}

function navigationActions(
	routes: AppRouteDefinition[],
	onNavigate: (route: AppRouteDefinition["id"]) => void,
): CommandPaletteAction[] {
	return routes.map((route) => ({
		id: `navigate-${route.id}`,
		label: `Open ${route.label}`,
		keywords: [route.id],
		onSelect: () => onNavigate(route.id),
	}));
}

export type { CommandPaletteAction };
export { launchContextActions, navigationActions };
