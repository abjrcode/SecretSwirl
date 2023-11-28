import { RunAppCommand } from "../../wailsjs/go/main/AppController";
import { awsiamidc, main } from "../../wailsjs/go/models";

export function Auth_IsVaultConfigured(): Promise<boolean> {
  return RunAppCommand("Auth_IsVaultConfigured", {})
}

export function Auth_ConfigureVault(password: string) {
  return RunAppCommand("Auth_ConfigureVault", {
    password,
  })
}

export function Auth_Unlock(password: string): Promise<boolean> {
  return RunAppCommand("Auth_Unlock", {
    password,
  })
}

export function Auth_Lock() {
  return RunAppCommand("Auth_Lock", {})
}


export function Dashboard_ListProviders(): Promise<main.Provider[]> {
  return RunAppCommand("Dashboard_ListProviders", {})
}
export function Dashboard_ListFavorites(): Promise<main.FavoriteInstance[]> {
  return RunAppCommand("Dashboard_ListFavorites", {})
}


export function AwsIamIdc_ListInstances(): Promise<string[]> {
  return RunAppCommand("AwsIamIdc_ListInstances", {})
}

export function AwsIamIdc_GetInstanceData(instanceId: string, forceRefresh: boolean): Promise<awsiamidc.AwsIdentityCenterCardData> {
  return RunAppCommand("AwsIamIdc_GetInstanceData", {
    instanceId,
    forceRefresh,
  })
}
export function AwsIamIdc_GetRoleCredentials(input: awsiamidc.AwsIamIdc_GetRoleCredentialsCommandInput): Promise<awsiamidc.AwsIdentityCenterAccountRoleCredentials> {
  return RunAppCommand("AwsIamIdc_GetRoleCredentials", input)
}

export function AwsIamIdc_Setup(input: awsiamidc.AwsIamIdc_SetupCommandInput): Promise<awsiamidc.AuthorizeDeviceFlowResult> {
  return RunAppCommand("AwsIamIdc_Setup", input)
}

export function AwsIamIdc_FinalizeSetup(input: awsiamidc.AwsIamIdc_FinalizeSetupCommandInput): Promise<string> {
  return RunAppCommand("AwsIamIdc_FinalizeSetup", input)
}

export function AwsIamIdc_MarkAsFavorite(instanceId: string) {
  return RunAppCommand("AwsIamIdc_MarkAsFavorite", { instanceId })
}
export function AwsIamIdc_UnmarkAsFavorite(instanceId: string) {
  return RunAppCommand("AwsIamIdc_UnmarkAsFavorite", { instanceId })
}
export function AwsIamIdc_RefreshAccessToken(instanceId: string): Promise<awsiamidc.AuthorizeDeviceFlowResult> {
  return RunAppCommand("AwsIamIdc_RefreshAccessToken", { instanceId })
}
export function AwsIamIdc_FinalizeRefreshAccessToken(input: awsiamidc.AwsIamIdc_FinalizeRefreshAccessTokenCommandInput) {
  return RunAppCommand("AwsIamIdc_FinalizeRefreshAccessToken", input)
}