import { Form, useActionData, useNavigate, useParams } from "react-router-dom"
import { useEffect } from "react"
import {
  AwsCredentialsFileSetupError,
  AwsCredentialsFileSetupResult,
} from "./new-data"
import { useToaster } from "../../../toast-provider/toast-context"

export function AwsCredentialsFileSetup() {
  const toaster = useToaster()
  const navigate = useNavigate()
  const setupResult = useActionData() as AwsCredentialsFileSetupResult | undefined
  const params = useParams()
  const providerCode = params["providerCode"]
  const instanceId = params["instanceId"]

  const filePath = "~/.aws/credentials"

  useEffect(() => {
    if (setupResult) {
      if (setupResult.success) {
        navigate("/")
      } else {
        switch (setupResult.code) {
          case AwsCredentialsFileSetupError.ErrInvalidLabel:
            toaster.showError("label must be between 1 and 50 characters")
            break
          case AwsCredentialsFileSetupError.ErrInvalidAwsProfileName:
            toaster.showError("AWS profile name must be between 1 and 50 characters")
            break
          case AwsCredentialsFileSetupError.ErrInvalidProviderCode:
            toaster.showError("Provider code must always be specified")
            break
          case AwsCredentialsFileSetupError.ErrInvalidProviderId:
            toaster.showError("Provider id must always be specified")
            break
          case AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered:
            toaster.showWarning("credentials file path already registered")
            break
        }
      }
    }
  }, [navigate, setupResult, toaster])

  return (
    <div className="h-screen flex flex-col items-center justify-center">
      <Form
        method="post"
        className="flex flex-col gap-4 border-2 p-6">
        <h1 className="text-primary text-4xl">AWS Credentials File</h1>
        <input
          type="hidden"
          name="providerCode"
          value={providerCode}
        />
        <input
          type="hidden"
          name="providerId"
          value={instanceId}
        />
        <label className="label">
          <span className="label-text">Credentials File Path</span>
        </label>
        <input
          name="filePath"
          type="text"
          className="input input-bordered input-primary w-96"
          defaultValue={filePath}
        />
        <label className="label">
          <span className="label-text">Aws Profile Name</span>
        </label>
        <input
          name="awsProfileName"
          type="text"
          minLength={1}
          maxLength={50}
          className="input input-bordered input-primary w-96"
          placeholder="default"
        />
        <label className="label">
          <span className="label-text">Label</span>
        </label>
        <input
          name="label"
          type="text"
          minLength={1}
          maxLength={50}
          className="input input-bordered input-primary w-96"
          placeholder="Default Credentials File"
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
