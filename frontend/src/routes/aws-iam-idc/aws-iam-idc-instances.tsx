import { Link, useLoaderData } from "react-router-dom"
import { AwsIamIdcCard } from "./aws-iam-idc-card"

export function AwsIamIdcInstances() {
  const loader = useLoaderData() as string[] | undefined

  if (!loader) {
    return null
  }

  return (
    <div>
      <h1 className="text-primary text-3xl">AWS IAM Identity Center Instances</h1>

      <br />
      <br />

      <ul className="flex gap-4">
        {loader.map((instance) => (
          <li key={instance}>
            <AwsIamIdcCard instanceId={instance} />
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
