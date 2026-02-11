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
	    }
	}

}

export namespace rewriter {
	
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

}

export namespace struct { Label string; Icon string; Description string } {
	
	export class  {
	    Label: string;
	    Icon: string;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Label = source["Label"];
	        this.Icon = source["Icon"];
	        this.Description = source["Description"];
	    }
	}

}

export namespace struct { Type string "json:\"type\""; Label string "json:\"label\""; Icon string "json:\"icon\""; Confidence float64 "json:\"confidence\"" } {
	
	export class  {
	    type: string;
	    label: string;
	    icon: string;
	    confidence: number;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.confidence = source["confidence"];
	    }
	}

}

export namespace struct { Type string "json:\"type\""; Label string "json:\"label\""; Icon string "json:\"icon\""; Description string "json:\"description\"" } {
	
	export class  {
	    type: string;
	    label: string;
	    icon: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new (source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.label = source["label"];
	        this.icon = source["icon"];
	        this.description = source["description"];
	    }
	}

}

