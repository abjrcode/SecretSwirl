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
	export class AwsIamIdc_FinalizeRefreshAccessTokenCommandInput {
	    instanceId: string;
	    region: string;
	    userCode: string;
	    deviceCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIamIdc_FinalizeRefreshAccessTokenCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.region = source["region"];
	        this.userCode = source["userCode"];
	        this.deviceCode = source["deviceCode"];
	    }
	}
	export class AwsIamIdc_FinalizeSetupCommandInput {
	    clientId: string;
	    startUrl: string;
	    awsRegion: string;
	    label: string;
	    userCode: string;
	    deviceCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIamIdc_FinalizeSetupCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clientId = source["clientId"];
	        this.startUrl = source["startUrl"];
	        this.awsRegion = source["awsRegion"];
	        this.label = source["label"];
	        this.userCode = source["userCode"];
	        this.deviceCode = source["deviceCode"];
	    }
	}
	export class AwsIamIdc_GetRoleCredentialsCommandInput {
	    instanceId: string;
	    accountId: string;
	    roleName: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIamIdc_GetRoleCredentialsCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.accountId = source["accountId"];
	        this.roleName = source["roleName"];
	    }
	}
	export class AwsIamIdc_SetupCommandInput {
	    startUrl: string;
	    awsRegion: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIamIdc_SetupCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.startUrl = source["startUrl"];
	        this.awsRegion = source["awsRegion"];
	        this.label = source["label"];
	    }
	}
	export class AwsIdentityCenterAccountRole {
	    roleName: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdentityCenterAccountRole(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.roleName = source["roleName"];
	    }
	}
	export class AwsIdentityCenterAccount {
	    accountId: string;
	    accountName: string;
	    roles: AwsIdentityCenterAccountRole[];
	
	    static createFrom(source: any = {}) {
	        return new AwsIdentityCenterAccount(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accountId = source["accountId"];
	        this.accountName = source["accountName"];
	        this.roles = this.convertValues(source["roles"], AwsIdentityCenterAccountRole);
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
	
	export class AwsIdentityCenterAccountRoleCredentials {
	    accessKeyId: string;
	    secretAccessKey: string;
	    sessionToken: string;
	    expiration: number;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdentityCenterAccountRoleCredentials(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accessKeyId = source["accessKeyId"];
	        this.secretAccessKey = source["secretAccessKey"];
	        this.sessionToken = source["sessionToken"];
	        this.expiration = source["expiration"];
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
	
	export class Auth_ConfigureVaultCommandInput {
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new Auth_ConfigureVaultCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.password = source["password"];
	    }
	}
	export class Auth_UnlockCommandInput {
	    password: string;
	
	    static createFrom(source: any = {}) {
	        return new Auth_UnlockCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.password = source["password"];
	    }
	}
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

