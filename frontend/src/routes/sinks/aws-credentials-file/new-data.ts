import { ActionFunctionArgs } from "react-router-dom"
import { ActionDataResult } from "../../../utils/action-data-result"
import { AwsCredentialsFile_NewInstance } from "../../../utils/ipc-adapter"

export enum AwsCredentialsFileSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
  ErrInvalidLabel = "INVALID_LABEL",
  ErrInvalidAwsProfileName = "INVALID_AWS_PROFILE_NAME",
  ErrInvalidProviderCode = "INVALID_PROVIDER_CODE",
  ErrInvalidProviderId = "INVALID_PROVIDER_ID",
}

export type AwsCredentialsFileSetupResult = ActionDataResult<string, AwsCredentialsFileSetupError>

export async function awsCredentialsFileSetupAction({ request }: ActionFunctionArgs): Promise<AwsCredentialsFileSetupResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const providerCode = updates["providerCode"].toString()
  const providerId = updates["providerId"].toString()
  const filePath = updates["filePath"].toString()
  const awsProfileName = updates["awsProfileName"].toString()
  const label = updates["label"].toString()

  try {
    return {
      success: true, result: await AwsCredentialsFile_NewInstance({
        providerCode,
        providerId,
        filePath,
        awsProfileName,
        label,
      })
    }
  } catch (e) {
    switch (e) {
      case AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered, error: e }
      case AwsCredentialsFileSetupError.ErrInvalidLabel:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInvalidLabel, error: e }
      case AwsCredentialsFileSetupError.ErrInvalidAwsProfileName:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInvalidAwsProfileName, error: e }
      case AwsCredentialsFileSetupError.ErrInvalidProviderCode:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInvalidProviderCode, error: e }
      case AwsCredentialsFileSetupError.ErrInvalidProviderId:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInvalidProviderId, error: e }
      default:
        throw e
    }
  }
}