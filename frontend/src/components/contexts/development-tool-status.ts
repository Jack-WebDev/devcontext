import type { DevelopmentToolStatus } from "../../lib/devctx-api.js";

interface DevelopmentToolStatusPresentation {
	label: string;
	message: string;
	recovery: string;
}

const developmentToolStatusPresentation: Record<
	DevelopmentToolStatus,
	DevelopmentToolStatusPresentation
> = {
	available: {
		label: "Available",
		message: "This development tool is available for this context.",
		recovery: "Add it to this context when you are ready.",
	},
	connected: {
		label: "Connected",
		message: "This development tool is connected to this context.",
		recovery: "No action is needed.",
	},
	needs_sign_in: {
		label: "Needs sign-in",
		message: "This development tool needs sign-in verification.",
		recovery: "Sign in, then return to verify the connection.",
	},
	not_configured: {
		label: "Not configured",
		message: "This development tool has not been configured for this context.",
		recovery: "Open its setup and complete the required configuration.",
	},
	not_found: {
		label: "Not found",
		message: "This development tool could not be found on this computer.",
		recovery: "Install it or select another development tool.",
	},
	unavailable: {
		label: "Unavailable",
		message: "Its current availability could not be confirmed.",
		recovery: "Check that it is installed and available, then try again.",
	},
	error: {
		label: "Error",
		message: "Dev Context could not read this development tool's status.",
		recovery: "Try again or inspect its configuration.",
	},
};

export { developmentToolStatusPresentation };
