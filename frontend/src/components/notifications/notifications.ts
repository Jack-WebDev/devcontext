import { toast } from "sonner";

import {
	type AppNotification,
	isToastEligible,
	notificationPresentation,
} from "./notification-policy";

function notify(notification: AppNotification) {
	if (!isToastEligible(notification)) {
		return;
	}
	const presentation = notificationPresentation(notification);
	toast[presentation.severity](presentation.title, {
		description: presentation.description,
	});
}

function notifyCodingToolLaunched({
	projectName,
	contextName,
	toolName,
}: {
	projectName: string;
	contextName: string;
	toolName: string;
}) {
	notify({ kind: "tool_launched", projectName, contextName, toolName });
}

export { notify, notifyCodingToolLaunched };
