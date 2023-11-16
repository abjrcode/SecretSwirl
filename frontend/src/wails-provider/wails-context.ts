import React from "react"
import { ShowErrorDialog } from '../../wailsjs/go/main/AppController'
import { BrowserOpenURL } from '../../wailsjs/runtime'

export const ContextValue = {
  runtime: {
    BrowserOpenURL,
    ShowErrorDialog,
  },
}

export const WailsContext = React.createContext(ContextValue)

export function useWails() {
  return React.useContext(WailsContext)
}
