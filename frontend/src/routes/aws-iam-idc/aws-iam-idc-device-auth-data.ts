import { ActionFunctionArgs, redirect } from "react-router-dom"
import { FinalizeRefreshAccessToken, FinalizeSetup } from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"

export async function awsIamIdcDeviceAuthAction({ request }: ActionFunctionArgs) {
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
      case "DEVICE_AUTH_FLOW_TIMED_OUT":
        return redirect("new")
      default:
        throw e
    }
  }

  return redirect("/")
}
