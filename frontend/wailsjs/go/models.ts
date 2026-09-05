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
	export class ContextTransferTool {
	    id: string;
	    options?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ContextTransferTool(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.options = source["options"];
	    }
	}
	export class ContextTransferLaunchTarget {
	    defaultTool: string;
	    tools: ContextTransferTool[];
	
	    static createFrom(source: any = {}) {
	        return new ContextTransferLaunchTarget(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultTool = source["defaultTool"];
	        this.tools = this.convertValues(source["tools"], ContextTransferTool);
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
	export class ContextTransferProvider {
	    id: string;
	    enabled: boolean;
	    options?: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ContextTransferProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.options = source["options"];
	    }
	}
	export class ContextTransferMetadata {
	    name: string;
	    metadata?: Record<string, string>;
	    providers: ContextTransferProvider[];
	    launchTarget: ContextTransferLaunchTarget;
	
	    static createFrom(source: any = {}) {
	        return new ContextTransferMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.metadata = source["metadata"];
	        this.providers = this.convertValues(source["providers"], ContextTransferProvider);
	        this.launchTarget = this.convertValues(source["launchTarget"], ContextTransferLaunchTarget);
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
	export class ContextMetadataExport {
	    version: number;
	    context: ContextTransferMetadata;
	
	    static createFrom(source: any = {}) {
	        return new ContextMetadataExport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.context = this.convertValues(source["context"], ContextTransferMetadata);
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
	
	
	
	
	export class CreateContextRequest {
	    contextId: string;
	    templateId?: string;
	    name?: string;
	    purpose?: string;
	    description?: string;
	    icon?: string;
	    accent?: string;
	    enabledProviderIds?: string[];
	    enabledDevelopmentToolIds?: string[];
	    toolId?: string;
	    importProviderIds?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateContextRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	        this.templateId = source["templateId"];
	        this.name = source["name"];
	        this.purpose = source["purpose"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.accent = source["accent"];
	        this.enabledProviderIds = source["enabledProviderIds"];
	        this.enabledDevelopmentToolIds = source["enabledDevelopmentToolIds"];
	        this.toolId = source["toolId"];
	        this.importProviderIds = source["importProviderIds"];
	    }
	}
	export class DuplicateContextRequest {
	    sourceContextId: string;
	    contextId: string;
	    name?: string;
	
	    static createFrom(source: any = {}) {
	        return new DuplicateContextRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sourceContextId = source["sourceContextId"];
	        this.contextId = source["contextId"];
	        this.name = source["name"];
	    }
	}
	export class ExportContextMetadataRequest {
	    contextId: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportContextMetadataRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	    }
	}
	export class GetContextDetailsRequest {
	    contextId: string;
	
	    static createFrom(source: any = {}) {
	        return new GetContextDetailsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	    }
	}
	export class GetDiagnosticsRequest {
	    contextId?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetDiagnosticsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	    }
	}
	export class GetHomeDashboardRequest {
	    projectPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new GetHomeDashboardRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
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
	export class GetRepairActionsRequest {
	    contextId: string;
	
	    static createFrom(source: any = {}) {
	        return new GetRepairActionsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	    }
	}
	export class ImportContextMetadataRequest {
	    contextId: string;
	    export: ContextMetadataExport;
	
	    static createFrom(source: any = {}) {
	        return new ImportContextMetadataRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	        this.export = this.convertValues(source["export"], ContextMetadataExport);
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
	export class PreflightLaunchProjectRequest {
	    projectPath?: string;
	    contextId: string;
	    confirmContextMismatch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PreflightLaunchProjectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	        this.contextId = source["contextId"];
	        this.confirmContextMismatch = source["confirmContextMismatch"];
	    }
	}
	export class RunRepairActionRequest {
	    contextId: string;
	    actionId: string;
	    confirmDestructive: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunRepairActionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	        this.actionId = source["actionId"];
	        this.confirmDestructive = source["confirmDestructive"];
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
	export class UpdateSettingsRequest {
	    closeAfterLaunch: boolean;
	    launchVerification: boolean;
	    rememberProjects: boolean;
	    trayEnabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UpdateSettingsRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.closeAfterLaunch = source["closeAfterLaunch"];
	        this.launchVerification = source["launchVerification"];
	        this.rememberProjects = source["rememberProjects"];
	        this.trayEnabled = source["trayEnabled"];
	    }
	}
	export class ValidateProjectDirectoryRequest {
	    projectPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidateProjectDirectoryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.projectPath = source["projectPath"];
	    }
	}

}

export namespace wailsapp {
	
	export class ApplicationMode {
	    type: string;
	    projectPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new ApplicationMode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.projectPath = source["projectPath"];
	    }
	}

}

