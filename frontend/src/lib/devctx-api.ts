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
  editor: EditorState;
  providers: ProviderState[];
  confidence?: LaunchConfidenceState;
  metadata?: Record<string, string>;
}

export interface EditorState {
  type: string;
}

export interface ProviderState {
  id: string;
  name: string;
  enabled: boolean;
  state: ProviderReadinessState;
  explanation?: string;
  identity: ProviderIdentityState;
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

export type LaunchConfidenceCheckComponent = "claude" | "codex" | "vscode" | "isolation";

export interface LaunchConfidenceState {
  contextId: string;
  status: LaunchConfidenceStatus;
  checks: LaunchConfidenceCheck[];
}

export interface LaunchConfidenceCheck {
  component: LaunchConfidenceCheckComponent;
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
  warnings: ResolutionWarning[];
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
  importProviderIds?: string[];
}

export interface CreateContextResult {
  context: ContextState;
}

export interface ProviderCredentialSession {
  providerId: string;
  name: string;
  metadataAvailable: boolean;
  fields: ProviderMetadataField[];
}

export interface DevContextApi {
  getLaunchState(request?: GetLaunchStateRequest): Promise<ApiResult<LaunchState>>;
  preflightLaunchProject(request: PreflightLaunchProjectRequest): Promise<ApiResult<PreflightLaunchProjectResult>>;
  launchProject(request: LaunchProjectRequest): Promise<ApiResult<LaunchProjectResult>>;
  bindProject(request: BindProjectRequest): Promise<ApiResult<ProjectBindingState>>;
  unbindProject(request?: UnbindProjectRequest): Promise<ApiResult<ProjectBindingState>>;
  createContext(request: CreateContextRequest): Promise<ApiResult<CreateContextResult>>;
}

export interface WailsBindings {
  getLaunchState(request: GetLaunchStateRequest): Promise<unknown>;
  preflightLaunchProject(request: PreflightLaunchProjectRequest): Promise<unknown>;
  launchProject(request: LaunchProjectRequest): Promise<unknown>;
  bindProject(request: BindProjectRequest): Promise<unknown>;
  unbindProject(request: UnbindProjectRequest): Promise<unknown>;
  createContext(request: CreateContextRequest): Promise<unknown>;
}

export function createDevContextApi(bindings: WailsBindings = generatedBindings): DevContextApi {
  return {
    getLaunchState(request = {}) {
      return callBinding(() => bindings.getLaunchState(request), normalizeLaunchState);
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
  };
}

const generatedBindings: WailsBindings = {
  async getLaunchState(request) {
    const bindings = await import("../../wailsjs/go/wailsapp/App");
    return bindings.GetLaunchState(request);
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
  return {
    component: normalizeLaunchConfidenceCheckComponent(object.component),
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
    case "claude":
    case "codex":
    case "vscode":
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
  return {
    project: normalizeProjectState(object.project),
    context: normalizeContextState(object.context),
    confidence: requiredLaunchConfidenceState(object.confidence),
    warnings: arrayValue(object.warnings).map(normalizeResolutionWarning),
  };
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

function normalizeProjectState(value: unknown): ProjectState {
  const object = objectValue(value);
  return {
    name: stringValue(object.name),
    path: stringValue(object.path),
  };
}

function normalizeContextState(value: unknown): ContextState {
  const object = objectValue(value);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    editor: normalizeEditorState(object.editor),
    providers: arrayValue(object.providers).map(normalizeProviderState),
    confidence: normalizeLaunchConfidenceState(object.confidence),
    metadata: optionalStringRecord(object.metadata),
  };
}

function normalizeEditorState(value: unknown): EditorState {
  const object = objectValue(value);
  return {
    type: stringValue(object.type),
  };
}

function normalizeProviderState(value: unknown): ProviderState {
  const object = objectValue(value);
  return {
    id: stringValue(object.id),
    name: stringValue(object.name),
    enabled: booleanValue(object.enabled),
    state: normalizeProviderReadinessState(object.state),
    explanation: optionalString(object.explanation),
    identity: normalizeProviderIdentityState(object.identity),
  };
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
  return {
    code: knownErrorCode(value.code),
    message: stringValue(value.message),
    recovery: stringValue(value.recovery),
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
