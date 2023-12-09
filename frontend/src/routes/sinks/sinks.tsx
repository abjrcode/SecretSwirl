import { Link, Outlet, useLoaderData, useOutlet } from "react-router-dom"
import { main } from "../../../wailsjs/go/models"

export function Sinks() {
  const outlet = useOutlet()
  const sinks = useLoaderData() as main.Sink[]

  return (
    <>
      {outlet == null ? (
        <div className="flex flex-col gap-10">
          <h1 className="text-primary text-6xl">Supported Sinks</h1>
          <ul className="flex gap-4">
            {sinks.map((sink) => (
              <li key={sink.code}>
                <Link
                  className="btn btn-primary"
                  to={`${sink.code}`}>
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
