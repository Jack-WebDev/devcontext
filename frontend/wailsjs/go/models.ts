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
	export class CreateContextRequest {
	    contextId: string;
	    name?: string;
	    description?: string;
	    icon?: string;
	    accent?: string;
	    enabledProviderIds?: string[];
	    toolId?: string;
	    importProviderIds?: string[];
	
	    static createFrom(source: any = {}) {
	        return new CreateContextRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.contextId = source["contextId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.icon = source["icon"];
	        this.accent = source["accent"];
	        this.enabledProviderIds = source["enabledProviderIds"];
	        this.toolId = source["toolId"];
	        this.importProviderIds = source["importProviderIds"];
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

