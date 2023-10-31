import { Link, isRouteErrorResponse, useRouteError } from "react-router-dom"

export function ErrorPage() {
  const error = useRouteError() as { statusText: string; message: string }
  console.error(error)

  if (isRouteErrorResponse(error)) {
    return (
      <>
        <div>
          <h1>Oops!</h1>
          <h2>{error.status}</h2>
          <p>{error.statusText}</p>
          {error.data?.message && <p>{error.data.message}</p>}
        </div>
        <Link
          to="/dashboard"
          className="link link-primary">
          Go Home
        </Link>
      </>
    )
  } else {
    return (
      <>
        <div>Oops</div>
        <Link
          to="/dashboard"
          className="link link-primary">
          Go Home
        </Link>
      </>
    )
  }
}
