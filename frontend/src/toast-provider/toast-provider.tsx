import { useRef, useState } from "react"
import { ToastContext, ToastOptions, ToastSpec } from "./toast-context"
import { Toast } from "../components/toast"

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const MaxToasts = 3
  const DefaultToastDuration = 3000
  const toastId = useRef(1)

  const toastQueue = useRef<ToastSpec[]>([])
  const [toasts, setToasts] = useState<React.ReactElement<ToastSpec>[]>([])

  function addToast(toast: ToastSpec) {
    if (toasts.length < MaxToasts) {
      setToasts((toasts) => [
        ...toasts,
        <Toast
          key={toast.id}
          {...toast}
        />,
      ])
    } else {
      toastQueue.current.push(toast)
    }
  }

  function removeToast(id: string) {
    if (toastQueue.current.length > 0) {
      const nextInQueue = toastQueue.current.shift()!
      const toastEl = (
        <Toast
          key={nextInQueue.id}
          {...nextInQueue}
        />
      )

      setToasts((toasts) => [...toasts.filter((toast) => toast.key !== id), toastEl])
    } else {
      setToasts((toasts) => toasts.filter((toast) => toast.key !== id))
    }
  }

  function showInfo(opts: string | ToastOptions) {
    addToast({
      id: `toast-${toastId.current++}`,
      type: "info",
      message: typeof opts === "string" ? opts : opts.message,
      duration: typeof opts === "object" ? opts.duration : DefaultToastDuration,
      onClose: removeToast,
    })
  }
  function showSuccess(opts: string | ToastOptions) {
    addToast({
      id: `toast-${toastId.current++}`,
      type: "success",
      message: typeof opts === "string" ? opts : opts.message,
      duration: typeof opts === "object" ? opts.duration : DefaultToastDuration,
      onClose: removeToast,
    })
  }
  function showWarning(opts: string | ToastOptions) {
    addToast({
      id: `toast-${toastId.current++}`,
      type: "warning",
      message: typeof opts === "string" ? opts : opts.message,
      duration: typeof opts === "object" ? opts.duration : DefaultToastDuration,
      onClose: removeToast,
    })
  }
  function showError(opts: string | ToastOptions) {
    addToast({
      id: `toast-${toastId.current++}`,
      type: "error",
      message: typeof opts === "string" ? opts : opts.message,
      duration: typeof opts === "object" ? opts.duration : DefaultToastDuration,
      onClose: removeToast,
    })
  }

  return (
    <ToastContext.Provider
      value={{
        toasts: toasts,
        showInfo,
        showSuccess,
        showWarning,
        showError,
      }}>
      {children}
    </ToastContext.Provider>
  )
}
