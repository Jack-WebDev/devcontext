type NotificationKind =
	| "provider_identity_observed"
	| "provider_attention"
	| "tool_launched"
	| "update_available";

interface ProviderIdentityObservedNotification {
	kind: "provider_identity_observed";
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
	| ProviderIdentityObservedNotification
	| ProviderAttentionNotification
	| ToolLaunchedNotification
	| UpdateAvailableNotification;

interface NotificationPresentation {
	kind: NotificationKind;
	title: string;
	description: string;
	severity: "success" | "warning" | "info";
}

function isToastEligible(notification: AppNotification): boolean {
	// A warning that requires a repair belongs on the affected screen, where its
	// recovery action remains available. Toasts only confirm non-critical events.
	return notification.kind !== "provider_attention";
}

// Keep this list deliberately small. Notifications are reserved for changes
// that require acknowledgement outside the current screen, not routine work
// such as preflight checks, refreshes, or saved preferences.
function notificationPresentation(
	notification: AppNotification,
): NotificationPresentation {
	switch (notification.kind) {
		case "provider_identity_observed":
			return {
				kind: notification.kind,
				title: `${notification.providerName} identity observed`,
				description: `${notification.providerName} account metadata was found locally in ${notification.contextName}.`,
				severity: "info",
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
				title: `${notification.toolName} started`,
				description: `Started for ${notification.projectName} in ${notification.contextName}.`,
				severity: "info",
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

export type { AppNotification, NotificationKind, NotificationPresentation };
export { isToastEligible, notificationPresentation };
