import { ActionFunctionArgs } from "react-router-dom"
import { AwsIamIdc_FinalizeRefreshAccessToken, AwsIamIdc_FinalizeSetup } from "../../utils/ipc-adapter"
import { ActionDataResult } from "../../utils/action-data-result"

export enum AwsIamIdcDeviceAuthFlowError {
  ErrDeviceAuthFlowNotAuthorized = "DEVICE_AUTH_FLOW_NOT_AUTHORIZED",
  ErrDeviceAuthFlowTimedOut = "DEVICE_AUTH_FLOW_TIMED_OUT",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrInvalidLabel = "INVALID_LABEL",
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
  const label = updates["label"].toString()
  const userCode = updates["userCode"].toString()
  const deviceCode = updates["deviceCode"].toString()

  try {
    switch (action) {
      case "setup":
        await AwsIamIdc_FinalizeSetup({
          startUrl,
          awsRegion,
          label,
          clientId,
          deviceCode,
          userCode,
        })
        break;
      case "refresh":
        await AwsIamIdc_FinalizeRefreshAccessToken({
          instanceId,
          region: awsRegion,
          deviceCode,
          userCode,
        }
        )
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
      case AwsIamIdcDeviceAuthFlowError.ErrInvalidLabel:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrInvalidLabel, error: e }
      case AwsIamIdcDeviceAuthFlowError.ErrTransientAwsClientError:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }

  return { success: true, result: null }
}
