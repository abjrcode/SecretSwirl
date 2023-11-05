import { Form, Link, useLoaderData } from "react-router-dom"
import { AwsIamIdcCard } from "../aws-iam-idc/aws-iam-idc-card"
import { main } from "../../../wailsjs/go/models"

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const ProviderComponentMap = new Map<string, React.FC<any>>([
  ["aws-iam-idc", AwsIamIdcCard],
])

export function Dashboard() {
  const favoriteProviders = useLoaderData() as main.ConfiguredProvider[]
  return (
    <>
      {...favoriteProviders.map((provider) => {
        const Component = ProviderComponentMap.get(provider.code)
        if (!Component) {
          throw new Error(`No component found for provider [${provider}]`)
        }
        return (
          <Component
            key={provider.instanceId}
            {...provider}
          />
        )
      })}
      <div>
        <Link
          className="btn btn-primary"
          to="/providers/new">
          New
        </Link>
      </div>
      <Form method="post">
        <button className="fixed top-5 right-5 btn btn-secondary btn-outline">
          lock
        </button>
      </Form>
    </>
  )
}
