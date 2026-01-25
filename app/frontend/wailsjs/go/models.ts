export namespace config {
	
	export class AppConfig {
	    api_key: string;
	    provider: string;
	    model: string;
	    output_path: string;
	    hardware_target: string;
	    delete_no_confirm: boolean;
	    incremental_save: boolean;
	    default_exp_pedal: number;
	    variax_enabled: boolean;
	    variax_hardware_model: string;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.api_key = source["api_key"];
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.output_path = source["output_path"];
	        this.hardware_target = source["hardware_target"];
	        this.delete_no_confirm = source["delete_no_confirm"];
	        this.incremental_save = source["incremental_save"];
	        this.default_exp_pedal = source["default_exp_pedal"];
	        this.variax_enabled = source["variax_enabled"];
	        this.variax_hardware_model = source["variax_hardware_model"];
	    }
	}

}

export namespace gemini {
	
	export class ChatMessage {
	    role: string;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.role = source["role"];
	        this.content = source["content"];
	    }
	}
	export class RigComponent {
	    type: string;
	    name: string;
	    description: string;
	    settings: string;
	
	    static createFrom(source: any = {}) {
	        return new RigComponent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.settings = source["settings"];
	    }
	}
	export class Snapshot {
	    name: string;
	    active_blocks: string[];
	    guitar_model?: string;
	    tuning?: string;
	    params?: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Snapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.active_blocks = source["active_blocks"];
	        this.guitar_model = source["guitar_model"];
	        this.tuning = source["tuning"];
	        this.params = source["params"];
	    }
	}
	export class RigDescription {
	    suggested_name: string;
	    explanation: string;
	    guitar_model: string;
	    tuning: string;
	    chain: RigComponent[];
	    snapshots?: Snapshot[];
	
	    static createFrom(source: any = {}) {
	        return new RigDescription(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.suggested_name = source["suggested_name"];
	        this.explanation = source["explanation"];
	        this.guitar_model = source["guitar_model"];
	        this.tuning = source["tuning"];
	        this.chain = this.convertValues(source["chain"], RigComponent);
	        this.snapshots = this.convertValues(source["snapshots"], Snapshot);
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

}

