import { useState, useRef, useEffect } from "react"
import { useAuth } from "../auth-provider/auth-context"

export function VaultDoor({
  verifyCombo,
}: {
  verifyCombo: (password: string) => void
}) {
  const { failedAttempts } = useAuth()
  const [loginPassword, setLoginPassword] = useState("")
  const loginBtnRef = useRef<HTMLButtonElement>(null)

  function onLoginPasswordChange(e: React.ChangeEvent<HTMLInputElement>) {
    loginBtnRef.current?.setCustomValidity("")
    setLoginPassword(e.target.value)
  }

  function onLoginBtnClicked(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
    e.preventDefault()

    verifyCombo(loginPassword)
  }

  useEffect(() => {
    if (failedAttempts > 0) {
      loginBtnRef.current?.setCustomValidity("Incorrect password")
    }
  }, [failedAttempts])

  return (
    <>
      <h1 className="text-4xl text-primary text-center">Master Password</h1>
      <form className="border-2 flex flex-col gap-4 p-4 m-4 group">
        <input
          name="masterPassword"
          type="password"
          className="input input-bordered input-primary w-96 group-invalid:border-error group-invalid:animate-shake transition-transform"
          autoComplete="off"
          onChange={onLoginPasswordChange}
        />
        <button
          ref={loginBtnRef}
          onClick={onLoginBtnClicked}
          className="btn btn-primary">
          unlock
        </button>
      </form>
    </>
  )
}
