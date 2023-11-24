import { Form, useActionData, useLocation, useNavigate } from "react-router-dom"
import { ExternalLink } from "../../components/external-link"
import {
  AwsIamIdcDeviceAuthFlowError,
  AwsIamIdcDeviceAuthFlowResult,
} from "./aws-iam-idc-device-auth-data"
import { useEffect, useRef } from "react"
import { useToaster } from "../../toast-provider/toast-context"

export function AwsIamIdcDeviceAuth() {
  const toaster = useToaster()
  const location = useLocation()
  const navigate = useNavigate()
  const actionData = useActionData() as AwsIamIdcDeviceAuthFlowResult | undefined

  const authFlowState = useRef(location.state)

  if (
    location.state &&
    authFlowState.current.verificationUriComplete !==
      location.state.verificationUriComplete
  ) {
    authFlowState.current = location.state
  }

  const {
    action,
    instanceId,
    clientId,
    startUrl,
    label,
    awsRegion,
    verificationUriComplete,
    userCode,
    deviceCode,
  } = authFlowState.current

  useEffect(() => {
    if (!actionData) {
      return
    }

    if (actionData.success === true) {
      return navigate("/")
    }

    if (actionData.success === false) {
      switch (actionData.code) {
        case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowNotAuthorized:
          toaster.showError(
            "You haven not authorized the device through the activation link :(\nPlease do so then click this button again",
          )
          return
        case AwsIamIdcDeviceAuthFlowError.ErrDeviceAuthFlowTimedOut:
          toaster.showError(
            "The device authorization flow timed out and we have to start over",
          )
          return navigate("/")
        case AwsIamIdcDeviceAuthFlowError.ErrInvalidStartUrl:
          toaster.showError(
            "The Start URL is not a valid AWS IAM Identity Center URL",
          )
          return
        case AwsIamIdcDeviceAuthFlowError.ErrInvalidAwsRegion:
          toaster.showError("The AWS region is not valid")
          return
        case AwsIamIdcDeviceAuthFlowError.ErrInvalidLabel:
          toaster.showError("The account label must be between 1 and 50 characters")
          return
        case AwsIamIdcDeviceAuthFlowError.ErrTransientAwsClientError:
          toaster.showWarning(
            "There was an error, but it might work if you try again a bit later",
          )
          return
      }
    }
  }, [toaster, navigate, actionData])

  return (
    <Form
      method="post"
      className="h-screen flex flex-col items-center justify-center gap-4 border-2 p-6">
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
        name="instanceId"
        value={instanceId}
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
        name="label"
        value={label}
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
