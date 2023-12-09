import { ActionFunctionArgs } from "react-router-dom"
import { ActionDataResult } from "../../../utils/action-data-result"
import { AwsCredentialsFile_NewInstance } from "../../../utils/ipc-adapter"

export enum AwsCredentialsFileSetupError {
  ErrInstanceAlreadyRegistered = "INSTANCE_ALREADY_REGISTERED",
  ErrInvalidLabel = "INVALID_LABEL",
}

export type AwsCredentialsFileSetupResult = ActionDataResult<string, AwsCredentialsFileSetupError>

export async function awsCredentialsFileSetupAction({ request }: ActionFunctionArgs): Promise<AwsCredentialsFileSetupResult> {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const filePath = updates["filePath"].toString()
  const label = updates["label"].toString()

  try {
    return {
      success: true,
      result: await AwsCredentialsFile_NewInstance({
        filePath,
        label,
      })
    }
  } catch (e) {
    switch (e) {
      case AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered, error: e }
      case AwsCredentialsFileSetupError.ErrInvalidLabel:
        return { success: false, code: AwsCredentialsFileSetupError.ErrInvalidLabel, error: e }
      default:
        throw e
    }
  }
}