import { useEffect, useState } from "react"
import { useToaster } from "./toast-provider/toast-context"
import { useAuth } from "./auth-provider/auth-context"
import { LockVault } from "../wailsjs/go/main/AuthController"
import { Link, useNavigate } from "react-router-dom"

export function Layout({ children }: { children: React.ReactNode }) {
  const toaster = useToaster()
  const authContext = useAuth()
  const navigate = useNavigate()

  const [toasts, setToasts] = useState<React.ReactElement[]>([])

  function updateToasts(toasts: React.ReactElement[]) {
    setToasts([...toasts])
  }

  useEffect(() => {
    toaster.iRenderToasts(updateToasts)
  }, [toaster])

  async function attemptLock() {
    await LockVault()

    navigate("/")
    authContext.onVaultLocked()
  }

  return (
    <>
      <div
        id="toastContainer"
        className="fixed right-5 top-5 w-[512px] z-50">
        {...toasts}
      </div>
      <div className="drawer drawer-open">
        <input
          id="my-drawer"
          type="checkbox"
          className="drawer-toggle"
        />
        <main className="drawer-content p-4">{children}</main>
        <div className="drawer-side">
          <label
            htmlFor="my-drawer"
            aria-label="close sidebar"
            className="drawer-overlay"></label>
          <div className="flex flex-col p-4 w-48 min-h-full bg-base-200 text-base-content">
            <nav className="flex-1 flex flex-col gap-2">
              <Link
                to="/"
                className="btn btn-primary btn-outline capitalize">
                dashboard
              </Link>
              <Link
                to="/providers"
                className="btn btn-primary btn-outline capitalize">
                providers
              </Link>
            </nav>

            <button
              onClick={attemptLock}
              className="btn btn-secondary btn-outline uppercase">
              lock
            </button>
          </div>
        </div>
      </div>
    </>
  )
}
