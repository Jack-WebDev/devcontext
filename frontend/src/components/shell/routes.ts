type AppRoute =
	| "home"
	| "contexts"
	| "projects"
	| "running"
	| "history"
	| "settings"
	| "trust"
	| "diagnostics";

interface AppRouteDefinition {
	id: AppRoute;
	label: string;
}

type ContextDetailDestination =
	| "overview"
	| "name-purpose"
	| "appearance"
	| "linked-projects"
	| "development-tools"
	| "launch-preferences";

interface ContextDetailRoute {
	contextId: string;
	destination: ContextDetailDestination;
}

const appRoutes: AppRouteDefinition[] = [
	{ id: "home", label: "Home" },
	{ id: "contexts", label: "Contexts" },
	{ id: "projects", label: "Projects" },
	{ id: "running", label: "Running" },
	{ id: "history", label: "History" },
	{ id: "settings", label: "Settings" },
	{ id: "trust", label: "Trust Center" },
];

const appRouteDefinitions: AppRouteDefinition[] = [
	...appRoutes,
	{ id: "diagnostics", label: "Diagnostics" },
];

function appRouteFromHash(hash: string): AppRoute {
	const route = hash.replace(/^#/, "");
	if (contextDetailRouteFromHash(hash)) return "contexts";
	return appRouteDefinitions.some((definition) => definition.id === route)
		? (route as AppRoute)
		: "home";
}

function contextDetailRouteFromHash(
	hash: string,
): ContextDetailRoute | undefined {
	const parts = hash.replace(/^#/, "").split("/");
	if (parts[0] !== "contexts" || !parts[1]) return undefined;
	const destination = parts[2] || "overview";
	if (
		destination !== "overview" &&
		destination !== "name-purpose" &&
		destination !== "appearance" &&
		destination !== "linked-projects" &&
		destination !== "development-tools" &&
		destination !== "launch-preferences"
	) {
		return undefined;
	}
	try {
		return { contextId: decodeURIComponent(parts[1]), destination };
	} catch {
		return undefined;
	}
}

function contextDetailHash(route: ContextDetailRoute): string {
	return `contexts/${encodeURIComponent(route.contextId)}/${route.destination}`;
}

function appRouteDefinition(route: AppRoute): AppRouteDefinition {
	return (
		appRouteDefinitions.find((definition) => definition.id === route) ??
		appRoutes[0]
	);
}

export type {
	AppRoute,
	AppRouteDefinition,
	ContextDetailDestination,
	ContextDetailRoute,
};
export {
	appRouteDefinition,
	appRouteDefinitions,
	appRouteFromHash,
	appRoutes,
	contextDetailHash,
	contextDetailRouteFromHash,
};
