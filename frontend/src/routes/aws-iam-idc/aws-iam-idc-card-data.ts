import { GetInstanceData } from '../../../wailsjs/go/awsiamidc/AwsIdentityCenterController'
import { awsiamidc } from '../../../wailsjs/go/models'
import { ActionDataResult } from '../../components/action-data-result'

export enum AwsIamIdcCardDataError {
  ErrAccessTokenExpired = "ACCESS_TOKEN_EXPIRED",
  ErrTransientAwsClientError = "TRANSIENT_AWS_CLIENT_ERROR",
}

export type AwsIamIdcCardDataResult = ActionDataResult<awsiamidc.AwsIdentityCenterCardData, AwsIamIdcCardDataError>

export async function awsIamIdcCardLoader({ request }: { request: Request }): Promise<AwsIamIdcCardDataResult> {
  const instanceId = new URL(request.url).searchParams.get('instanceId')
  const refresh = new URL(request.url).searchParams.get('refresh') === 'true'

  if (!instanceId) {
    throw new Response('instanceId is required', { status: 400 })
  }

  try {
    return { success: true, result: await GetInstanceData(instanceId, refresh) }
  }
  catch (error) {
    switch (error) {
      case AwsIamIdcCardDataError.ErrAccessTokenExpired:
        return { success: false, code: AwsIamIdcCardDataError.ErrAccessTokenExpired, error: error }
      case AwsIamIdcCardDataError.ErrTransientAwsClientError:
        return { success: false, code: AwsIamIdcCardDataError.ErrTransientAwsClientError, error: error }
      default:
        throw error
    }
  }
}