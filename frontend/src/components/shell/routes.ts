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
	| "launch-preferences"
	| "environment"
	| "activity"
	| "advanced";

interface ContextDetailRoute {
	contextId: string;
	destination: ContextDetailDestination;
}

interface ProjectDetailRoute {
	projectPath: string;
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
	if (projectDetailRouteFromHash(hash)) return "projects";
	return appRouteDefinitions.some((definition) => definition.id === route)
		? (route as AppRoute)
		: "home";
}

function projectDetailRouteFromHash(
	hash: string,
): ProjectDetailRoute | undefined {
	const parts = hash.replace(/^#/, "").split("/");
	if (parts[0] !== "projects" || !parts[1] || parts[2]) return undefined;
	try {
		return { projectPath: decodeURIComponent(parts[1]) };
	} catch {
		return undefined;
	}
}

function projectDetailHash(route: ProjectDetailRoute): string {
	return `projects/${encodeURIComponent(route.projectPath)}`;
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
		destination !== "launch-preferences" &&
		destination !== "environment" &&
		destination !== "activity" &&
		destination !== "advanced"
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
	ProjectDetailRoute,
};
export {
	appRouteDefinition,
	appRouteDefinitions,
	appRouteFromHash,
	appRoutes,
	contextDetailHash,
	contextDetailRouteFromHash,
	projectDetailHash,
	projectDetailRouteFromHash,
};
