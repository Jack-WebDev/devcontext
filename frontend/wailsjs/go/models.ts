export namespace application {
	
	export class BindProjectRequest {
	    projectPath?: string;
	    contextId: string;
	
	    static createFrom(source: any = {}) {
	        return new BindProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.contextId = source["contextId"];
	    }
	}
	export class ContextMismatch {
	    projectPath: string;
	    boundContextId: string;
	    requestedContextId: string;
	
	    static createFrom(source: any = {}) {
	        return new ContextMismatch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.boundContextId = source["boundContextId"];
	        this.requestedContextId = source["requestedContextId"];
	    }
	}
	export class ProviderState {
	    id: string;
	    name: string;
	    enabled: boolean;
	    state: string;
	    explanation?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProviderState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.enabled = source["enabled"];
	        this.state = source["state"];
	        this.explanation = source["explanation"];
	    }
	}
	export class EditorState {
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new EditorState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	    }
	}
	export class ContextState {
	    id: string;
	    name: string;
	    editor: EditorState;
	    providers: ProviderState[];
	    metadata?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ContextState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.editor = this.convertValues(source["editor"], EditorState);
	        this.providers = this.convertValues(source["providers"], ProviderState);
	        this.metadata = source["metadata"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Error {
	    code: string;
	    message: string;
	    recovery: string;
	    contextMismatch?: ContextMismatch;
	
	    static createFrom(source: any = {}) {
	        return new Error(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.recovery = source["recovery"];
	        this.contextMismatch = this.convertValues(source["contextMismatch"], ContextMismatch);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class GetLaunchStateRequest {
	    projectPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetLaunchStateRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	    }
	}
	export class LaunchProjectRequest {
	    projectPath?: string;
	    contextId: string;
	    confirmContextMismatch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LaunchProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.contextId = source["contextId"];
	        this.confirmContextMismatch = source["confirmContextMismatch"];
	    }
	}
	export class ResolutionWarning {
	    code: string;
	    message: string;
	    projectPath?: string;
	    boundContextId?: string;
	    requestedContextId?: string;
	
	    static createFrom(source: any = {}) {
	        return new ResolutionWarning(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	        this.projectPath = source["projectPath"];
	        this.boundContextId = source["boundContextId"];
	        this.requestedContextId = source["requestedContextId"];
	    }
	}
	export class ProjectState {
	    name: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.path = source["path"];
	    }
	}
	export class LaunchProjectResult {
	    project: ProjectState;
	    context: ContextState;
	    warnings?: ResolutionWarning[];
	
	    static createFrom(source: any = {}) {
	        return new LaunchProjectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], ProjectState);
	        this.context = this.convertValues(source["context"], ContextState);
	        this.warnings = this.convertValues(source["warnings"], ResolutionWarning);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ProjectBindingState {
	    projectPath: string;
	    bound: boolean;
	    contextId?: string;
	    dangling: boolean;
	    missingContextId?: string;
	    recovery?: string;
	
	    static createFrom(source: any = {}) {
	        return new ProjectBindingState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.bound = source["bound"];
	        this.contextId = source["contextId"];
	        this.dangling = source["dangling"];
	        this.missingContextId = source["missingContextId"];
	        this.recovery = source["recovery"];
	    }
	}
	export class LaunchState {
	    project: ProjectState;
	    contexts: ContextState[];
	    binding: ProjectBindingState;
	    selectedContextId?: string;
	    selectionRequired: boolean;
	    resolutionSource?: string;
	    warnings?: ResolutionWarning[];
	    firstRun: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LaunchState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project = this.convertValues(source["project"], ProjectState);
	        this.contexts = this.convertValues(source["contexts"], ContextState);
	        this.binding = this.convertValues(source["binding"], ProjectBindingState);
	        this.selectedContextId = source["selectedContextId"];
	        this.selectionRequired = source["selectionRequired"];
	        this.resolutionSource = source["resolutionSource"];
	        this.warnings = this.convertValues(source["warnings"], ResolutionWarning);
	        this.firstRun = source["firstRun"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class UnbindProjectRequest {
	    projectPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new UnbindProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	    }
	}

}

