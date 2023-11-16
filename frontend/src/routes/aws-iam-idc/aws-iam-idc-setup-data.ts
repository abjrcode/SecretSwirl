import { ActionFunctionArgs } from "react-router-dom"
import {
  Setup,
} from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"
import { awsiamidc } from "../../../wailsjs/go/models"

enum AwsIamIdcSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
}

export type AwsIamIdcSetupResult = {
  success: true
  result: awsiamidc.AuthorizeDeviceFlowResult
} | {
  success: false
  error: string
}

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
        return { success: false, error: "The Start URL and region combination have already been added before" }
      default:
        throw e
    }
  }
}
