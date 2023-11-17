// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {awsiamidc} from '../models';
import {context} from '../models';
import {logging} from '../models';

export function FinalizeRefreshAccessToken(arg1:string,arg2:string,arg3:string,arg4:string,arg5:string):Promise<void>;

export function FinalizeSetup(arg1:string,arg2:string,arg3:string,arg4:string,arg5:string):Promise<void>;

export function GetInstanceData(arg1:string):Promise<awsiamidc.AwsIdentityCenterCardData>;

export function Init(arg1:context.Context,arg2:logging.ErrorHandler):Promise<void>;

export function RefreshAccessToken(arg1:string):Promise<awsiamidc.AuthorizeDeviceFlowResult>;

export function Setup(arg1:string,arg2:string):Promise<awsiamidc.AuthorizeDeviceFlowResult>;
