import { Outlet, useNavigate } from "react-router-dom"
import {
  ConfigureVault,
  UnlockVault,
  LockVault,
} from "../../../wailsjs/go/main/AuthController"
import { Layout } from "../../layout"
import { useAuth } from "../../auth-provider/auth-context"
import { VaultBuilder } from "./vault-builder"
import { VaultDoor } from "./vault-door"
import { useState } from "react"

export function Vault(props: { isVaultConfigured: boolean }) {
  const navigate = useNavigate()
  const authContext = useAuth()

  const [isVaultConfigured, setIsVaultConfigured] = useState(props.isVaultConfigured)

  async function buildVault(password: string) {
    await ConfigureVault(password)
    setIsVaultConfigured(true)
    authContext.onVaultConfigured()
  }

  async function attemptUnlock(password: string) {
    const success = await UnlockVault(password)
    if (success) {
      authContext.onVaultUnlocked()
    } else {
      authContext.onVaultUnlockFailed()
    }
  }

  async function attemptLock() {
    await LockVault()

    navigate("/")
    authContext.onVaultLocked()
  }

  if (authContext.isAuthenticated === false) {
    if (isVaultConfigured) {
      return (
        <Layout>
          <VaultDoor verifyCombo={attemptUnlock} />
        </Layout>
      )
    }

    return (
      <Layout>
        <VaultBuilder onBuild={buildVault} />
      </Layout>
    )
  }

  return (
    <>
      <button
        onClick={attemptLock}
        className="fixed top-5 right-5 btn btn-secondary btn-outline">
        lock
      </button>
      <Layout>
        <Outlet />
      </Layout>
    </>
  )
}
