import { Form, useLocation, useNavigate } from "react-router-dom"
import { ExternalLink } from "../../components/external-link"

export function AwsIamIdcDeviceAuth() {
  const location = useLocation()
  const navigate = useNavigate()

  const {
    action,
    verificationUriComplete,
    clientId,
    startUrl,
    awsRegion,
    userCode,
    deviceCode,
  } = location.state

  return (
    <Form
      method="post"
      className="flex flex-col gap-4 border-2 p-6">
      <p>
        Please authorize the request by visiting{" "}
        <ExternalLink
          href={verificationUriComplete}
          text={verificationUriComplete}
        />
        . You have a total of 5 (five) minutes to do so!
      </p>
      <input
        type="hidden"
        name="action"
        value={action}
      />
      <input
        type="hidden"
        name="clientId"
        value={clientId}
      />
      <input
        type="hidden"
        name="startUrl"
        value={startUrl}
      />
      <input
        type="hidden"
        name="awsRegion"
        value={awsRegion}
      />
      <input
        type="hidden"
        name="userCode"
        value={userCode}
      />
      <input
        type="hidden"
        name="deviceCode"
        value={deviceCode}
      />
      <button
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
