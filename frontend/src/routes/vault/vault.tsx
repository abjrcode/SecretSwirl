import { Outlet } from "react-router-dom"
import { ConfigureVault, UnlockVault } from "../../../wailsjs/go/main/AuthController"
import { Layout } from "../../layout"
import { useAuth } from "../../auth-provider/auth-context"
import { VaultBuilder } from "./vault-builder"
import { VaultDoor } from "./vault-door"
import { useState } from "react"

export function Vault(props: { isVaultConfigured: boolean }) {
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

  if (authContext.isAuthenticated === false) {
    if (isVaultConfigured) {
      return (
        <div className="h-screen flex flex-col justify-center items-center gap-8">
          <VaultDoor verifyCombo={attemptUnlock} />
        </div>
      )
    }

    return (
      <div className="h-screen flex flex-col justify-center items-center gap-8">
        <VaultBuilder onBuild={buildVault} />
      </div>
    )
  }

  return (
    <Layout>
      <Outlet />
    </Layout>
  )
}
