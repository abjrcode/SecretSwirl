import React from "react"
import { ShowErrorDialog, ShowWarningDialog } from '../../wailsjs/go/main/AppController'
import { BrowserOpenURL } from '../../wailsjs/runtime'

export const ContextValue = {
  runtime: {
    BrowserOpenURL,
    ShowWarningDialog,
    ShowErrorDialog,
  },
}

export const WailsContext = React.createContext(ContextValue)

export function useWails() {
  return React.useContext(WailsContext)
}
