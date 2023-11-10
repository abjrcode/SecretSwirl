import { Form, useActionData, useNavigate } from "react-router-dom"
import { ExternalLink } from "../../components/external-link"
import { awsiamidc } from "../../../wailsjs/go/models"

export function AwsIamIdcSetup() {
  const navigate = useNavigate()

  const startUrl = "https://d-99670c0d3d.awsapps.com/start"
  const awsRegion = "eu-central-1"

  const deviceAuthFlowResult = useActionData() as
    | awsiamidc.AuthorizeDeviceFlowResult
    | undefined

  if (deviceAuthFlowResult) {
    return (
      <Form
        method="post"
        className="flex flex-col gap-4 border-2 p-6">
        <p>
          Please authorize the request by visiting
          <ExternalLink
            href={deviceAuthFlowResult.verificationUri}
            text={deviceAuthFlowResult.verificationUri}
          />
          . You have a total of 5 (five) minutes to do so!
        </p>
        <input
          type="hidden"
          name="clientId"
          value={deviceAuthFlowResult.clientId}
        />
        <input
          type="hidden"
          name="startUrl"
          value={deviceAuthFlowResult.startUrl}
        />
        <input
          type="hidden"
          name="awsRegion"
          value={deviceAuthFlowResult.region}
        />
        <input
          type="hidden"
          name="userCode"
          value={deviceAuthFlowResult.userCode}
        />
        <input
          type="hidden"
          name="deviceCode"
          value={deviceAuthFlowResult.deviceCode}
        />
        <button
          name="action"
          value="activate"
          type="submit"
          className="btn btn-primary">
          Activate
        </button>
        <button
          type="reset"
          onClick={() => navigate("/")}
          className="btn btn-secondary">
          Abort
        </button>
      </Form>
    )
  }

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
        name="action"
        value="configure"
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
