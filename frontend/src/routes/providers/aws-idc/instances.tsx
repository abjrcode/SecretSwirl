import { Link, useLoaderData } from "react-router-dom"
import { AwsIdcCard } from "./card"

export function AwsIdcInstances() {
  const loader = useLoaderData() as string[] | undefined

  if (!loader) {
    return null
  }

  return (
    <div>
      <h1 className="text-primary text-3xl">AWS Identity Center Instances</h1>

      <br />
      <br />

      <ul className="flex gap-4">
        {loader.map((instance) => (
          <li key={instance}>
            <AwsIdcCard instanceId={instance} />
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
