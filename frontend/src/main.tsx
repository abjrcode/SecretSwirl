import "./tailwind.css"
import React from "react"
import ReactDOM from "react-dom/client"
import { createHashRouter, RouterProvider } from "react-router-dom"
import { ErrorPage } from "./error-page"
import { Dashboard } from "./routes/dashboard/dashboard"
import {
  dashboardAction,
  loader as dashboardLoader,
} from "./routes/dashboard/dashboard-data"
import { loader as providersNewLoader } from "./routes/providers/providers-new-data"
import { awsIamIdcCardLoader } from "./routes/aws-iam-idc/aws-iam-idc-card-data"
import { ProvidersNew } from "./routes/providers/providers-new"
import { AwsIamIdcNew } from "./routes/aws-iam-idc/aws-iam-idc-new"
import { awsIamIdcNewConfigureAction } from "./routes/aws-iam-idc/aws-iam-idc-new-data"
import { MasterPassword } from "./routes/master-password/master-password"
import {
  isMasterPasswordSetup,
  setupMasterPasswordOrLogin,
} from "./routes/master-password/master-password-data"

if (import.meta.env.DEV) {
  document.documentElement.classList.add("debug-screens")
}

const router = createHashRouter([
  {
    path: "/",
    element: <MasterPassword />,
    errorElement: <ErrorPage />,
    loader: isMasterPasswordSetup,
    action: setupMasterPasswordOrLogin,
  },
  {
    path: "/dashboard",
    element: <Dashboard />,
    errorElement: <ErrorPage />,
    loader: dashboardLoader,
    action: dashboardAction,
  },
  {
    path: "/providers/new",
    element: <ProvidersNew />,
    errorElement: <ErrorPage />,
    loader: providersNewLoader,
    children: [
      {
        path: "aws-iam-idc",
        element: <AwsIamIdcNew />,
        action: awsIamIdcNewConfigureAction,
      },
    ],
  },
  {
    path: "/aws-iam-idc-card",
    loader: awsIamIdcCardLoader,
  },
])

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />,
  </React.StrictMode>,
)
