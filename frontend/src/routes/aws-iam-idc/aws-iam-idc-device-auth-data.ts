import { ActionFunctionArgs } from "react-router-dom"
import { FinalizeRefreshAccessToken, FinalizeSetup } from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"
import { ActionDataResult } from "../../components/action-data-result"

export enum AwsIamIdcDeviceAuthFlowError {
  ErrDeviceAuthFlowNotAuthorized = "DEVICE_AUTH_FLOW_NOT_AUTHORIZED",
  ErrDeviceAuthFlowTimedOut = "DEVICE_AUTH_FLOW_TIMED_OUT",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIamIdcDeviceAuthFlowResult = ActionDataResult<null, AwsIamIdcDeviceAuthFlowError>

export async function awsIamIdcDeviceAuthAction({ request }: ActionFunctionArgs): Promise<AwsIamIdcDeviceAuthFlowResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const action = updates["action"].toString()
  const instanceId = updates["instanceId"].toString()
  const clientId = updates["clientId"].toString()
  const startUrl = updates["startUrl"].toString()
  const awsRegion = updates["awsRegion"].toString()
  const userCode = updates["userCode"].toString()
  const deviceCode = updates["deviceCode"].toString()

  try {
    switch (action) {
      case "setup":
        await FinalizeSetup(clientId, startUrl, awsRegion, userCode, deviceCode)
        break;
      case "refresh":
        await FinalizeRefreshAccessToken(instanceId, awsRegion, userCode, deviceCode)
        break;
      default:
        throw new Error(`Unknown action: ${action}`)
    }
  } catch (e) {
    switch (e) {
      case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized, error: e }
      case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut, error: e }
      case AwsIamIdcDeviceAuthFlowError.ErrInvalidStartUrl:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrInvalidStartUrl, error: e }
      case AwsIamIdcDeviceAuthFlowError.ErrInvalidAwsRegion:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrInvalidAwsRegion, error: e }
      case AwsIamIdcDeviceAuthFlowError.ErrTransientAwsClientError:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }

  return { success: true, result: null }
}
