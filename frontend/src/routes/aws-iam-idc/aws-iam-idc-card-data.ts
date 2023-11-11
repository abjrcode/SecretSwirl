import { GetInstanceData } from '../../../wailsjs/go/awsiamidc/AwsIdentityCenterController'

export async function awsIamIdcCardLoader({ request }: { request: Request }) {
  const instanceId = new URL(request.url).searchParams.get('instanceId')

  if (!instanceId) {
    throw new Response('instanceId is required', { status: 400 })
  }

  try {
    return await GetInstanceData(instanceId)
  }
  catch (error) {
    if (error === "ACCESS_TOKEN_EXPIRED"
      || error === "CLIENT_EXPIRED"
      || error === "CLIENT_ALREADY_REGISTERED") {
      return error
    }

    throw error
  }
}