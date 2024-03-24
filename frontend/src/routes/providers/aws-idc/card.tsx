import React from "react"
import { useFetcher, useNavigate, useRevalidator } from "react-router-dom"

import { AwsIdcCardDataError, AwsIdcCardDataResult } from "./card-data"
import { useToaster } from "../../../toast-provider/toast-context"
import {
  AwsIdc_CopyRoleCredentials,
  AwsIdc_MarkAsFavorite,
  AwsIdc_RefreshAccessToken,
  AwsIdc_SaveRoleCredentials,
  AwsIdc_UnmarkAsFavorite,
  Plumbing_DisconnectSink,
} from "../../../utils/ipc-adapter"
import { createSinkCard } from "../../../components/sink-component-map"
import { ProviderCodes } from "../../../utils/provider-sink-codes"
import { Modal } from "../../../components/modal"

export function AwsIdcCard({ instanceId }: { instanceId: string }) {
  const toaster = useToaster()
  const navigate = useNavigate()
  const fetcher = useFetcher()
  const validator = useRevalidator()

  const [isModalOpen, setIsModalOpen] = React.useState(false)

  const [awsProfile, setAwsProfile] = React.useState("default")

  const selectedAwsAccountRole = React.useRef<{
    accountId: string
    roleName: string
  } | null>(null)

  async function authorizeDevice(instanceId: string) {
    const deviceAuthFlowResult = await AwsIdc_RefreshAccessToken(instanceId)

    navigate(`/providers/${ProviderCodes.AwsIdc}/device-auth`, {
      state: {
        action: "refresh",
        instanceId: deviceAuthFlowResult.instanceId,
        clientId: deviceAuthFlowResult.clientId,
        startUrl: deviceAuthFlowResult.startUrl,
        awsRegion: deviceAuthFlowResult.region,
        label: deviceAuthFlowResult.label,
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
      urlSearchParams.append("refresh", "false")
      fetcher.load(
        `/internal/api/${ProviderCodes.AwsIdc}-card?${urlSearchParams.toString()}`,
      )
    }
  }, [instanceId, fetcher])

  function forceRefresh() {
    const urlSearchParams = new URLSearchParams()
    urlSearchParams.append("instanceId", instanceId)
    urlSearchParams.append("refresh", "true")

    fetcher.load(
      `/internal/api/${ProviderCodes.AwsIdc}-card?${urlSearchParams.toString()}`,
    )
  }

  async function markAsFavorite() {
    await AwsIdc_MarkAsFavorite(instanceId)
    validator.revalidate()
  }

  async function unmarkAsFavorite() {
    await AwsIdc_UnmarkAsFavorite(instanceId)
    validator.revalidate()
  }

  async function copyCredentials(
    instanceId: string,
    accountId: string,
    roleName: string,
  ) {
    try {
      await AwsIdc_CopyRoleCredentials({
        instanceId,
        accountId,
        roleName,
      })

      toaster.showSuccess("Credentials copied to clipboard!")
    } catch (e) {
      switch (e) {
        case "STALE_AWS_ACCESS_TOKEN":
          toaster.showWarning(
            "Your AWS access token has gone stale because you signed out through the AWS Console! Reloading ...",
          )
          validator.revalidate()
          return
        default:
          toaster.showError("Something went wrong! Please try again.")
      }
    }
  }

  function openSaveDialog(accountId: string, roleName: string) {
    selectedAwsAccountRole.current = { accountId, roleName }
    setIsModalOpen(true)
  }

  function closeSaveDialog() {
    selectedAwsAccountRole.current = null
    setIsModalOpen(false)
  }

  async function saveAsProfile() {
    if (!selectedAwsAccountRole.current) {
      return
    }

    try {
      await AwsIdc_SaveRoleCredentials({
        instanceId,
        accountId: selectedAwsAccountRole.current.accountId,
        roleName: selectedAwsAccountRole.current.roleName,
        awsProfile,
      })
    } catch (e) {
      switch (e) {
        case "INVALID_CREDENTIALS_FILE":
        case "EMPTY_PROFILE":
        case "EMPTY_KEY":
        case "EMPTY_KEY_VALUE":
          toaster.showError(
            "Your AWS credentials file could not be parsed because it is not valid!",
          )
          return
        case "STALE_AWS_ACCESS_TOKEN":
          toaster.showWarning(
            "Your AWS access token has gone stale because you signed out through the AWS Console! Reloading ...",
          )
          validator.revalidate()
          return
        default:
          toaster.showError("Something went wrong! Please try again.")
          return
      }
    } finally {
      closeSaveDialog()
    }

    toaster.showSuccess("Credentials saved successfully!")
    closeSaveDialog()
  }

  async function disconnectPipe(sinkCode: string, sinkId: string) {
    await Plumbing_DisconnectSink({ sinkCode, sinkId })
    fetcher.data = undefined
    validator.revalidate()
  }

  const cardDataResult = fetcher.data as AwsIdcCardDataResult | undefined

  if (cardDataResult === undefined) {
    return (
      <div className="flex flex-col gap-4 items-center w-96">
        <div className="skeleton h-4 w-full"></div>
        <div className="skeleton h-48 w-5/6"></div>
        <div className="skeleton h-4 w-full"></div>
      </div>
    )
  }

  if (cardDataResult.success) {
    const cardData = cardDataResult.result

    return (
      <>
        <div className="card gap-6 px-6 py-4 max-w-lg card-bordered border-secondary bg-base-200 drop-shadow-lg">
          <div
            role="heading"
            className="card-title justify-between">
            <div className="inline-flex items-center justify-center gap-2">
              <button
                onClick={cardData.isFavorite ? unmarkAsFavorite : markAsFavorite}>
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill={cardData.isFavorite ? "currentColor" : "none"}
                  viewBox="0 0 24 24"
                  strokeWidth={1.5}
                  stroke="currentColor"
                  className="w-6 h-6">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M11.48 3.499a.562.562 0 011.04 0l2.125 5.111a.563.563 0 00.475.345l5.518.442c.499.04.701.663.321.988l-4.204 3.602a.563.563 0 00-.182.557l1.285 5.385a.562.562 0 01-.84.61l-4.725-2.885a.563.563 0 00-.586 0L6.982 20.54a.562.562 0 01-.84-.61l1.285-5.386a.562.562 0 00-.182-.557l-4.204-3.602a.563.563 0 01.321-.988l5.518-.442a.563.563 0 00.475-.345L11.48 3.5z"
                  />
                </svg>
              </button>
              <h1 className="text-2xl font-semibold">{cardData.label}</h1>
            </div>
            <button
              className={fetcher.state !== "idle" ? "animate-spin" : ""}
              onClick={forceRefresh}>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                strokeWidth={1.5}
                stroke="currentColor"
                className="w-6 h-6">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"
                />
              </svg>
            </button>
          </div>
          <div className="card-body">
            {!cardData.isAccessTokenExpired && (
              <>
                <h2 className="text-xl font-semibold">Accounts</h2>
                <ul className="list-disc pl-4 space-y-4">
                  {cardData.accounts.map((account) => (
                    <li key={account.accountId}>
                      <h3 className="text-lg">
                        {account.accountName} ({account.accountId})
                      </h3>
                      <ul className="list-inside space-y-2 list-disc pl-4">
                        {account.roles.map((role) => (
                          <li
                            key={role.roleName}
                            className="inline-flex items-center gap-2">
                            <span>{role.roleName}</span>
                            <div className="inline-flex gap-2">
                              <button
                                onClick={() =>
                                  copyCredentials(
                                    instanceId,
                                    account.accountId,
                                    role.roleName,
                                  )
                                }
                                className="btn btn-accent btn-xs uppercase">
                                copy credentials
                              </button>
                              <button
                                onClick={() =>
                                  openSaveDialog(account.accountId, role.roleName)
                                }
                                className="btn btn-accent btn-xs uppercase">
                                save
                              </button>
                            </div>
                          </li>
                        ))}
                      </ul>
                    </li>
                  ))}
                </ul>
              </>
            )}

            {cardData.isAccessTokenExpired && (
              <>
                <h2 className="text-xl">Access Token Has Expired</h2>
                <p>Please renew it by authorizing the device again.</p>
              </>
            )}

            {cardData.sinks.length > 0 && (
              <>
                <div className="divider"></div>
                <h2 className="text-xl font-semibold">Connections</h2>
                <ul className="list-none pl-4 space-y-4">
                  {cardData.sinks.map((sink) => (
                    <li key={sink.sinkId}>
                      <div className="flex gap-4">
                        <div className="flex-1">
                          {createSinkCard(sink.sinkCode, sink.sinkId)}
                        </div>
                        <button
                          className="font-bold"
                          onClick={() => disconnectPipe(sink.sinkCode, sink.sinkId)}>
                          X
                        </button>
                      </div>
                    </li>
                  ))}
                </ul>
              </>
            )}
          </div>

          <div className="card-actions items-center justify-between">
            {cardData.isAccessTokenExpired && (
              <div className="flex flex-col gap-2">
                <button
                  className="btn btn-primary"
                  onClick={() => authorizeDevice(instanceId)}>
                  Renew
                </button>
              </div>
            )}
            <div className="flex flex-col gap-2">
              <p className="badge badge-outline">
                {!cardData.isAccessTokenExpired
                  ? `Expires in ${cardData.accessTokenExpiresIn}`
                  : cardData.accessTokenExpiresIn !== "stale"
                  ? `Token expired ${cardData.accessTokenExpiresIn}`
                  : "Stale"}
              </p>
            </div>
          </div>
        </div>

        <Modal
          isOpen={isModalOpen}
          onClose={closeSaveDialog}
          closedByClickOutside={true}>
          <div className="flex flex-col gap-4 p-6">
            <div>
              <label className="label">
                <span className="label-text">AWS Profile</span>
              </label>
              <input
                name="awsProfile"
                type="text"
                className="input input-bordered input-primary w-full"
                onChange={(e) => setAwsProfile(e.target.value)}
                value={awsProfile}
              />
            </div>
            <button
              onClick={saveAsProfile}
              className="btn btn-primary">
              Save
            </button>
            <button
              onClick={closeSaveDialog}
              className="btn btn-secondary">
              Cancel
            </button>
          </div>
        </Modal>
      </>
    )
  }

  switch (cardDataResult.code) {
    case AwsIdcCardDataError.ErrTransientAwsClientError:
      return (
        <div className="card gap-6 px-6 py-4 card-bordered border-secondary bg-base-200 drop-shadow-lg">
          <div
            role="heading"
            className="card-title justify-between">
            <h1 className="text-2xl font-semibold">AWS Identity Center</h1>
          </div>
          <div className="card-body">
            <h2 className="text-xl">
              There was a temporary error, try refreshing the card data after a
              while!
            </h2>
          </div>
          <div className="card-actions justify-between"></div>
        </div>
      )
  }
}
