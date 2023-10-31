import { ActionFunctionArgs, redirect } from 'react-router-dom';
import { IsPasswordSetup, Login, SetupMasterPassword } from '../../../wailsjs/go/main/AuthController'

export function isMasterPasswordSetup() {
  return IsPasswordSetup();
}

export async function setupMasterPasswordOrLogin({ request }: ActionFunctionArgs) {
  const formData = await request.formData()
  const updates = Object.fromEntries(formData)

  const requestedAction = updates["action"].toString()

  switch (requestedAction) {
    case "setup":
      await SetupMasterPassword(updates["masterPassword"].toString())
      return redirect("/dashboard")
    case "login": {
      const password = updates["masterPassword"].toString()
      const success = await Login(password)
      if (success) {
        return redirect("/dashboard")
      }
      return new Response(`incorrect password: ${password}`, { status: 401 })
    }
    default:
      throw new Response("[PasswordSetup/Login] Invalid action", { status: 400 })
  }
}