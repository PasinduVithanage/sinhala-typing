export namespace config {

	export class Config {
	    mapping_profile: string;
	    auto_start: boolean;
	    auto_fix_clipboard: boolean;
	    use_zwj_clusters: boolean;
	    toggle_hotkey: string;
	    run_at_startup: boolean;

	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mapping_profile = source["mapping_profile"];
	        this.auto_start = source["auto_start"];
	        this.auto_fix_clipboard = source["auto_fix_clipboard"];
	        this.use_zwj_clusters = source["use_zwj_clusters"];
	        this.toggle_hotkey = source["toggle_hotkey"];
	        this.run_at_startup = source["run_at_startup"];
	    }
	}

}

export namespace normalization {
	
	export class Detection {
	    StartIndex: number;
	    EndIndex: number;
	    Issue: number;
	    Context: string;
	
	    static createFrom(source: any = {}) {
	        return new Detection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.StartIndex = source["StartIndex"];
	        this.EndIndex = source["EndIndex"];
	        this.Issue = source["Issue"];
	        this.Context = source["Context"];
	    }
	}

}

