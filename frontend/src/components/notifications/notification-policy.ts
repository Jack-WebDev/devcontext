type NotificationKind = "provider_verified" | "provider_attention" | "tool_launched" | "update_available";

interface ProviderVerifiedNotification {
  kind: "provider_verified";
  providerName: string;
  contextName: string;
}

interface ProviderAttentionNotification {
  kind: "provider_attention";
  providerName: string;
  contextName: string;
  message: string;
}

interface ToolLaunchedNotification {
  kind: "tool_launched";
  projectName: string;
  contextName: string;
  toolName: string;
}

interface UpdateAvailableNotification {
  kind: "update_available";
  version: string;
}

type AppNotification =
  | ProviderVerifiedNotification
  | ProviderAttentionNotification
  | ToolLaunchedNotification
  | UpdateAvailableNotification;

interface NotificationPresentation {
  kind: NotificationKind;
  title: string;
  description: string;
  severity: "success" | "warning" | "info";
}

// Keep this list deliberately small. Notifications are reserved for changes
// that require acknowledgement outside the current screen, not routine work
// such as preflight checks, refreshes, or saved preferences.
function notificationPresentation(notification: AppNotification): NotificationPresentation {
  switch (notification.kind) {
    case "provider_verified":
      return {
        kind: notification.kind,
        title: `${notification.providerName} verified`,
        description: `${notification.providerName} is ready in ${notification.contextName}.`,
        severity: "success",
      };
    case "provider_attention":
      return {
        kind: notification.kind,
        title: `${notification.providerName} needs attention`,
        description: notification.message,
        severity: "warning",
      };
    case "tool_launched":
      return {
        kind: notification.kind,
        title: `${notification.toolName} launched`,
        description: `${notification.projectName} opened in ${notification.contextName}.`,
        severity: "success",
      };
    case "update_available":
      return {
        kind: notification.kind,
        title: "Update available",
        description: `Dev Context ${notification.version} is ready to install.`,
        severity: "info",
      };
  }
}

export { notificationPresentation };
export type { AppNotification, NotificationKind, NotificationPresentation };
