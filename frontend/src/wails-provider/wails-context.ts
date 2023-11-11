import React from "react"
import { BrowserOpenURL } from '../../wailsjs/runtime'

export const ContextValue = {
  runtime: {
    BrowserOpenURL,
  },
}

export const WailsContext = React.createContext(ContextValue)

export function useWails() {
  return React.useContext(WailsContext)
}
