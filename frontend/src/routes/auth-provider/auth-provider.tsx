import React from "react"
import { AuthContext, AuthState } from "./auth-context"

export function AuthProvider({
  initialAuthState,
  children,
}: {
  initialAuthState: AuthState
  children: React.ReactNode
}) {
  const [authState, setAuthState] = React.useState<AuthState>(initialAuthState)

  function handleConfigure() {
    const newState = {
      ...authState,
      isAuthenticated: true,
      username: "admin",
      failedAttempts: 0,
    }

    sessionStorage.setItem("auth_state", JSON.stringify(newState))
    setAuthState(newState)
  }

  function handleUnlock() {
    const newState = {
      ...authState,
      isAuthenticated: true,
      username: "admin",
      failedAttempts: 0,
    }

    sessionStorage.setItem("auth_state", JSON.stringify(newState))
    setAuthState(newState)
  }

  function handleLock() {
    const newState = {
      ...authState,
      isAuthenticated: false,
      username: "",
      failedAttempts: 0,
    }

    sessionStorage.setItem("auth_state", JSON.stringify(newState))
    setAuthState(newState)
  }

  function handleUnlockFailed() {
    const newState = {
      ...authState,
      failedAttempts: authState.failedAttempts + 1,
    }

    sessionStorage.setItem("auth_state", JSON.stringify(newState))
    setAuthState(newState)
  }

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated: authState.isAuthenticated,
        username: authState.username,
        failedAttempts: authState.failedAttempts,
        onVaultConfigured: handleConfigure,
        onVaultUnlocked: handleUnlock,
        onVaultLocked: handleLock,
        onVaultUnlockFailed: handleUnlockFailed,
      }}>
      {children}
    </AuthContext.Provider>
  )
}
