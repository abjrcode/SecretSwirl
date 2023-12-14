import { AwsCredentialsFile } from "../routes/sinks/aws-credentials-file/card"
import { SinkCodes } from "../utils/provider-sink-codes"

export function createSinkCard(sinkCode: string, sinkId: string) {
  switch (sinkCode) {
    case SinkCodes.AwsCredentialsFile:
      return <AwsCredentialsFile instanceId={sinkId} />

    default:
      throw new Error("Unknown sinkCode: " + sinkCode)
  }
}
