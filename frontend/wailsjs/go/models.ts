export namespace awsiamidc {
	
	export class AuthorizeDeviceFlowResult {
	    instanceId: string;
	    startUrl: string;
	    region: string;
	    label: string;
	    clientId: string;
	    verificationUri: string;
	    userCode: string;
	    expiresIn: number;
	    deviceCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthorizeDeviceFlowResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.startUrl = source["startUrl"];
	        this.region = source["region"];
	        this.label = source["label"];
	        this.clientId = source["clientId"];
	        this.verificationUri = source["verificationUri"];
	        this.userCode = source["userCode"];
	        this.expiresIn = source["expiresIn"];
	        this.deviceCode = source["deviceCode"];
	    }
	}
	export class AwsIdentityCenterAccount {
	    accountId: string;
	    accountName: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdentityCenterAccount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountId = source["accountId"];
	        this.accountName = source["accountName"];
	    }
	}
	export class AwsIdentityCenterCardData {
	    instanceId: string;
	    enabled: boolean;
	    label: string;
	    isFavorite: boolean;
	    isAccessTokenExpired: boolean;
	    accessTokenExpiresIn: string;
	    accounts: AwsIdentityCenterAccount[];
	
	    static createFrom(source: any = {}) {
	        return new AwsIdentityCenterCardData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.enabled = source["enabled"];
	        this.label = source["label"];
	        this.isFavorite = source["isFavorite"];
	        this.isAccessTokenExpired = source["isAccessTokenExpired"];
	        this.accessTokenExpiresIn = source["accessTokenExpiresIn"];
	        this.accounts = this.convertValues(source["accounts"], AwsIdentityCenterAccount);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

export namespace main {
	
	export class FavoriteInstance {
	    providerCode: string;
	    instanceId: string;
	
	    static createFrom(source: any = {}) {
	        return new FavoriteInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.providerCode = source["providerCode"];
	        this.instanceId = source["instanceId"];
	    }
	}
	export class Provider {
	    code: string;
	    name: string;
	    iconSvgBase64: string;
	
	    static createFrom(source: any = {}) {
	        return new Provider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
	        this.iconSvgBase64 = source["iconSvgBase64"];
	    }
	}

}

