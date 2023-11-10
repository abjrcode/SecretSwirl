import React from "react"

export type AuthState = {
  isAuthenticated: boolean,
  username: string
  failedAttempts: number
}

export type AuthContextState = AuthState & {
  onVaultConfigured: () => void
  onVaultLocked: () => void
  onVaultUnlocked: () => void
  onVaultUnlockFailed: () => void
}

export const AuthContext = React.createContext<AuthContextState>(null as never)

export function useAuth() {
  return React.useContext(AuthContext)
}
