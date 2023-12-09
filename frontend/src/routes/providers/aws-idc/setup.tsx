import { Form, useActionData, useNavigate } from "react-router-dom"
import { useEffect } from "react"
import { useWails } from "../../../wails-provider/wails-context"
import { AwsIdcSetupError, AwsIdcSetupResult } from "./setup-data"
import { useToaster } from "../../../toast-provider/toast-context"

export function AwsIdcSetup() {
  const wails = useWails()
  const toaster = useToaster()
  const navigate = useNavigate()

  const startUrl = "https://my-app.awsapps.com/start"
  const awsRegion = "eu-central-1"

  const setupResult = useActionData() as AwsIdcSetupResult | undefined

  useEffect(() => {
    if (setupResult) {
      if (setupResult.success) {
        const result = setupResult.result
        navigate("../device-auth", {
          state: {
            action: "setup",
            clientId: result.clientId,
            startUrl: result.startUrl,
            label: result.label,
            awsRegion: result.region,
            verificationUriComplete: result.verificationUri,
            userCode: result.userCode,
            deviceCode: result.deviceCode,
          },
        })
      } else {
        switch (setupResult.code) {
          case AwsIdcSetupError.ErrInvalidStartUrl:
            toaster.showError("The Start URL is not valid")
            break
          case AwsIdcSetupError.ErrInvalidAwsRegion:
            toaster.showError("The AWS region is not valid")
            break
          case AwsIdcSetupError.ErrInvalidLabel:
            toaster.showError(
              "The account label must be between 1 and 50 characters",
            )
            break
          case AwsIdcSetupError.ErrInstanceAlreadyRegistered:
            toaster.showWarning("AWS Identity Center already exists")
            break
          case AwsIdcSetupError.ErrTransientAwsClientError:
            toaster.showWarning(
              "There was an error, but it might work if you try again a bit later",
            )
            break
        }
      }
    }
  }, [navigate, setupResult, wails, toaster])

  return (
    <div className="h-screen flex flex-col items-center justify-center">
      <Form
        method="post"
        className="flex flex-col gap-4 border-2 p-6">
        <h1 className="text-primary text-4xl">AWS Identity Center</h1>
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
        <label className="label">
          <span className="label-text">Label</span>
        </label>
        <input
          name="label"
          type="text"
          minLength={1}
          maxLength={50}
          className="input input-bordered input-primary w-96"
          placeholder="Personal AWS Account"
        />
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
    </div>
  )
}
