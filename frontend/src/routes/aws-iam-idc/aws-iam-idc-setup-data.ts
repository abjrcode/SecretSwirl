import { ActionFunctionArgs } from "react-router-dom"
import {
  Setup,
} from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"

export async function awsIamIdcSetupAction({ request }: ActionFunctionArgs) {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  return Setup(updates["startUrl"].toString(), updates.awsRegion.toString())
}
