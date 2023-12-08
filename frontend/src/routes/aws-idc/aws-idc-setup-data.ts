import { ActionFunctionArgs } from "react-router-dom"
import { awsidc } from "../../../wailsjs/go/models"
import { ActionDataResult } from "../../utils/action-data-result"
import { AwsIdc_Setup } from "../../utils/ipc-adapter"

export enum AwsIdcSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrInvalidLabel = "INVALID_LABEL",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIdcSetupResult = ActionDataResult<awsidc.AuthorizeDeviceFlowResult, AwsIdcSetupError>

export async function awsIdcSetupAction({ request }: ActionFunctionArgs): Promise<AwsIdcSetupResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const startUrl = updates["startUrl"].toString()
  const awsRegion = updates["awsRegion"].toString()
  const label = updates["label"].toString()

  try {
    return {
      success: true,
      result: await AwsIdc_Setup({
        startUrl,
        awsRegion,
        label,
      })
    }
  } catch (e) {
    switch (e) {
      case AwsIdcSetupError.ErrInstanceAlreadyRegistered:
        return { success: false, code: AwsIdcSetupError.ErrInstanceAlreadyRegistered, error: e }
      case AwsIdcSetupError.ErrInvalidStartUrl:
        return { success: false, code: AwsIdcSetupError.ErrInvalidStartUrl, error: e }
      case AwsIdcSetupError.ErrInvalidAwsRegion:
        return { success: false, code: AwsIdcSetupError.ErrInvalidAwsRegion, error: e }
      case AwsIdcSetupError.ErrInvalidLabel:
        return { success: false, code: AwsIdcSetupError.ErrInvalidLabel, error: e }
      case AwsIdcSetupError.ErrTransientAwsClientError:
        return { success: false, code: AwsIdcSetupError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }
}