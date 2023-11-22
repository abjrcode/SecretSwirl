import { Link, useLoaderData } from "react-router-dom"
import { AwsIamIdcCard } from "../aws-iam-idc/aws-iam-idc-card"
import { main } from "../../../wailsjs/go/models"

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const ProviderComponentMap = new Map<string, React.FC<any>>([
  ["aws-iam-idc", AwsIamIdcCard],
])

export function Dashboard() {
  const favoriteInstances = useLoaderData() as main.FavoriteInstance[]

  return (
    <>
      {...favoriteInstances.map((favorite) => {
        const Component = ProviderComponentMap.get(favorite.providerCode)
        if (!Component) {
          throw new Error(
            `No component found for provider of type [${favorite.providerCode}] and ID [${favorite.instanceId}]`,
          )
        }
        return (
          <Component
            key={favorite.instanceId}
            {...favorite}
          />
        )
      })}
      <Link
        className="btn btn-primary"
        to="/providers">
        New
      </Link>
    </>
  )
}
