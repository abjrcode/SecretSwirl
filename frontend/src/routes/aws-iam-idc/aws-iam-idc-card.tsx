import React from "react"
import { useFetcher, useNavigate } from "react-router-dom"
import { RefreshAccessToken } from "../../../wailsjs/go/awsiamidc/AwsIdentityCenterController"
import { awsiamidc } from "../../../wailsjs/go/models"

export function AwsIamIdcCard({
  instanceId,
  displayName,
}: {
  instanceId: string
  displayName: string
}) {
  const navigate = useNavigate()
  const fetcher = useFetcher()

  async function authorizeDevice(instanceId: string) {
    const deviceAuthFlowResult = await RefreshAccessToken(instanceId)

    navigate("/providers/aws-iam-idc/device-auth", {
      state: {
        action: "refresh",
        clientId: deviceAuthFlowResult.clientId,
        startUrl: deviceAuthFlowResult.startUrl,
        awsRegion: deviceAuthFlowResult.region,
        verificationUriComplete: deviceAuthFlowResult.verificationUri,
        userCode: deviceAuthFlowResult.userCode,
        deviceCode: deviceAuthFlowResult.deviceCode,
      },
    })
  }

  React.useEffect(() => {
    if (fetcher.state === "idle" && !fetcher.data) {
      const urlSearchParams = new URLSearchParams()
      urlSearchParams.append("instanceId", instanceId)
      fetcher.load(`/internal/api/aws-iam-idc-card?${urlSearchParams.toString()}`)
    }
  }, [instanceId, fetcher])

  const cardData = fetcher.data as
    | awsiamidc.AwsIdentityCenterCardData
    | string
    | undefined

  if (cardData === undefined) {
    return <div>Loading...</div>
  }

  if (typeof cardData === "string") {
    if (cardData === "ACCESS_TOKEN_EXPIRED") {
      return (
        <div className="card gap-6 px-6 py-4 card-bordered border-secondary bg-base-200 drop-shadow-lg">
          <div
            role="heading"
            className="card-title justify-between">
            <h1 className="text-2xl font-semibold">{displayName}</h1>
          </div>
          <div className="card-body">
            <h2 className="text-xl">Access Token Expired</h2>
          </div>
          <div className="card-actions justify-between">
            <div className="flex items-center gap-4">
              <button
                className="btn btn-primary"
                onClick={async () => await authorizeDevice(instanceId)}>
                Get new token
              </button>
            </div>
          </div>
        </div>
      )
    }

    throw new Error(`Unexpected fetcher.data: ${cardData}`)
  }

  return (
    <div className="card gap-6 px-6 py-4 card-bordered border-secondary bg-base-200 drop-shadow-lg">
      <div
        role="heading"
        className="card-title justify-between">
        <h1 className="text-2xl font-semibold">{displayName}</h1>
      </div>
      <div className="card-body">
        <h2 className="text-xl">Accounts</h2>
        <ul className="list-disc pl-4 space-y-4">
          {cardData.accounts.map((account) => (
            <li key={account.accountId}>
              <h3 className="text-lg">
                {account.accountName} ({account.accountId})
              </h3>
              <ul className="list-inside space-y-2 list-disc pl-4">
                <li>
                  <span>Admin</span>
                  <div className="inline-flex gap-2">
                    <button className="btn btn-secondary btn-outline btn-xs">
                      copy credentials
                    </button>
                    <a className="link link-secondary"> Management console </a>
                  </div>
                </li>
                <li>
                  <span>Viewer</span>
                  <div className="inline-flex gap-2">
                    <button className="btn btn-secondary btn-outline btn-xs">
                      copy credentials
                    </button>
                    <a className="link link-secondary"> Management console </a>
                  </div>
                </li>
              </ul>
            </li>
          ))}
        </ul>
      </div>
      <div className="card-actions items-center justify-between">
        <div className="flex flex-col gap-2">
          <p className="badge badge-outline">
            Expires In: {cardData.accessTokenExpiresIn}
          </p>
        </div>
      </div>
    </div>
  )
}
