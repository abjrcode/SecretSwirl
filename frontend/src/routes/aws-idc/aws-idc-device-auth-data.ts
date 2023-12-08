import { ActionFunctionArgs } from "react-router-dom"
import { AwsIdc_FinalizeRefreshAccessToken, AwsIdc_FinalizeSetup } from "../../utils/ipc-adapter"
import { ActionDataResult } from "../../utils/action-data-result"

export enum AwsIdcDeviceAuthFlowError {
  ErrDeviceAuthFlowNotAuthorized = "DEVICE_AUTH_FLOW_NOT_AUTHORIZED",
  ErrDeviceAuthFlowTimedOut = "DEVICE_AUTH_FLOW_TIMED_OUT",
  ErrInvalidStartUrl = "INVALID_START_URL",
  ErrInvalidAwsRegion = "INVALID_AWS_REGION",
  ErrInvalidLabel = "INVALID_LABEL",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIdcDeviceAuthFlowResult = ActionDataResult<null, AwsIdcDeviceAuthFlowError>

export async function awsIdcDeviceAuthAction({ request }: ActionFunctionArgs): Promise<AwsIdcDeviceAuthFlowResult> {
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
        await AwsIdc_FinalizeSetup({
          startUrl,
          awsRegion,
          label,
          clientId,
          deviceCode,
          userCode,
        })
        break;
      case "refresh":
        await AwsIdc_FinalizeRefreshAccessToken({
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
      case AwsIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized, error: e }
      case AwsIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut, error: e }
      case AwsIdcDeviceAuthFlowError.ErrInvalidStartUrl:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrInvalidStartUrl, error: e }
      case AwsIdcDeviceAuthFlowError.ErrInvalidAwsRegion:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrInvalidAwsRegion, error: e }
      case AwsIdcDeviceAuthFlowError.ErrInvalidLabel:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrInvalidLabel, error: e }
      case AwsIdcDeviceAuthFlowError.ErrTransientAwsClientError:
        return { success: false, code: AwsIdcDeviceAuthFlowError.ErrTransientAwsClientError, error: e }
      default:
        throw e
    }
  }

  return { success: true, result: null }
}
