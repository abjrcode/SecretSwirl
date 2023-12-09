import React from "react"
import { useFetcher } from "react-router-dom"

import { AwsCredentialsFileCardDataResult } from "./card-data"

export function AwsCredentialsFileCard({ instanceId }: { instanceId: string }) {
  const fetcher = useFetcher()

  React.useEffect(() => {
    if (fetcher.state === "idle" && !fetcher.data) {
      const urlSearchParams = new URLSearchParams()
      urlSearchParams.append("instanceId", instanceId)
      fetcher.load(
        `/internal/api/aws-credentials-file-card?${urlSearchParams.toString()}`,
      )
    }
  }, [instanceId, fetcher])

  const cardDataResult = fetcher.data as AwsCredentialsFileCardDataResult | undefined

  if (cardDataResult === undefined) {
    return <div className="skeleton items-center w-48"></div>
  }

  if (cardDataResult.success) {
    const cardData = cardDataResult.result

    return (
      <div className="card gap-6 px-6 py-4 max-w-lg card-bordered border-secondary bg-base-200 drop-shadow-lg">
        <div
          role="heading"
          className="card-title justify-between">
          <div className="inline-flex items-center justify-center gap-2">
            <h1 className="text-2xl font-semibold">{cardData.label}</h1>
          </div>
        </div>
        <div className="card-body">{cardData.filePath}</div>
        <div className="card-actions items-center justify-between"></div>
      </div>
    )
  }
}
