type AppRoute = "home" | "contexts" | "projects" | "running" | "history" | "settings" | "diagnostics";

interface AppRouteDefinition {
  id: AppRoute;
  label: string;
}

const appRoutes: AppRouteDefinition[] = [
  {id: "home", label: "Home"},
  {id: "contexts", label: "Contexts"},
  {id: "projects", label: "Projects"},
  {id: "running", label: "Running"},
  {id: "history", label: "History"},
  {id: "settings", label: "Settings"},
];

const appRouteDefinitions: AppRouteDefinition[] = [
  ...appRoutes,
  {id: "diagnostics", label: "Diagnostics"},
];

function appRouteFromHash(hash: string): AppRoute {
  const route = hash.replace(/^#/, "");
  return appRouteDefinitions.some((definition) => definition.id === route) ? route as AppRoute : "home";
}

function appRouteDefinition(route: AppRoute): AppRouteDefinition {
  return appRouteDefinitions.find((definition) => definition.id === route) ?? appRoutes[0];
}

export { appRouteDefinition, appRouteDefinitions, appRouteFromHash, appRoutes };
export type { AppRoute, AppRouteDefinition };
