import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom"

export function ErrorPage() {
  const error = useRouteError() as { statusText: string; message: string }
  console.error(error)

  if (isRouteErrorResponse(error)) {
    return (
      <div className="h-screen flex flex-col gap-4 items-center justify-center">
        <div>
          <h1 className="text-5xl">Oops!</h1>
          <h2>{error.status}</h2>
          <p>{error.statusText}</p>
          {error.data?.message && <p>{error.data.message}</p>}
        </div>
        <Link
          to="/"
          reloadDocument
          className="link link-primary">
          Go Home
        </Link>
      </div>
    )
  } else {
    return (
      <div className="h-screen flex flex-col gap-4 items-center justify-center">
        <h1 className="text-5xl">Oops</h1>
        <Link
          to="/"
          reloadDocument
          className="link link-primary">
          Go Home
        </Link>
      </div>
    )
  }
}
