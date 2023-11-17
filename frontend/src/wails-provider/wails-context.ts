import React from "react"
import { ShowErrorDialog, ShowWarningDialog, CatchUnhandledError } from '../../wailsjs/go/main/AppController'
import { BrowserOpenURL } from '../../wailsjs/runtime'

export const ContextValue = {
  runtime: {
    BrowserOpenURL,
    ShowWarningDialog,
    ShowErrorDialog,
    CatchUnhandledError,
  },
}

export const WailsContext = React.createContext(ContextValue)

export function useWails() {
  return React.useContext(WailsContext)
}
