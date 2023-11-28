import { ActionFunctionArgs } from "react-router-dom"
import { awsiamidc } from "../../../wailsjs/go/models"
import { ActionDataResult } from "../../utils/action-data-result"
import { AwsIamIdc_Setup } from "../../utils/ipc-adapter"

export enum AwsIamIdcSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrInvalidLabel = "INVALID_LABEL",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIamIdcSetupResult = ActionDataResult<awsiamidc.AuthorizeDeviceFlowResult, AwsIamIdcSetupError>

export async function awsIamIdcSetupAction({ request }: ActionFunctionArgs): Promise<AwsIamIdcSetupResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const startUrl = updates["startUrl"].toString()
  const awsRegion = updates["awsRegion"].toString()
  const label = updates["label"].toString()

  try {
    return {
      success: true,
      result: await AwsIamIdc_Setup({
        startUrl,
        awsRegion,
        label,
      })
    }
  } catch (e) {
    switch (e) {
      case AwsIamIdcSetupError.ErrInstanceAlreadyRegistered:
        return { success: false, code: AwsIamIdcSetupError.ErrInstanceAlreadyRegistered, error: e }
      case AwsIamIdcSetupError.ErrInvalidStartUrl:
        return { success: false, code: AwsIamIdcSetupError.ErrInvalidStartUrl, error: e }
      case AwsIamIdcSetupError.ErrInvalidAwsRegion:
        return { success: false, code: AwsIamIdcSetupError.ErrInvalidAwsRegion, error: e }
      case AwsIamIdcSetupError.ErrInvalidLabel:
        return { success: false, code: AwsIamIdcSetupError.ErrInvalidLabel, error: e }
      case AwsIamIdcSetupError.ErrTransientAwsClientError:
        return { success: false, code: AwsIamIdcSetupError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }
}