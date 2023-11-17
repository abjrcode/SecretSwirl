import { ActionFunctionArgs } from "react-router-dom"
import {
  Setup,
} from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"
import { awsiamidc } from "../../../wailsjs/go/models"
import { ActionDataResult } from "../../components/action-data-result"

export enum AwsIamIdcSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIamIdcSetupResult = ActionDataResult<awsiamidc.AuthorizeDeviceFlowResult, AwsIamIdcSetupError>

export async function awsIamIdcSetupAction({ request }: ActionFunctionArgs): Promise<AwsIamIdcSetupResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  try {
    return {
      success: true,
      result: await Setup(updates["startUrl"].toString(), updates.awsRegion.toString())
    }
  } catch (e) {
    switch (e) {
      case AwsIamIdcSetupError.ErrInstanceAlreadyRegistered:
        return { success: false, code: AwsIamIdcSetupError.ErrInstanceAlreadyRegistered, error: e }
      case AwsIamIdcSetupError.ErrInvalidStartUrl:
        return { success: false, code: AwsIamIdcSetupError.ErrInvalidStartUrl, error: e }
      case AwsIamIdcSetupError.ErrInvalidAwsRegion:
        return { success: false, code: AwsIamIdcSetupError.ErrInvalidAwsRegion, error: e }
      case AwsIamIdcSetupError.ErrTransientAwsClientError:
        return { success: false, code: AwsIamIdcSetupError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }
}