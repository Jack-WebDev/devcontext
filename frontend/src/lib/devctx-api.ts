export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: DisplayError };

export type ErrorCode =
  | "canceled"
  | "validation_error"
  | "context_mismatch_requires_confirmation"
  | "launch_error"
  | "internal_error"
  | "unexpected_error";

export interface DisplayError {
  code: ErrorCode;
  message: string;
  recovery: string;
  technicalDetails?: string;
  contextMismatch?: ContextMismatch;
}

export interface ContextMismatch {
  projectPath: string;
  boundContextId: string;
  requestedContextId: string;
}

export interface GetLaunchStateRequest {
  projectPath?: string;
}

export interface GetHomeDashboardRequest {
  projectPath?: string;
}

export interface HomeDashboardState {
  project: ProjectState;
  currentContext?: HomeCurrentContextState;
  recentProjects: RecentProjectState[];
  running: HomeRunningSummary;
  activity: HomeActivitySummary;
}

export interface HomeCurrentContextState {
  id: string;
  name: string;
  tool: ToolState;
  confidence: LaunchConfidenceState;
}

export interface RecentProjectsState {
  projects: RecentProjectState[];
}

export interface RecentProjectState {
  project: ProjectState;
  contextId: string;
  contextName?: string;
  lastLaunchedAt: string;
}

export interface ContextListState {
  contexts: ContextListItem[];
}

export interface ContextListItem {
  context: ContextState;
  enabledProviders: ProviderState[];
  projectCount: number;
  lastUsedAt?: string;
}

export interface GetContextDetailsRequest {
  contextId: string;
}

export interface ContextDetailsState {
  context: ContextState;
  location: string;
  createdAt: string;
  projectCount: number;
  lastUsedAt?: string;
  enabledProviders: ProviderState[];
}

export interface HomeRunningSummary {
  count: number;
}

export interface HomeActivitySummary {
  count: number;
}

export interface LaunchState {
  project: ProjectState;
  contexts: ContextState[];
  binding: ProjectBindingState;
  confidence?: LaunchConfidenceState;
  selectedContextId?: string;
  selectionRequired: boolean;
  resolutionSource?: string;
  warnings: ResolutionWarning[];
  firstRun: boolean;
  providerCredentialSessions: ProviderCredentialSession[];
}

export interface ProjectState {
  name: string;
  path: string;
}

export interface ContextState {
  id: string;
  name: string;
  description?: string;
  tool: ToolState;
  availableTools: ToolOption[];
  providers: ProviderState[];
  confidence?: LaunchConfidenceState;
  metadata?: Record<string, string>;
}

export interface ToolState {
  id: string;
  name: string;
  status: LaunchConfidenceStatus;
  message: string;
  actionHint?: string;
}

export interface ToolOption {
  id: string;
  name: string;
}

export interface ProviderState {
  id: string;
  name: string;
  enabled: boolean;
  state: ProviderReadinessState;
  explanation?: string;
  actionHint?: string;
  setupAction?: ProviderSetupAction;
  identity: ProviderIdentityState;
}

export type ProviderSetupState = "open_and_configure" | "waiting_for_sign_in" | "verified";

export interface ProviderSetupAction {
  state: ProviderSetupState;
  label: string;
  message: string;
}

export type ProviderReadinessState = "ready" | "not_configured" | "directory_missing" | "unavailable";

export type ProviderIdentityStatus = "verified" | "unavailable" | "none" | "mismatch_evidence";

export interface ProviderIdentityState {
  status: ProviderIdentityStatus;
  message?: string;
  fields: ProviderMetadataField[];
}

export interface ProviderMetadataField {
  label: string;
  value: string;
}

export type LaunchConfidenceStatus = "ready" | "needs_attention" | "blocked";

export type LaunchConfidenceCheckComponent = "provider" | "tool" | "isolation";

export interface LaunchConfidenceState {
  contextId: string;
  status: LaunchConfidenceStatus;
  checks: LaunchConfidenceCheck[];
}

export interface LaunchConfidenceCheck {
  component: LaunchConfidenceCheckComponent;
  providerId?: string;
  toolId?: string;
  severity: LaunchConfidenceStatus;
  label: string;
  message: string;
  actionHint?: string;
}

export interface ProjectBindingState {
  projectPath: string;
  bound: boolean;
  contextId?: string;
  dangling: boolean;
  missingContextId?: string;
  recovery?: string;
}

export interface ResolutionWarning {
  code: string;
  message: string;
  projectPath?: string;
  boundContextId?: string;
  requestedContextId?: string;
}

export interface LaunchProjectRequest {
  projectPath?: string;
  contextId: string;
  confirmContextMismatch?: boolean;
}

export interface PreflightLaunchProjectRequest {
  projectPath?: string;
  contextId: string;
  confirmContextMismatch?: boolean;
}

export interface PreflightLaunchProjectResult {
  project: ProjectState;
  context: ContextState;
  confidence: LaunchConfidenceState;
  verificationSteps?: LaunchVerificationStep[];
  warnings: ResolutionWarning[];
}

export type LaunchVerificationStepStatus = "pending" | LaunchConfidenceStatus;

export interface LaunchVerificationStep {
  id: string;
  label: string;
  status: LaunchVerificationStepStatus;
  message: string;
}

export interface LaunchProjectResult {
  project: ProjectState;
  context: ContextState;
  warnings: ResolutionWarning[];
}

export interface BindProjectRequest {
  projectPath?: string;
  contextId: string;
}

export interface UnbindProjectRequest {
  projectPath?: string;
}

export interface CreateContextRequest {
  contextId: string;
  name?: string;
  description?: string;
  icon?: string;
  accent?: string;
  enabledProviderIds?: string[];
  toolId?: string;
  importProviderIds?: string[];
}

export interface CreateContextResult {
  context: ContextState;
}
export interface ProjectsState { projects: ProjectListItem[]; }
export interface ProjectListItem { project: ProjectState; contextId?: string; contextName?: string; lastLaunchedAt?: string; running: boolean; }

export interface GetDiagnosticsRequest { contextId?: string; }
export interface DiagnosticsState { groups: DiagnosticGroup[]; }
export interface DiagnosticGroup { id: string; label: string; checks: DiagnosticCheck[]; }
export interface DiagnosticCheck {
  id: string;
  severity: "ready" | "needs_attention" | "blocked";
  label: string;
  message: string;
  details: DiagnosticDetail[];
}
export interface DiagnosticDetail { label: string; value: string; isPath: boolean; }

export interface GetRepairActionsRequest { contextId: string; }
export interface RepairActionsState { actions: RepairAction[]; }
export interface RepairAction {
  id: string;
  label: string;
  description: string;
  destructive: boolean;
  requiresConfirmation: boolean;
  targets: RepairTarget[];
}
export interface RepairTarget { label: string; path: string; kind: string; }
export interface RunRepairActionRequest { contextId: string; actionId: string; confirmDestructive?: boolean; }
export interface RunRepairActionResult { actionId: string; diagnostics: DiagnosticsState; }
export interface HistoryState { entries: HistoryEntry[]; }
export type HistoryCategory = "launch" | "configuration" | "warning";
export interface HistoryEntry { event: string; category: HistoryCategory; timestamp: string; projectPath?: string; contextId?: string; toolId?: string; message: string; }

export interface RunningEnvironmentsState { environments: RunningEnvironmentState[]; }
export interface RunningEnvironmentState {
  id: string;
  project: ProjectState;
  context: RunningEnvironmentContextState;
  tool: ToolOption;
  startedAt: string;
  process: RunningEnvironmentProcessState;
  session: RunningEnvironmentSessionState;
  launch: RunningEnvironmentLaunchState;
}
export interface RunningEnvironmentContextState { id: string; name: string; }
export interface RunningEnvironmentProcessState { state: string; pid?: number; }
export interface RunningEnvironmentSessionState { id?: string; state: string; }
export interface RunningEnvironmentLaunchState { source: string; resolutionSource: string; }

export interface ProviderCredentialSession {
  providerId: string;
  name: string;
  metadataAvailable: boolean;
  fields: ProviderMetadataField[];
}

export interface DevContextApi {
  getLaunchState(request?: GetLaunchStateRequest): Promise<ApiResult<LaunchState>>;
  getHomeDashboard(request?: GetHomeDashboardRequest): Promise<ApiResult<HomeDashboardState>>;
  getRecentProjects(): Promise<ApiResult<RecentProjectsState>>;
  getContexts(): Promise<ApiResult<ContextListState>>;
  getContextDetails(request: GetContextDetailsRequest): Promise<ApiResult<ContextDetailsState>>;
  preflightLaunchProject(request: PreflightLaunchProjectRequest): Promise<ApiResult<PreflightLaunchProjectResult>>;
  launchProject(request: LaunchProjectRequest): Promise<ApiResult<LaunchProjectResult>>;
  bindProject(request: BindProjectRequest): Promise<ApiResult<ProjectBindingState>>;
  unbindProject(request?: UnbindProjectRequest): Promise<ApiResult<ProjectBindingState>>;
  createContext(request: CreateContextRequest): Promise<ApiResult<CreateContextResult>>;
  getProjects(): Promise<ApiResult<ProjectsState>>;
  getDiagnostics(request?: GetDiagnosticsRequest): Promise<ApiResult<DiagnosticsState>>;
  getRepairActions(request: GetRepairActionsRequest): Promise<ApiResult<RepairActionsState>>;
  runRepairAction(request: RunRepairActionRequest): Promise<ApiResult<RunRepairActionResult>>;
  getHistory(): Promise<ApiResult<HistoryState>>;
  getRunningEnvironments(): Promise<ApiResult<RunningEnvironmentsState>>;
}

export interface WailsBindings {
  getLaunchState(request: GetLaunchStateRequest): Promise<unknown>;
  getHomeDashboard(request: GetHomeDashboardRequest): Promise<unknown>;
  getRecentProjects(): Promise<unknown>;
  getContexts(): Promise<unknown>;
  getContextDetails(request: GetContextDetailsRequest): Promise<unknown>;
  preflightLaunchProject(request: PreflightLaunchProjectRequest): Promise<unknown>;
  launchProject(request: LaunchProjectRequest): Promise<unknown>;
  bindProject(request: BindProjectRequest): Promise<unknown>;
  unbindProject(request: UnbindProjectRequest): Promise<unknown>;
  createContext(request: CreateContextRequest): Promise<unknown>;
  getProjects(): Promise<unknown>;
  getDiagnostics(request: GetDiagnosticsRequest): Promise<unknown>;
  getRepairActions(request: GetRepairActionsRequest): Promise<unknown>;
  runRepairAction(request: RunRepairActionRequest): Promise<unknown>;
  getHistory(): Promise<unknown>;
  getRunningEnvironments(): Promise<unknown>;
}

export function createDevContextApi(bindings: WailsBindings = generatedBindings): DevContextApi {
  return {
    getLaunchState(request = {}) {
      return callBinding(() => bindings.getLaunchState(request), normalizeLaunchState);
    },
    getHomeDashboard(request = {}) {
      return callBinding(() => bindings.getHomeDashboard(request), normalizeHomeDashboardState);
    },
    getRecentProjects() {
      return callBinding(() => bindings.getRecentProjects(), normalizeRecentProjectsState);
    },
    getContexts() {
      return callBinding(() => bindings.getContexts(), normalizeContextListState);
    },
    getContextDetails(request) {
      return callBinding(() => bindings.getContextDetails(request), normalizeContextDetailsState);
    },
    preflightLaunchProject(request) {
      return callBinding(
        () =>
          bindings.preflightLaunchProject({
            ...request,
            confirmContextMismatch: request.confirmContextMismatch ?? false,
          }),
        normalizePreflightLaunchProjectResult,
      );
    },
    launchProject(request) {
      return callBinding(
        () =>
          bindings.launchProject({
            ...request,
            confirmContextMismatch: request.confirmContextMismatch ?? false,
          }),
        normalizeLaunchProjectResult,
      );
    },
    bindProject(request) {
      return callBinding(() => bindings.bindProject(request), normalizeProjectBindingState);
    },
    unbindProject(request = {}) {
      return callBinding(() => bindings.unbindProject(request), normalizeProjectBindingState);
    },
    createContext(request) {
      return callBinding(() => bindings.createContext(request), normalizeCreateContextResult);
    },
    getProjects() { return callBinding(() => bindings.getProjects(), normalizeProjectsState); },
    getDiagnostics(request = {}) { return callBinding(() => bindings.getDiagnostics(request), normalizeDiagnosticsState); },
    getRepairActions(request) { return callBinding(() => bindings.getRepairActions(request), normalizeRepairActionsState); },
    runRepairAction(request) { return callBinding(() => bindings.runRepairAction({...request, confirmDestructive: request.confirmDestructive ?? false}), normalizeRunRepairActionResult); },
    getHistory() { return callBinding(() => bindings.getHistory(), normalizeHistoryState); },
    getRunningEnvironments() { return callBinding(() => bindings.getRunningEnvironments(), normalizeRunningEnvironmentsState); },
  };
}

const generatedBindings: WailsBindings = {
  async getLaunchState(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetLaunchState(request);
  },
  async getHomeDashboard(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetHomeDashboard(request);
  },
  async getRecentProjects() {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetRecentProjects();
  },
  async getContexts() {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetContexts();
  },
  async getContextDetails(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetContextDetails(request);
  },
  async preflightLaunchProject(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.PreflightLaunchProject({
      ...request,
      confirmContextMismatch: request.confirmContextMismatch ?? false,
    });
  },
  async launchProject(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.LaunchProject({
      ...request,
      confirmContextMismatch: request.confirmContextMismatch ?? false,
    });
  },
  async bindProject(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.BindProject(request);
  },
  async unbindProject(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.UnbindProject(request);
  },
  async createContext(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.CreateContext(request);
  },
  async getProjects() { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.GetProjects(); },
  async getDiagnostics(request) { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.GetDiagnostics(request); },
  async getRepairActions(request) { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.GetRepairActions(request); },
  async runRepairAction(request) { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.RunRepairAction({...request, confirmDestructive: request.confirmDestructive ?? false}); },
  async getHistory() { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.GetHistory(); },
  async getRunningEnvironments() { const bindings = await import("../../wailsjs/go/wailsapp/App"); return bindings.GetRunningEnvironments(); },
};

export const devContextApi = createDevContextApi();

async function callBinding<T>(
  operation: () => Promise<unknown>,
  normalize: (value: unknown) => T,
): Promise<ApiResult<T>> {
  try {
    const value = unwrapBindingValue(await operation());
    if (isApplicationError(value)) {
      return { ok: false, error: normalizeApplicationError(value) };
    }
    return { ok: true, data: normalize(value) };
  } catch (error) {
    return { ok: false, error: normalizeRejectedError(error) };
  }
}

function unwrapBindingValue(value: unknown): unknown {
  if (!Array.isArray(value) || value.length !== 2) {
    return value;
  }

  const [result, error] = value;
  if (isApplicationError(error)) {
    return error;
  }
  if (error === undefined || error === null) {
    return result;
  }
  return error;
}

function normalizeLaunchState(value: unknown): LaunchState {
  const object = objectValue(value);
  return {
    project: normalizeProjectState(object.project),
    contexts: arrayValue(object.contexts).map(normalizeContextState),
    binding: normalizeProjectBindingState(object.binding),
    confidence: normalizeLaunchConfidenceState(object.confidence),
    selectedContextId: optionalString(object.selectedContextId),
    selectionRequired: booleanValue(object.selectionRequired),
    resolutionSource: optionalString(object.resolutionSource),
    warnings: arrayValue(object.warnings).map(normalizeResolutionWarning),
    firstRun: booleanValue(object.firstRun),
    providerCredentialSessions: arrayValue(object.providerCredentialSessions).map(normalizeProviderCredentialSession),
  };
}

function normalizeHomeDashboardState(value: unknown): HomeDashboardState {
  const object = objectValue(value);
  const currentContext = normalizeHomeCurrentContextState(object.currentContext);
  return {
    project: normalizeProjectState(object.project),
    ...(currentContext === undefined ? {} : {currentContext}),
    recentProjects: arrayValue(object.recentProjects).map(normalizeRecentProjectState),
    running: normalizeHomeRunningSummary(object.running),
    activity: normalizeHomeActivitySummary(object.activity),
  };
}

function normalizeHomeCurrentContextState(value: unknown): HomeCurrentContextState | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }

  const object = objectValue(value);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    tool: normalizeToolState(object.tool),
    confidence: requiredLaunchConfidenceState(object.confidence),
  };
}

function normalizeRecentProjectsState(value: unknown): RecentProjectsState {
  return {projects: arrayValue(objectValue(value).projects).map(normalizeRecentProjectState)};
}

function normalizeRecentProjectState(value: unknown): RecentProjectState {
  const object = objectValue(value);
  const contextName = optionalString(object.contextName);
  return {
    project: normalizeProjectState(object.project),
    contextId: stringValue(object.contextId),
    ...(contextName === undefined ? {} : {contextName}),
    lastLaunchedAt: timestampValue(object.lastLaunchedAt),
  };
}

function normalizeContextListState(value: unknown): ContextListState {
  return {contexts: arrayValue(objectValue(value).contexts).map(normalizeContextListItem)};
}

function normalizeContextListItem(value: unknown): ContextListItem {
  const object = objectValue(value);
  const lastUsedAt = optionalTimestamp(object.lastUsedAt);
  return {
    context: normalizeContextState(object.context),
    enabledProviders: arrayValue(object.enabledProviders).map(normalizeProviderState),
    projectCount: numberValue(object.projectCount),
    ...(lastUsedAt === undefined ? {} : {lastUsedAt}),
  };
}

function normalizeContextDetailsState(value: unknown): ContextDetailsState {
  const object = objectValue(value);
  const lastUsedAt = optionalTimestamp(object.lastUsedAt);
  return {
    context: normalizeContextState(object.context),
    location: stringValue(object.location),
    createdAt: timestampValue(object.createdAt),
    projectCount: numberValue(object.projectCount),
    enabledProviders: arrayValue(object.enabledProviders).map(normalizeProviderState),
    ...(lastUsedAt === undefined ? {} : {lastUsedAt}),
  };
}

function normalizeHomeRunningSummary(value: unknown): HomeRunningSummary {
  return {count: numberValue(objectValue(value).count)};
}

function normalizeHomeActivitySummary(value: unknown): HomeActivitySummary {
  return {count: numberValue(objectValue(value).count)};
}

function normalizeLaunchConfidenceState(value: unknown): LaunchConfidenceState | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }

  const object = objectValue(value);
  return {
    contextId: stringValue(object.contextId),
    status: normalizeLaunchConfidenceStatus(object.status),
    checks: arrayValue(object.checks).map(normalizeLaunchConfidenceCheck),
  };
}

function normalizeLaunchConfidenceCheck(value: unknown): LaunchConfidenceCheck {
  const object = objectValue(value);
  const component = normalizeLaunchConfidenceCheckComponent(object.component);
  const providerId = optionalString(object.providerId);
  const toolId = optionalString(object.toolId);
  if (component === "provider" && providerId === undefined) {
	throw new Error("Invalid Dev Context response.");
  }
	if (component === "tool" && toolId === undefined) {
		throw new Error("Invalid Dev Context response.");
	}
  return {
    component,
    ...(providerId === undefined ? {} : {providerId}),
	...(toolId === undefined ? {} : {toolId}),
    severity: normalizeLaunchConfidenceStatus(object.severity),
    label: stringValue(object.label),
    message: stringValue(object.message),
    actionHint: optionalString(object.actionHint),
  };
}

function normalizeLaunchConfidenceStatus(value: unknown): LaunchConfidenceStatus {
  switch (value) {
    case "ready":
    case "needs_attention":
    case "blocked":
      return value;
    default:
      throw new Error("Invalid Dev Context response.");
  }
}

function normalizeLaunchConfidenceCheckComponent(value: unknown): LaunchConfidenceCheckComponent {
  switch (value) {
    case "provider":
	case "tool":
    case "isolation":
      return value;
    default:
      throw new Error("Invalid Dev Context response.");
  }
}

function normalizeLaunchProjectResult(value: unknown): LaunchProjectResult {
  const object = objectValue(value);
  return {
    project: normalizeProjectState(object.project),
    context: normalizeContextState(object.context),
    warnings: arrayValue(object.warnings).map(normalizeResolutionWarning),
  };
}

function normalizePreflightLaunchProjectResult(value: unknown): PreflightLaunchProjectResult {
  const object = objectValue(value);
  const verificationSteps = arrayValue(object.verificationSteps).map(normalizeLaunchVerificationStep);
  return {
    project: normalizeProjectState(object.project),
    context: normalizeContextState(object.context),
    confidence: requiredLaunchConfidenceState(object.confidence),
    ...(verificationSteps.length === 0 ? {} : {verificationSteps}),
    warnings: arrayValue(object.warnings).map(normalizeResolutionWarning),
  };
}

function normalizeLaunchVerificationStep(value: unknown): LaunchVerificationStep {
  const object = objectValue(value);
  return {
    id: stringValue(object.id),
    label: stringValue(object.label),
    status: normalizeLaunchVerificationStepStatus(object.status),
    message: stringValue(object.message),
  };
}

function normalizeLaunchVerificationStepStatus(value: unknown): LaunchVerificationStepStatus {
  if (value === "pending") {
    return value;
  }
  return normalizeLaunchConfidenceStatus(value);
}

function requiredLaunchConfidenceState(value: unknown): LaunchConfidenceState {
  const confidence = normalizeLaunchConfidenceState(value);
  if (confidence === undefined) {
    throw new Error("Invalid Dev Context response.");
  }
  return confidence;
}

function normalizeCreateContextResult(value: unknown): CreateContextResult {
  const object = objectValue(value);
  return {
    context: normalizeContextState(object.context),
  };
}
function normalizeProjectsState(value: unknown): ProjectsState { return {projects: arrayValue(objectValue(value).projects).map(normalizeProjectListItem)}; }
function normalizeProjectListItem(value: unknown): ProjectListItem { const object = objectValue(value); const contextId = optionalString(object.contextId); const contextName = optionalString(object.contextName); const lastLaunchedAt = optionalTimestamp(object.lastLaunchedAt); return {project: normalizeProjectState(object.project), ...(contextId === undefined ? {} : {contextId}), ...(contextName === undefined ? {} : {contextName}), ...(lastLaunchedAt === undefined ? {} : {lastLaunchedAt}), running: booleanValue(object.running)}; }
function normalizeDiagnosticsState(value: unknown): DiagnosticsState { return {groups: arrayValue(objectValue(value).groups).map(normalizeDiagnosticGroup)}; }
function normalizeDiagnosticGroup(value: unknown): DiagnosticGroup { const object = objectValue(value); return {id: stringValue(object.id), label: stringValue(object.label), checks: arrayValue(object.checks).map(normalizeDiagnosticCheck)}; }
function normalizeDiagnosticCheck(value: unknown): DiagnosticCheck { const object = objectValue(value); return {id: stringValue(object.id), severity: normalizeLaunchConfidenceStatus(object.severity), label: stringValue(object.label), message: stringValue(object.message), details: arrayValue(object.details).map(normalizeDiagnosticDetail)}; }
function normalizeDiagnosticDetail(value: unknown): DiagnosticDetail { const object = objectValue(value); return {label: stringValue(object.label), value: stringValue(object.value), isPath: booleanValue(object.isPath)}; }
function normalizeRepairActionsState(value: unknown): RepairActionsState { return {actions: arrayValue(objectValue(value).actions).map(normalizeRepairAction)}; }
function normalizeRepairAction(value: unknown): RepairAction { const object = objectValue(value); return {id: stringValue(object.id), label: stringValue(object.label), description: stringValue(object.description), destructive: booleanValue(object.destructive), requiresConfirmation: booleanValue(object.requiresConfirmation), targets: arrayValue(object.targets).map(normalizeRepairTarget)}; }
function normalizeRepairTarget(value: unknown): RepairTarget { const object = objectValue(value); return {label: stringValue(object.label), path: stringValue(object.path), kind: stringValue(object.kind)}; }
function normalizeRunRepairActionResult(value: unknown): RunRepairActionResult { const object = objectValue(value); return {actionId: stringValue(object.actionId), diagnostics: normalizeDiagnosticsState(object.diagnostics)}; }
function normalizeHistoryState(value: unknown): HistoryState { return {entries: arrayValue(objectValue(value).entries).map(normalizeHistoryEntry)}; }
function normalizeHistoryEntry(value: unknown): HistoryEntry { const object = objectValue(value); return {event: stringValue(object.event), category: normalizeHistoryCategory(object.category), timestamp: stringValue(object.timestamp), projectPath: optionalString(object.projectPath), contextId: optionalString(object.contextId), toolId: optionalString(object.toolId), message: stringValue(object.message)}; }
function normalizeRunningEnvironmentsState(value: unknown): RunningEnvironmentsState { return {environments: arrayValue(objectValue(value).environments).map(normalizeRunningEnvironmentState)}; }
function normalizeRunningEnvironmentState(value: unknown): RunningEnvironmentState { const object = objectValue(value); return {id: stringValue(object.id), project: normalizeProjectState(object.project), context: normalizeRunningEnvironmentContextState(object.context), tool: normalizeToolOption(object.tool), startedAt: stringValue(object.startedAt), process: normalizeRunningEnvironmentProcessState(object.process), session: normalizeRunningEnvironmentSessionState(object.session), launch: normalizeRunningEnvironmentLaunchState(object.launch)}; }
function normalizeRunningEnvironmentContextState(value: unknown): RunningEnvironmentContextState { const object = objectValue(value); return {id: stringValue(object.id), name: stringValue(object.name)}; }
function normalizeRunningEnvironmentProcessState(value: unknown): RunningEnvironmentProcessState { const object = objectValue(value); const pid = optionalNumber(object.pid); return {state: stringValue(object.state), ...(pid === undefined ? {} : {pid})}; }
function normalizeRunningEnvironmentSessionState(value: unknown): RunningEnvironmentSessionState { const object = objectValue(value); const id = optionalString(object.id); return {state: stringValue(object.state), ...(id === undefined ? {} : {id})}; }
function normalizeRunningEnvironmentLaunchState(value: unknown): RunningEnvironmentLaunchState { const object = objectValue(value); return {source: stringValue(object.source), resolutionSource: stringValue(object.resolutionSource)}; }

function normalizeHistoryCategory(value: unknown): HistoryCategory {
  if (value === "launch" || value === "warning") {
    return value;
  }
  return "configuration";
}

function normalizeProjectState(value: unknown): ProjectState {
  const object = objectValue(value);
  return {
    name: stringValue(object.name),
    path: stringValue(object.path),
  };
}

function normalizeContextState(value: unknown): ContextState {
  const object = objectValue(value);
	const description = optionalString(object.description);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    ...(description === undefined ? {} : {description}),
    tool: normalizeToolState(object.tool),
    availableTools: arrayValue(object.availableTools).map(normalizeToolOption),
    providers: arrayValue(object.providers).map(normalizeProviderState),
    confidence: normalizeLaunchConfidenceState(object.confidence),
    metadata: optionalStringRecord(object.metadata),
  };
}

function normalizeToolState(value: unknown): ToolState {
  const object = objectValue(value);
	const actionHint = optionalString(object.actionHint);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    status: normalizeLaunchConfidenceStatus(object.status),
    message: stringValue(object.message),
    ...(actionHint === undefined ? {} : {actionHint}),
  };
}

function normalizeToolOption(value: unknown): ToolOption {
  const object = objectValue(value);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
  };
}

function normalizeProviderState(value: unknown): ProviderState {
  const object = objectValue(value);
	const actionHint = optionalString(object.actionHint);
	const setupAction = normalizeProviderSetupAction(object.setupAction);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    enabled: booleanValue(object.enabled),
    state: normalizeProviderReadinessState(object.state),
    explanation: optionalString(object.explanation),
    ...(actionHint === undefined ? {} : {actionHint}),
	...(setupAction === undefined ? {} : {setupAction}),
    identity: normalizeProviderIdentityState(object.identity),
  };
}

function normalizeProviderSetupAction(value: unknown): ProviderSetupAction | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }

  const object = objectValue(value);
  return {
    state: normalizeProviderSetupState(object.state),
    label: stringValue(object.label),
    message: stringValue(object.message),
  };
}

function normalizeProviderSetupState(value: unknown): ProviderSetupState {
  switch (value) {
    case "open_and_configure":
    case "waiting_for_sign_in":
    case "verified":
      return value;
    default:
      throw new Error("Invalid Dev Context response.");
  }
}

function normalizeProviderReadinessState(value: unknown): ProviderReadinessState {
  switch (value) {
    case "ready":
    case "not_configured":
    case "directory_missing":
    case "unavailable":
      return value;
    default:
      throw new Error("Invalid Dev Context response.");
  }
}

function normalizeProviderIdentityState(value: unknown): ProviderIdentityState {
  if (value === undefined || value === null) {
    return { status: "none", fields: [] };
  }

  const object = objectValue(value);
  return {
    status: normalizeProviderIdentityStatus(object.status),
    message: optionalString(object.message),
    fields: arrayValue(object.fields).map(normalizeProviderMetadataField),
  };
}

function normalizeProviderIdentityStatus(value: unknown): ProviderIdentityStatus {
  switch (value) {
    case "verified":
    case "unavailable":
    case "none":
    case "mismatch_evidence":
      return value;
    default:
      throw new Error("Invalid Dev Context response.");
  }
}

function normalizeProviderCredentialSession(value: unknown): ProviderCredentialSession {
  const object = objectValue(value);
  return {
    providerId: stringValue(object.providerId),
    name: stringValue(object.name),
    metadataAvailable: booleanValue(object.metadataAvailable),
    fields: arrayValue(object.fields).map(normalizeProviderMetadataField),
  };
}

function normalizeProviderMetadataField(value: unknown): ProviderMetadataField {
  const object = objectValue(value);
  return {
    label: stringValue(object.label),
    value: stringValue(object.value),
  };
}

function normalizeProjectBindingState(value: unknown): ProjectBindingState {
  const object = objectValue(value);
  return {
    projectPath: stringValue(object.projectPath),
    bound: booleanValue(object.bound),
    contextId: optionalString(object.contextId),
    dangling: booleanValue(object.dangling),
    missingContextId: optionalString(object.missingContextId),
    recovery: optionalString(object.recovery),
  };
}

function normalizeResolutionWarning(value: unknown): ResolutionWarning {
  const object = objectValue(value);
  return {
    code: stringValue(object.code),
    message: stringValue(object.message),
    projectPath: optionalString(object.projectPath),
    boundContextId: optionalString(object.boundContextId),
    requestedContextId: optionalString(object.requestedContextId),
  };
}

function normalizeApplicationError(value: ApplicationErrorLike): DisplayError {
  const technicalDetails = optionalString(value.technicalDetails);
  return {
    code: knownErrorCode(value.code),
    message: stringValue(value.message),
    recovery: stringValue(value.recovery),
    ...(technicalDetails === undefined ? {} : {technicalDetails}),
    contextMismatch: normalizeContextMismatch(value.contextMismatch),
  };
}

function normalizeRejectedError(value: unknown): DisplayError {
  if (isApplicationError(value)) {
    return normalizeApplicationError(value);
  }

  if (value instanceof Error && value.message !== "") {
    return unexpectedError(value.message);
  }

  if (typeof value === "string" && value.trim() !== "") {
    return unexpectedError(value);
  }

  const object = value !== null && typeof value === "object" ? (value as Record<string, unknown>) : undefined;
  const message = object ? optionalString(object.message) : undefined;
  return unexpectedError(message ?? "Dev Context could not complete the request.");
}

function unexpectedError(message: string): DisplayError {
  return {
    code: "unexpected_error",
    message,
    recovery: "Retry the action. If it keeps failing, include the error details in a bug report.",
  };
}

function normalizeContextMismatch(value: unknown): ContextMismatch | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }

  const object = objectValue(value);
  return {
    projectPath: stringValue(object.projectPath),
    boundContextId: stringValue(object.boundContextId),
    requestedContextId: stringValue(object.requestedContextId),
  };
}

interface ApplicationErrorLike {
  code: unknown;
  message: unknown;
  recovery: unknown;
  technicalDetails?: unknown;
  contextMismatch?: unknown;
}

function isApplicationError(value: unknown): value is ApplicationErrorLike {
  if (value === null || typeof value !== "object") {
    return false;
  }

  const object = value as Record<string, unknown>;
  return typeof object.code === "string" && typeof object.message === "string" && typeof object.recovery === "string";
}

function objectValue(value: unknown): Record<string, unknown> {
  if (value !== null && typeof value === "object") {
    return value as Record<string, unknown>;
  }
  throw new Error("Invalid Dev Context response.");
}

function arrayValue(value: unknown): unknown[] {
  if (value === undefined || value === null) {
    return [];
  }
  if (Array.isArray(value)) {
    return value;
  }
  throw new Error("Invalid Dev Context response.");
}

function stringValue(value: unknown): string {
  if (typeof value === "string") {
    return value;
  }
  throw new Error("Invalid Dev Context response.");
}

function timestampValue(value: unknown): string {
  const timestamp = stringValue(value);
  if (Number.isNaN(Date.parse(timestamp))) {
    throw new Error("Invalid Dev Context response.");
  }
  return timestamp;
}

function optionalTimestamp(value: unknown): string | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }
  return timestampValue(value);
}

function optionalString(value: unknown): string | undefined {
  if (value === undefined || value === null || value === "") {
    return undefined;
  }
  return stringValue(value);
}

function booleanValue(value: unknown): boolean {
  if (typeof value === "boolean") {
    return value;
  }
  throw new Error("Invalid Dev Context response.");
}

function numberValue(value: unknown): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  throw new Error("Invalid Dev Context response.");
}

function optionalNumber(value: unknown): number | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }
  return numberValue(value);
}

function optionalStringRecord(value: unknown): Record<string, string> | undefined {
  if (value === undefined || value === null) {
    return undefined;
  }

  const object = objectValue(value);
  const record: Record<string, string> = {};
  for (const [key, entry] of Object.entries(object)) {
    record[key] = stringValue(entry);
  }
  return record;
}

function knownErrorCode(value: unknown): ErrorCode {
  switch (value) {
    case "canceled":
    case "validation_error":
    case "context_mismatch_requires_confirmation":
    case "launch_error":
    case "internal_error":
      return value;
    default:
      return "unexpected_error";
  }
}
