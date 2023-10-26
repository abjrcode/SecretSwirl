import React from "react"
import ReactDOM from "react-dom/client"
import { createHashRouter, RouterProvider } from "react-router-dom"
import "./tailwind.css"
import { ErrorPage } from "./error-page"
import { Dashboard } from "./routes/dashboard"

if (import.meta.env.DEV) {
  document.documentElement.classList.add("debug-screens")
}

const router = createHashRouter([
  {
    path: "/",
    element: <Dashboard />,
    errorElement: <ErrorPage />,
  },
])

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
)
