import { ActionFunctionArgs, redirect } from "react-router-dom"
import {
  FinalizeSetup,
  Setup,
} from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"

export async function awsIamIdcNewConfigureAction({ request }: ActionFunctionArgs) {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const requestedAction = updates["action"].toString()

  if (requestedAction === "configure") {
    return Setup(updates["startUrl"].toString(), updates.awsRegion.toString())
  }

  const clientId = updates["clientId"].toString()
  const startUrl = updates["startUrl"].toString()
  const awsRegion = updates["awsRegion"].toString()
  const userCode = updates["userCode"].toString()
  const deviceCode = updates["deviceCode"].toString()

  try {
    await FinalizeSetup(clientId, startUrl, awsRegion, userCode, deviceCode)
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
