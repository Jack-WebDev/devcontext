import type { DisplayError } from "../../lib/devctx-api";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert.js";

interface GuiErrorNoticeProps {
  error: DisplayError;
}

function GuiErrorNotice({ error }: GuiErrorNoticeProps) {
  return (
    <Alert variant="destructive" className="border-destructive/30">
      <AlertTitle className="text-xs uppercase tracking-wide">{errorTitle(error.code)}</AlertTitle>
      <AlertDescription>
        <p className="font-medium">{error.message}</p>
        <p className="mt-1">{error.recovery}</p>
      </AlertDescription>
    </Alert>
  );
}

function errorTitle(code: DisplayError["code"]): string {
  switch (code) {
    case "validation_error":
      return "Check the project or context";
    case "launch_error":
      return "Launch failed";
    case "context_mismatch_requires_confirmation":
      return "Context mismatch";
    case "canceled":
      return "Canceled";
    case "internal_error":
      return "Dev Context failed";
    case "unexpected_error":
      return "Unexpected error";
  }
}

export { GuiErrorNotice, errorTitle };
