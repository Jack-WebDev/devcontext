type LaunchSuccessCloseBehavior = "keep_open" | "close_selector";

const defaultLaunchSuccessCloseBehavior: LaunchSuccessCloseBehavior =
	"close_selector";

function shouldCloseSelectorAfterLaunch(
	behavior: LaunchSuccessCloseBehavior,
): boolean {
	return behavior === "close_selector";
}

export type { LaunchSuccessCloseBehavior };
export { defaultLaunchSuccessCloseBehavior, shouldCloseSelectorAfterLaunch };
