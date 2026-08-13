import type { DisplayError } from "../../lib/devctx-api";

interface GuiErrorNoticeProps {
  error: DisplayError;
}

function GuiErrorNotice({ error }: GuiErrorNoticeProps) {
  return (
    <section className="border border-destructive/30 bg-destructive/5 p-4 text-sm text-destructive" role="alert">
      <p className="text-xs font-semibold uppercase tracking-wide">{errorTitle(error.code)}</p>
      <p className="mt-2 font-medium">{error.message}</p>
      <p className="mt-1 text-destructive/80">{error.recovery}</p>
    </section>
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
