import { Link, Outlet, useLoaderData, useOutlet } from "react-router-dom"
import { main } from "../../../wailsjs/go/models"

export function Providers() {
  const outlet = useOutlet()
  const providers = useLoaderData() as main.Provider[]

  return (
    <>
      {outlet == null ? (
        <div className="flex flex-col gap-10">
          <h1 className="text-primary text-6xl">Supported Providers</h1>
          <ul className="flex gap-4">
            {providers.map((provider) => (
              <li key={provider.code}>
                <Link
                  className="btn btn-primary"
                  to={`${provider.code}`}>
                  {provider.name}
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
