import React from "react"
import { WailsContext, ContextValue } from "./wails-context"

export function WailsProvider({ children }: { children: React.ReactNode }) {
  return (
    <WailsContext.Provider value={ContextValue}>{children}</WailsContext.Provider>
  )
}
