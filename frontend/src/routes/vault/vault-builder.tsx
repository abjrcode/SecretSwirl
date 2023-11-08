import { useState, useRef } from "react"
import { Form } from "react-router-dom"

export function VaultBuilder({
  onBuild: onSetup,
}: {
  onBuild: (password: string) => void
}) {
  const setupPasswordBtnRef = useRef<HTMLButtonElement>(null)

  const [{ masterPassword, passwordConfirmation, passwordsMatch }, setPasswords] =
    useState<{
      masterPassword: string
      passwordConfirmation: string
      passwordsMatch: boolean | null
    }>({
      masterPassword: "",
      passwordConfirmation: "",
      passwordsMatch: null,
    })

  function onPasswordChange(e: React.ChangeEvent<HTMLInputElement>) {
    const { value } = e.target

    let doPasswordsMatch: boolean | null = null

    if (value === "" || passwordConfirmation === "") {
      doPasswordsMatch = null
    } else {
      doPasswordsMatch = value === passwordConfirmation
    }

    setPasswords((state) => ({
      ...state,
      masterPassword: value,
      passwordsMatch: doPasswordsMatch,
    }))

    if (setupPasswordBtnRef.current) {
      if (doPasswordsMatch === null || doPasswordsMatch === true) {
        setupPasswordBtnRef.current.setCustomValidity("")
      } else {
        setupPasswordBtnRef.current.setCustomValidity("Passwords do not match")
      }
    }
  }

  function onConfirmationChange(e: React.ChangeEvent<HTMLInputElement>) {
    const { value } = e.target

    let doPasswordsMatch: boolean | null = null

    if (value === "" || masterPassword === "") {
      doPasswordsMatch = null
    } else {
      doPasswordsMatch = value === masterPassword
    }

    setPasswords((state) => ({
      ...state,
      passwordConfirmation: value,
      passwordsMatch: doPasswordsMatch,
    }))

    if (setupPasswordBtnRef.current) {
      if (doPasswordsMatch === null || doPasswordsMatch === true) {
        setupPasswordBtnRef.current.setCustomValidity("")
      } else {
        setupPasswordBtnRef.current.setCustomValidity("Passwords do not match")
      }
    }
  }

  function onSetupBtnClicked(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
    e.preventDefault()

    setupPasswordBtnRef.current?.setCustomValidity("")
    onSetup(masterPassword)
  }

  return (
    <div className="h-screen flex flex-col items-center justify-center gap-8">
      <h1 className="text-primary text-4xl">Master Password Setup</h1>
      <p className="text-center">
        You have not set up a master password yet! <br />
        The password you set here will be used to encrypt your data and prevent
        unauthorized access.
      </p>
      <p className="bg-warning text-warning-content p-2">
        Make sure you memorize this or save it in a safe place. <br />
        If it is lost, there is no way to recover your data!
      </p>
      <Form className="group flex flex-col gap-4 border-2 p-8">
        <label className="flex gap-4 items-center">
          <span className="w-36">Master Password</span>
          <input
            name="masterPassword"
            type="password"
            className="input input-bordered input-primary w-96"
            autoComplete="off"
            value={masterPassword}
            onChange={onPasswordChange}
            required
          />
        </label>
        <label className="flex gap-4 items-center">
          <span className="w-36">Confirm Password</span>
          <input
            name="passwordConfirmation"
            type="password"
            className="input input-bordered input-primary w-96"
            autoComplete="off"
            value={passwordConfirmation}
            onChange={onConfirmationChange}
            required
          />
        </label>
        <p
          className={
            passwordsMatch === false
              ? "p-2 text-center bg-error text-error-content"
              : passwordsMatch === true
              ? "p-2 text-center bg-success text-success-content"
              : "hidden p-2 text-center bg-neutral text-neutral-content"
          }>
          {passwordsMatch === false
            ? "Passwords do not match :("
            : passwordsMatch === true
            ? "Wohooo! Passwords match!"
            : "N/A"}
        </p>
        <button
          type="submit"
          ref={setupPasswordBtnRef}
          onClick={onSetupBtnClicked}
          className="btn btn-primary group-invalid:btn-disabled">
          READY
        </button>
      </Form>
    </div>
  )
}
