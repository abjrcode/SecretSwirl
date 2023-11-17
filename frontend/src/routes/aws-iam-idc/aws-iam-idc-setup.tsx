import { Form, useActionData, useNavigate } from "react-router-dom"
import { useEffect } from "react"
import { useWails } from "../../wails-provider/wails-context"
import { AwsIamIdcSetupError, AwsIamIdcSetupResult } from "./aws-iam-idc-setup-data"
import { useToaster } from "../../toast-provider/toast-context"

export function AwsIamIdcSetup() {
  const wails = useWails()
  const toaster = useToaster()
  const navigate = useNavigate()

  const startUrl = "https://my-app.awsapps.com/start"
  const awsRegion = "eu-central-1"

  const setupResult = useActionData() as AwsIamIdcSetupResult | undefined

  useEffect(() => {
    if (setupResult) {
      if (setupResult.success) {
        const result = setupResult.result
        navigate("../device-auth", {
          state: {
            action: "setup",
            clientId: result.clientId,
            startUrl: result.startUrl,
            awsRegion: result.region,
            verificationUriComplete: result.verificationUri,
            userCode: result.userCode,
            deviceCode: result.deviceCode,
          },
        })
      } else {
        switch (setupResult.code) {
          case AwsIamIdcSetupError.ErrInvalidStartUrl:
            toaster.showError("The Start URL is not valid")
            break
          case AwsIamIdcSetupError.ErrInvalidAwsRegion:
            toaster.showError("The AWS region is not valid")
            break
          case AwsIamIdcSetupError.ErrInstanceAlreadyRegistered:
            wails.runtime.ShowWarningDialog(
              "Start URL and Region combination are already registered",
            )
            break
          case AwsIamIdcSetupError.ErrTransientAwsClientError:
            toaster.showWarning(
              "There was an error, but it might work if you try again a bit later",
            )
            break
        }
      }
    }
  }, [navigate, setupResult, wails, toaster])

  return (
    <Form
      method="post"
      className="flex flex-col gap-4 border-2 p-6">
      <h1 className="text-primary text-4xl">AWS IAM IDC</h1>
      <label className="label">
        <span className="label-text">Start URL</span>
      </label>
      <input
        name="startUrl"
        type="url"
        className="input input-bordered input-primary w-96"
        defaultValue={startUrl}
      />
      <label className="label">
        <span className="label-text">AWS Region</span>
      </label>
      <select
        name="awsRegion"
        className="select select-primary"
        defaultValue={awsRegion}>
        <option value={awsRegion}>{awsRegion}</option>
      </select>
      <button
        type="submit"
        className="btn btn-primary">
        Configure
      </button>
      <button
        type="reset"
        onClick={() => navigate(-1)}
        className="btn btn-secondary">
        Cancel
      </button>
    </Form>
  )
}
