import type { DisplayError } from "../../lib/devctx-api";
import { Alert, AlertDescription, AlertTitle } from "../ui/alert.js";

interface GuiErrorNoticeProps {
  error: DisplayError;
}

function GuiErrorNotice({ error }: GuiErrorNoticeProps) {
  return (
    <Alert variant="destructive" className="border-destructive/30">
      <AlertTitle className="text-xs uppercase tracking-wide">{errorTitle(error.code)}</AlertTitle>
      <AlertDescription className="space-y-3">
        <ErrorSection label="What happened">{error.message}</ErrorSection>
        <ErrorSection label="Why it matters">{errorImpact(error.code)}</ErrorSection>
        <ErrorSection label="What to do">{error.recovery}</ErrorSection>
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

function ErrorSection({ label, children }: { label: string; children: string }) {
  return (
    <div>
      <p className="text-xs font-semibold tracking-wide uppercase">{label}</p>
      <p className="mt-1">{children}</p>
    </div>
  );
}

function errorImpact(code: DisplayError["code"]): string {
  switch (code) {
    case "context_mismatch_requires_confirmation":
      return "Launching with a different context could use the wrong development identity for this project.";
    case "validation_error":
      return "Dev Context needs a valid project and context before it can prepare an isolated launch.";
    case "launch_error":
      return "The selected coding tool did not start, so this project was not opened in the requested context.";
    case "canceled":
      return "The requested action did not run.";
    case "internal_error":
    case "unexpected_error":
      return "Dev Context could not complete the requested action safely.";
  }
}

export { GuiErrorNotice, errorImpact, errorTitle };
