export namespace config {
	
	export class Config {
	    server_url: string;
	    model: string;
	    api_key?: string;
	    default_style: string;
	    auto_start: boolean;
	    hotkey: string;
	    monitor_clipboard: boolean;
	    first_run: boolean;
	    custom_prompts?: Record<string, any>;
	    auto_paste_mode: string;
	    popup_position_mode: string;
	    mini_mode: boolean;
	    auto_minimize_on_copy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server_url = source["server_url"];
	        this.model = source["model"];
	        this.api_key = source["api_key"];
	        this.default_style = source["default_style"];
	        this.auto_start = source["auto_start"];
	        this.hotkey = source["hotkey"];
	        this.monitor_clipboard = source["monitor_clipboard"];
	        this.first_run = source["first_run"];
	        this.custom_prompts = source["custom_prompts"];
	        this.auto_paste_mode = source["auto_paste_mode"];
	        this.popup_position_mode = source["popup_position_mode"];
	        this.mini_mode = source["mini_mode"];
	        this.auto_minimize_on_copy = source["auto_minimize_on_copy"];
	    }
	}

}

export namespace diffmatchpatch {
	
	export class Diff {
	    Type: number;
	    Text: string;
	
	    static createFrom(source: any = {}) {
	        return new Diff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Text = source["Text"];
	    }
	}

}

export namespace rewriter {
	
	export class DiffResult {
	    diffs: diffmatchpatch.Diff[];
	    html: string;
	    has_diff: boolean;
	    additions: number;
	    deletions: number;
	
	    static createFrom(source: any = {}) {
	        return new DiffResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.diffs = this.convertValues(source["diffs"], diffmatchpatch.Diff);
	        this.html = source["html"];
	        this.has_diff = source["has_diff"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
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
	export class RewriteOption {
	    style: string;
	    text: string;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new RewriteOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.style = source["style"];
	        this.text = source["text"];
	        this.error = source["error"];
	    }
	}
	export class StyleInfoData {
	    label: string;
	    icon: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new StyleInfoData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	    }
	}
	export class TextTypeDetected {
	    type: string;
	    label: string;
	    icon: string;
	    confidence: number;
	
	    static createFrom(source: any = {}) {
	        return new TextTypeDetected(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.confidence = source["confidence"];
	    }
	}
	export class TextTypeInfo {
	    Type: string;
	    Label: string;
	    Icon: string;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new TextTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Type = source["Type"];
	        this.Label = source["Label"];
	        this.Icon = source["Icon"];
	        this.Description = source["Description"];
	    }
	}

}

