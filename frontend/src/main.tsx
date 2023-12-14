import "./tailwind.css"
import "./logger-adapter"
import React from "react"
import ReactDOM from "react-dom/client"
import { createHashRouter, RouterProvider } from "react-router-dom"

import { Dashboard } from "./routes/dashboard/dashboard"
import { dashboardLoader } from "./routes/dashboard/dashboard-data"
import { providersLoader } from "./routes/providers/providers-data"
import { awsIdcCardLoader } from "./routes/providers/aws-idc/card-data"
import { Providers } from "./routes/providers/providers"
import { AwsIdcSetup } from "./routes/providers/aws-idc/setup"
import { awsIdcSetupAction } from "./routes/providers/aws-idc/setup-data"
import { Vault } from "./routes/vault/vault"
import { AuthProvider } from "./auth-provider/auth-provider"
import { ErrorPage } from "./error-page"
import { WailsProvider } from "./wails-provider/wails-provider"
import { AwsIdcDeviceAuth } from "./routes/providers/aws-idc/device-auth"
import { awsIdcDeviceAuthAction } from "./routes/providers/aws-idc/device-auth-data"
import { ToastProvider } from "./toast-provider/toast-provider"
import { AwsIdcInstances } from "./routes/providers/aws-idc/instances"
import { awsIdcInstancesData } from "./routes/providers/aws-idc/instances-data"
import { Auth_IsVaultConfigured } from "./utils/ipc-adapter"
import { CompatibleSinks } from "./routes/sinks/compatible-sinks"
import { AwsCredentialsFileSetup } from "./routes/sinks/aws-credentials-file/new"
import { awsCredentialsFileSetupAction } from "./routes/sinks/aws-credentials-file/new-data"
import { awsCredentialsFileCardLoader } from "./routes/sinks/aws-credentials-file/card-data"
import { ProviderCodes, SinkCodes } from "./utils/provider-sink-codes"
import { compatibleSinksLoader } from "./routes/sinks/compatible-sinks-data"

const devMode = import.meta.env.DEV

if (devMode) {
  document.documentElement.classList.add("debug-screens")
}

console.log("starting frontend application ...")

void (async function main() {
  const router = createHashRouter([
    {
      element: <Vault isVaultConfigured={await Auth_IsVaultConfigured()} />,
      errorElement: <ErrorPage devMode />,
      children: [
        {
          path: "/",
          element: <Dashboard />,
          loader: dashboardLoader,
        },
        {
          path: "/providers",
          element: <Providers />,
          loader: providersLoader,
          children: [
            {
              path: ProviderCodes.AwsIdc,
              children: [
                {
                  index: true,
                  element: <AwsIdcInstances />,
                  loader: awsIdcInstancesData,
                },
                {
                  path: "setup",
                  element: <AwsIdcSetup />,
                  action: awsIdcSetupAction,
                },
                {
                  path: "device-auth",
                  element: <AwsIdcDeviceAuth />,
                  action: awsIdcDeviceAuthAction,
                },
              ],
            },
          ],
        },
        {
          path: "/sinks",
          element: <CompatibleSinks />,
          loader: compatibleSinksLoader,
          children: [
            {
              path: SinkCodes.AwsCredentialsFile,
              children: [
                {
                  path: "setup/:providerCode/:instanceId",
                  element: <AwsCredentialsFileSetup />,
                  action: awsCredentialsFileSetupAction,
                },
              ],
            },
          ],
        },
      ],
    },
    {
      path: "/internal",
      children: [
        {
          path: "api",
          children: [
            {
              path: `${ProviderCodes.AwsIdc}-card`,
              loader: awsIdcCardLoader,
            },
            {
              path: `${SinkCodes.AwsCredentialsFile}-card`,
              loader: awsCredentialsFileCardLoader,
            },
          ],
        },
      ],
    },
  ])

  const sessionState = sessionStorage.getItem("auth_state")
  const initialState = sessionState
    ? JSON.parse(sessionState)
    : {
        isAuthenticated: false,
        username: "",
        failedAttempts: 0,
      }

  ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
      <AuthProvider initialAuthState={initialState}>
        <WailsProvider>
          <ToastProvider>
            <RouterProvider router={router} />
          </ToastProvider>
        </WailsProvider>
      </AuthProvider>
    </React.StrictMode>,
  )
})()
