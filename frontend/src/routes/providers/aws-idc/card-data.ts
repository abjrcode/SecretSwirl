import { awsidc } from '../../../../wailsjs/go/models'
import { ActionDataResult } from '../../../utils/action-data-result'
import { AwsIdc_GetInstanceData } from '../../../utils/ipc-adapter'

export enum AwsIdcCardDataError {
  ErrAccessTokenExpired = "ACCESS_TOKEN_EXPIRED",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIdcCardDataResult = ActionDataResult<awsidc.AwsIdentityCenterCardData, AwsIdcCardDataError>

export async function awsIdcCardLoader({ request }: { request: Request }): Promise<AwsIdcCardDataResult> {
  const instanceId = new URL(request.url).searchParams.get('instanceId')
  const refresh = new URL(request.url).searchParams.get('refresh') === 'true'

  if (!instanceId) {
    throw new Response('instanceId is required', { status: 400 })
  }

  try {
    return { success: true, result: await AwsIdc_GetInstanceData(instanceId, refresh) }
  }
  catch (error) {
    switch (error) {
      case AwsIdcCardDataError.ErrAccessTokenExpired:
        return { success: false, code: AwsIdcCardDataError.ErrAccessTokenExpired, error: error }
      case AwsIdcCardDataError.ErrTransientAwsClientError:
        return { success: false, code: AwsIdcCardDataError.ErrTransientAwsClientError, error: error }
      default:
        throw error
    }
  }
}