import { RunAppCommand } from "../../wailsjs/go/main/AppController";
import { awsidc, main } from "../../wailsjs/go/models";

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


export function AwsIdc_ListInstances(): Promise<string[]> {
  return RunAppCommand("AwsIdc_ListInstances", {})
}

export function AwsIdc_GetInstanceData(instanceId: string, forceRefresh: boolean): Promise<awsidc.AwsIdentityCenterCardData> {
  return RunAppCommand("AwsIdc_GetInstanceData", {
    instanceId,
    forceRefresh,
  })
}
export function AwsIdc_GetRoleCredentials(input: awsidc.AwsIdc_GetRoleCredentialsCommandInput): Promise<awsidc.AwsIdentityCenterAccountRoleCredentials> {
  return RunAppCommand("AwsIdc_GetRoleCredentials", input)
}

export function AwsIdc_Setup(input: awsidc.AwsIdc_SetupCommandInput): Promise<awsidc.AuthorizeDeviceFlowResult> {
  return RunAppCommand("AwsIdc_Setup", input)
}

export function AwsIdc_FinalizeSetup(input: awsidc.AwsIdc_FinalizeSetupCommandInput): Promise<string> {
  return RunAppCommand("AwsIdc_FinalizeSetup", input)
}

export function AwsIdc_MarkAsFavorite(instanceId: string) {
  return RunAppCommand("AwsIdc_MarkAsFavorite", { instanceId })
}
export function AwsIdc_UnmarkAsFavorite(instanceId: string) {
  return RunAppCommand("AwsIdc_UnmarkAsFavorite", { instanceId })
}
export function AwsIdc_RefreshAccessToken(instanceId: string): Promise<awsidc.AuthorizeDeviceFlowResult> {
  return RunAppCommand("AwsIdc_RefreshAccessToken", { instanceId })
}
export function AwsIdc_FinalizeRefreshAccessToken(input: awsidc.AwsIdc_FinalizeRefreshAccessTokenCommandInput) {
  return RunAppCommand("AwsIdc_FinalizeRefreshAccessToken", input)
}