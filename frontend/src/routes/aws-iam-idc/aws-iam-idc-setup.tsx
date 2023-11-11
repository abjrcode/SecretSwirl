import { Form, useActionData, useNavigate } from "react-router-dom"
import { awsiamidc } from "../../../wailsjs/go/models"
import { useEffect } from "react"

export function AwsIamIdcSetup() {
  const navigate = useNavigate()

  const startUrl = "https://my-app.awsapps.com/start"
  const awsRegion = "eu-central-1"

  const deviceAuthFlowResult = useActionData() as
    | awsiamidc.AuthorizeDeviceFlowResult
    | undefined

  useEffect(() => {
    if (deviceAuthFlowResult) {
      navigate("../device-auth", {
        state: {
          action: "setup",
          clientId: deviceAuthFlowResult.clientId,
          startUrl: deviceAuthFlowResult.startUrl,
          awsRegion: deviceAuthFlowResult.region,
          verificationUriComplete: deviceAuthFlowResult.verificationUri,
          userCode: deviceAuthFlowResult.userCode,
          deviceCode: deviceAuthFlowResult.deviceCode,
        },
      })
    }
  }, [navigate, deviceAuthFlowResult])

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
