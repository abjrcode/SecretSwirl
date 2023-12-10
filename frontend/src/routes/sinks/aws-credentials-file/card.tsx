import React from "react"
import { useFetcher } from "react-router-dom"

import { AwsCredentialsFileCardDataResult } from "./card-data"
import { SinkCodes } from "../../../utils/provider-sink-codes"

export function AwsCredentialsFile({ instanceId }: { instanceId: string }) {
  const fetcher = useFetcher()

  React.useEffect(() => {
    if (fetcher.state === "idle" && !fetcher.data) {
      const urlSearchParams = new URLSearchParams()
      urlSearchParams.append("instanceId", instanceId)
      fetcher.load(
        `/internal/api/${
          SinkCodes.AwsCredentialsFile
        }-card?${urlSearchParams.toString()}`,
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
      <div className="border-black border-[1px] rounded p-1 bg-secondary-content shadow-md">
        <h1 className="font-medium">{cardData.label}</h1>
        <div className="text-sm pl-4">{cardData.filePath}</div>
      </div>
    )
  }
}
