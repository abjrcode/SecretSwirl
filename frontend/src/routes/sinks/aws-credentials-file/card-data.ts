import { awscredssink } from '../../../../wailsjs/go/models'
import { ActionDataResult } from '../../../utils/action-data-result'
import { AwsCredentialsSink_GetInstanceData } from '../../../utils/ipc-adapter'

export enum AwsCredentialsFileCardDataError { }

export type AwsCredentialsFileCardDataResult = ActionDataResult<awscredssink.AwsCredentialsSinkInstance, AwsCredentialsFileCardDataError>

export async function awsCredentialsFileCardLoader({ request }: { request: Request }): Promise<AwsCredentialsFileCardDataResult> {
  const instanceId = new URL(request.url).searchParams.get('instanceId')

  if (!instanceId) {
    throw new Response('instanceId is required', { status: 400 })
  }

  try {
    return { success: true, result: await AwsCredentialsSink_GetInstanceData(instanceId) }
  }
  catch (error) {
    switch (error) {

      default:
        throw error
    }
  }
}