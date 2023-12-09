import { Link, useLoaderData } from "react-router-dom"
import { AwsCredentialsFileCard } from "./card"

export function AwsCredentialsFileInstances() {
  const loader = useLoaderData() as string[] | undefined

  if (!loader) {
    return null
  }

  return (
    <div>
      <h1 className="text-primary text-3xl">AWS Credentials File Instances</h1>

      <br />
      <br />

      <ul className="flex gap-4">
        {loader.map((instance) => (
          <li key={instance}>
            <AwsCredentialsFileCard instanceId={instance} />
          </li>
        ))}

        <li>
          <Link
            to="./setup"
            className="btn btn-primary btn-outline">
            Add Instance
          </Link>
        </li>
      </ul>
    </div>
  )
}
