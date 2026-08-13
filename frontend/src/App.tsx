import { useEffect, useState } from "react";

import { GuiErrorNotice } from "./components/selector/GuiErrorNotice";
import { SelectorView } from "./components/selector/SelectorView";
import { createOnboardingContextAndRefresh } from "./components/selector/onboarding-action";
import {
  devContextApi,
  type ApiResult,
  type CreateContextResult,
  type DisplayError,
  type LaunchState,
} from "./lib/devctx-api";
import { devContextWindow } from "./lib/devctx-window";

type LaunchStateLoad =
  | { status: "loading" }
  | { status: "loaded"; data: LaunchState }
  | { status: "error"; error: DisplayError };

function App() {
  const [launchState, setLaunchState] = useState<LaunchStateLoad>({ status: "loading" });

  useEffect(() => {
    let active = true;

    devContextApi.getLaunchState().then((result) => {
      if (!active) {
        return;
      }

      if (result.ok) {
        setLaunchState({ status: "loaded", data: result.data });
        return;
      }

      setLaunchState({ status: "error", error: result.error });
    });

    return () => {
      active = false;
    };
  }, []);

  async function handleCreateContext(contextId: string): Promise<ApiResult<CreateContextResult>> {
    const result = await createOnboardingContextAndRefresh({
      contextId,
      createContext: (requestedContextId) => devContextApi.createContext({ contextId: requestedContextId }),
      getLaunchState: () => devContextApi.getLaunchState(),
    });
    if (result.ok) {
      setLaunchState({ status: "loaded", data: result.launchState });
      return { ok: true, data: result.created };
    }

    return { ok: false, error: result.error };
  }

  return (
    <main className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-background">
        <div className="mx-auto flex h-16 max-w-5xl items-center px-6">
          <h1 className="text-base font-semibold">Dev Context</h1>
        </div>
      </header>

      <div className="mx-auto max-w-5xl px-6 py-8">
        <section aria-labelledby="context-selector-heading" className="space-y-6">
          <div>
            <h2 id="context-selector-heading" className="text-2xl font-semibold">
              Context selector
            </h2>
          </div>

          {renderSelectorContent(launchState, handleCreateContext)}
        </section>
      </div>
    </main>
  );
}

function renderSelectorContent(
  launchState: LaunchStateLoad,
  onCreateContext: (contextId: string) => Promise<ApiResult<CreateContextResult>>,
) {
  if (launchState.status === "loading") {
    return <p className="text-sm text-muted-foreground">Loading selector...</p>;
  }

  if (launchState.status === "error") {
    return <GuiErrorNotice error={launchState.error} />;
  }

  return (
    <SelectorView
      launchState={launchState.data}
      onBindProject={(request) => devContextApi.bindProject(request)}
      onLaunchProject={(request) => devContextApi.launchProject(request)}
      onCancel={() => devContextWindow.closeSelector()}
      onCreatePersonalContext={() => onCreateContext("personal")}
      onCreateCompanyContext={() => onCreateContext("company")}
    />
  );
}

export default App;
