import { GetInstanceData } from '../../../wailsjs/go/awsiamidc/AwsIdentityCenterController'

export async function awsIamIdcCardLoader({ request }: { request: Request }) {
  const instanceId = new URL(request.url).searchParams.get('instanceId')

  if (!instanceId) {
    throw new Response('instanceId is required', { status: 400 })
  }

  return GetInstanceData(instanceId)
}