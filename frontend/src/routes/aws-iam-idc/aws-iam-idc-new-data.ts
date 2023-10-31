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

  await FinalizeSetup(3)

  return redirect("/dashboard")
}
