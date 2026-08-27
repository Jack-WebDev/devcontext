type LaunchSuccessCloseBehavior = "keep_open" | "close_selector";

const defaultLaunchSuccessCloseBehavior: LaunchSuccessCloseBehavior = "keep_open";

function shouldCloseSelectorAfterLaunch(behavior: LaunchSuccessCloseBehavior): boolean {
  return behavior === "close_selector";
}

export {
  defaultLaunchSuccessCloseBehavior,
  shouldCloseSelectorAfterLaunch,
};
export type { LaunchSuccessCloseBehavior };
