import { Link, Outlet, useLoaderData, useOutlet } from "react-router-dom"
import { main } from "../../../wailsjs/go/models"

export function ProvidersNew() {
  const outlet = useOutlet()
  const providers = useLoaderData() as main.Provider[]

  return (
    <>
      {outlet == null ? (
        <>
          <h1 className="text-primary text-6xl">Supported Providers</h1>
          <ul>
            {providers.map((provider) => (
              <li key={provider.code}>
                <Link
                  className="btn btn-primary"
                  to={provider.code}>
                  {provider.name}
                </Link>
              </li>
            ))}
          </ul>
          <Link
            to="/"
            className="link link-secondary">
            &#8592; dashboard
          </Link>
        </>
      ) : null}
      <Outlet />
    </>
  )
}
