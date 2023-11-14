import { useCallback, useEffect, useRef, useState, useTransition } from "react"

function ToastCloseButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      className="place-self-start justify-self-start"
      onClick={onClick}>
      <svg
        className="fill-current h-5 w-5"
        role="button"
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 20 20">
        <title>Close</title>
        <path d="M14.348 14.849a1.2 1.2 0 0 1-1.697 0L10 11.819l-2.651 3.029a1.2 1.2 0 1 1-1.697-1.697l2.758-3.15-2.759-3.152a1.2 1.2 0 1 1 1.697-1.697L10 8.183l2.651-3.031a1.2 1.2 0 1 1 1.697 1.697l-2.758 3.152 2.758 3.15a1.2 1.2 0 0 1 0 1.698z" />
      </svg>
    </button>
  )
}

export function Toast({
  id,
  type,
  message,
  duration,
  onClose,
}: {
  id: string
  type: "info" | "success" | "warning" | "error"
  message: string
  duration: number
  onClose: (toastId: string) => void
}) {
  const [, startTransition] = useTransition()

  // Can be replaced in the future with synthetic events that trigger native dom animations
  const [isExiting, setIsExiting] = useState(false)

  const isEntering = useRef(true)

  // animation times need to stay in-sync but tailwind
  // will not be able to extract class names if we use variables
  const AnimationDuration = 300

  const animationContainerClass = `transition-max-height duration-300 ease-out mb-4 ${
    isExiting ? "max-h-0" : "max-h-[300px]"
  }`

  const toastAnimationClass = `${
    isEntering.current ? "animate-[slideInLeft_0.5s_ease-in_both]" : ""
  } ${isExiting ? "animate-[slideOutLeft_0.3s_ease-out_both]" : ""}`

  const exitAnimationTimer = useRef<number | null>(null)
  const toastRemovalTimer = useRef<number | null>(null)

  function closeToast() {
    setIsExiting(true)
  }

  const scheduleExitAnimation = useCallback(
    function () {
      exitAnimationTimer.current = setTimeout(() => {
        startTransition(() => {
          setIsExiting(true)
        })
      }, duration)
    },
    [duration],
  )

  const scheduleToastRemoval = useCallback(
    function () {
      exitAnimationTimer.current = setTimeout(() => {
        startTransition(() => {
          onClose(id)
        })
      }, AnimationDuration)
    },
    [id, onClose],
  )

  useEffect(() => {
    isEntering.current = false

    if (isExiting) {
      scheduleToastRemoval()
    } else {
      scheduleExitAnimation()
    }

    return () => {
      if (exitAnimationTimer.current) {
        clearTimeout(exitAnimationTimer.current)
        exitAnimationTimer.current = null
      }

      if (toastRemovalTimer.current) {
        clearTimeout(toastRemovalTimer.current)
        toastRemovalTimer.current = null
      }
    }
  }, [scheduleExitAnimation, scheduleToastRemoval, isExiting])

  function pauseExit() {
    if (exitAnimationTimer.current) {
      clearTimeout(exitAnimationTimer.current)
      exitAnimationTimer.current = null
    }

    if (toastRemovalTimer.current) {
      clearTimeout(toastRemovalTimer.current)
      toastRemovalTimer.current = null
    }
  }

  function resumeExit() {
    if (isExiting) {
      scheduleToastRemoval()
    } else {
      scheduleExitAnimation()
    }
  }

  switch (type) {
    case "info":
      return (
        <div
          onMouseOver={pauseExit}
          onMouseLeave={resumeExit}
          className={animationContainerClass}>
          <div
            role="alert"
            className={`alert alert-info text-info-content overflow-hidden rounded-sm shadow-md ${toastAnimationClass}`}>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              className="stroke-current shrink-0 w-6 h-6">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{message}</span>

            <ToastCloseButton onClick={closeToast} />
          </div>
        </div>
      )
    case "success":
      return (
        <div
          onMouseOver={pauseExit}
          onMouseLeave={resumeExit}
          className={animationContainerClass}>
          <div
            role="alert"
            className={`alert alert-success text-success-content overflow-hidden rounded-sm shadow-md ${toastAnimationClass}`}>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>{message}</span>

            <ToastCloseButton onClick={closeToast} />
          </div>
        </div>
      )
    case "warning":
      return (
        <div
          onMouseOver={pauseExit}
          onMouseLeave={resumeExit}
          className={animationContainerClass}>
          <div
            role="alert"
            className={`alert alert-warning text-warning-content overflow-hidden rounded-sm shadow-md ${toastAnimationClass}`}>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
            <span>{message}</span>

            <ToastCloseButton onClick={closeToast} />
          </div>
        </div>
      )
    case "error":
      return (
        <div
          onMouseOver={pauseExit}
          onMouseLeave={resumeExit}
          className={animationContainerClass}>
          <div
            role="alert"
            className={`alert alert-error text-error-content overflow-hidden rounded-sm shadow-md ${toastAnimationClass}`}>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="stroke-current shrink-0 h-6 w-6"
              fill="none"
              viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth="2"
                d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z"
              />
            </svg>
            <span>{message}</span>

            <ToastCloseButton onClick={closeToast} />
          </div>
        </div>
      )
    default:
      return null
  }
}
