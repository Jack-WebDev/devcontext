type AppRoute = "home" | "contexts" | "projects" | "running" | "history" | "settings";

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

function appRouteFromHash(hash: string): AppRoute {
  const route = hash.replace(/^#/, "");
  return appRoutes.some((definition) => definition.id === route) ? route as AppRoute : "home";
}

function appRouteDefinition(route: AppRoute): AppRouteDefinition {
  return appRoutes.find((definition) => definition.id === route) ?? appRoutes[0];
}

export { appRouteDefinition, appRouteFromHash, appRoutes };
export type { AppRoute, AppRouteDefinition };
