import { Form, useActionData, useNavigate } from "react-router-dom"
import { useEffect } from "react"
import { useWails } from "../../../wails-provider/wails-context"
import {
  AwsCredentialsFileSetupError,
  AwsCredentialsFileSetupResult,
} from "./new-data"
import { useToaster } from "../../../toast-provider/toast-context"

export function AwsCredentialsFileNew() {
  const wails = useWails()
  const toaster = useToaster()
  const navigate = useNavigate()

  const filePath = "~/.aws/credentials"

  const setupResult = useActionData() as AwsCredentialsFileSetupResult | undefined

  useEffect(() => {
    if (setupResult) {
      if (setupResult.success) {
        navigate("../")
      } else {
        switch (setupResult.code) {
          case AwsCredentialsFileSetupError.ErrInvalidLabel:
            toaster.showError("The label must be between 1 and 50 characters")
            break
          case AwsCredentialsFileSetupError.ErrInstanceAlreadyRegistered:
            toaster.showWarning("credentials file path already registered")
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
        <h1 className="text-primary text-4xl">AWS Credentials File</h1>
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
