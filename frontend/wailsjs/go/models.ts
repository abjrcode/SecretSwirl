export namespace awscredssink {
	
	export class AwsCredentialsSinkInstance {
	    instanceId: string;
	    version: number;
	    filePath: string;
	    awsProfileName: string;
	    label: string;
	    providerCode: string;
	    providerId: string;
	    createdAt: number;
	    lastDrainedAt?: number;
	
	    static createFrom(source: any = {}) {
	        return new AwsCredentialsSinkInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.version = source["version"];
	        this.filePath = source["filePath"];
	        this.awsProfileName = source["awsProfileName"];
	        this.label = source["label"];
	        this.providerCode = source["providerCode"];
	        this.providerId = source["providerId"];
	        this.createdAt = source["createdAt"];
	        this.lastDrainedAt = source["lastDrainedAt"];
	    }
	}
	export class AwsCredentialsSink_NewInstanceCommandInput {
	    filePath: string;
	    awsProfileName: string;
	    label: string;
	    providerCode: string;
	    providerId: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsCredentialsSink_NewInstanceCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.filePath = source["filePath"];
	        this.awsProfileName = source["awsProfileName"];
	        this.label = source["label"];
	        this.providerCode = source["providerCode"];
	        this.providerId = source["providerId"];
	    }
	}

}

export namespace awsidc {
	
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
	
	export class AwsIdc_CopyRoleCredentialsCommandInput {
	    instanceId: string;
	    accountId: string;
	    roleName: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdc_CopyRoleCredentialsCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.accountId = source["accountId"];
	        this.roleName = source["roleName"];
	    }
	}
	export class AwsIdc_FinalizeRefreshAccessTokenCommandInput {
	    instanceId: string;
	    region: string;
	    userCode: string;
	    deviceCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdc_FinalizeRefreshAccessTokenCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.region = source["region"];
	        this.userCode = source["userCode"];
	        this.deviceCode = source["deviceCode"];
	    }
	}
	export class AwsIdc_FinalizeSetupCommandInput {
	    clientId: string;
	    startUrl: string;
	    awsRegion: string;
	    label: string;
	    userCode: string;
	    deviceCode: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdc_FinalizeSetupCommandInput(source);
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
	export class AwsIdc_SaveRoleCredentialsCommandInput {
	    instanceId: string;
	    accountId: string;
	    roleName: string;
	    awsProfile: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdc_SaveRoleCredentialsCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.instanceId = source["instanceId"];
	        this.accountId = source["accountId"];
	        this.roleName = source["roleName"];
	        this.awsProfile = source["awsProfile"];
	    }
	}
	export class AwsIdc_SetupCommandInput {
	    startUrl: string;
	    awsRegion: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new AwsIdc_SetupCommandInput(source);
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
	
	export class AwsIdentityCenterCardData {
	    instanceId: string;
	    enabled: boolean;
	    label: string;
	    isFavorite: boolean;
	    isAccessTokenExpired: boolean;
	    accessTokenExpiresIn: string;
	    accounts: AwsIdentityCenterAccount[];
	    sinks: plumbing.SinkInstance[];
	
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
	        this.sinks = this.convertValues(source["sinks"], plumbing.SinkInstance);
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
	export class CompatibleSink {
	    code: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new CompatibleSink(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.name = source["name"];
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

export namespace plumbing {
	
	export class DisconnectSinkCommandInput {
	    sinkCode: string;
	    sinkId: string;
	
	    static createFrom(source: any = {}) {
	        return new DisconnectSinkCommandInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sinkCode = source["sinkCode"];
	        this.sinkId = source["sinkId"];
	    }
	}
	export class SinkInstance {
	    sinkCode: string;
	    sinkId: string;
	
	    static createFrom(source: any = {}) {
	        return new SinkInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sinkCode = source["sinkCode"];
	        this.sinkId = source["sinkId"];
	    }
	}

}

