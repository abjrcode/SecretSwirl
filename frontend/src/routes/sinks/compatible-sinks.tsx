import {
  Link,
  Outlet,
  useLoaderData,
  useOutlet,
  useSearchParams,
} from "react-router-dom"
import { main } from "../../../wailsjs/go/models"

export function CompatibleSinks() {
  const outlet = useOutlet()
  const [searchParams] = useSearchParams()
  const sinks = useLoaderData() as main.CompatibleSink[] | undefined

  if (!sinks) {
    return <>Loading ...</>
  }

  return (
    <>
      {outlet == null ? (
        <div className="flex flex-col gap-10">
          <h1 className="text-primary text-6xl">Compatible Sinks</h1>
          <ul className="flex gap-4">
            {sinks.map((sink) => (
              <li key={sink.code}>
                <Link
                  className="btn btn-primary"
                  to={`${sink.code}/setup/${searchParams.get(
                    "providerCode",
                  )}/${searchParams.get("providerId")}`}>
                  {sink.name}
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ) : null}
      <Outlet />
    </>
  )
}
