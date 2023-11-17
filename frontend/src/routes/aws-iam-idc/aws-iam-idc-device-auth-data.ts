import { ActionFunctionArgs } from "react-router-dom"
import { FinalizeRefreshAccessToken, FinalizeSetup } from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"

export enum AwsIamIdcDeviceAuthFlowError {
  ErrDeviceAuthFlowNotAuthorized = "DEVICE_AUTH_FLOW_NOT_AUTHORIZED",
  ErrDeviceAuthFlowTimedOut = "DEVICE_AUTH_FLOW_TIMED_OUT"
}

export type AwsIamIdcDeviceAuthFlowResult = {
  success: true
} | {
  success: false
  code: AwsIamIdcDeviceAuthFlowError
  actualError: unknown
}

export async function awsIamIdcDeviceAuthAction({ request }: ActionFunctionArgs): Promise<AwsIamIdcDeviceAuthFlowResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const action = updates["action"].toString()
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
        await FinalizeRefreshAccessToken(clientId, startUrl, awsRegion, userCode, deviceCode)
        break;
      default:
        throw new Error(`Unknown action: ${action}`)
    }
  } catch (e) {
    switch (e) {
      case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized, actualError: e }
      case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut:
        return { success: false, code: AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut, actualError: e }
      default:
        throw e
    }
  }

  return { success: true }
}
