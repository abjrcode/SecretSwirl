import { useRef } from "react"
import {
  ToastContainer,
  ToastContext,
  ToastOptions,
  ToastSpec,
} from "./toast-context"
import { Toast } from "../components/toast"

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const MaxToasts = 3
  const DefaultToastDuration = 3000
  const toastId = useRef(1)

  const toastQueue = useRef<ToastSpec[]>([])
  const toasts = useRef<React.ReactElement<ToastSpec>[]>([])
  const notifyContainer = useRef<ToastContainer | null>(null)

  function addToast(toast: ToastSpec) {
    if (toasts.current.length < MaxToasts) {
      toasts.current.push(
        <Toast
          key={toast.id}
          {...toast}
        />,
      )
    } else {
      toastQueue.current.push(toast)
    }

    if (notifyContainer.current) notifyContainer.current(toasts.current)
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

      toasts.current = toasts.current.filter((toast) => toast.key !== id)
      toasts.current.push(toastEl)
    } else {
      toasts.current = toasts.current.filter((toast) => toast.key !== id)
    }

    if (notifyContainer.current) notifyContainer.current(toasts.current)
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
        iRenderToasts: (fn) => {
          notifyContainer.current = fn
        },
        showInfo,
        showSuccess,
        showWarning,
        showError,
      }}>
      {children}
    </ToastContext.Provider>
  )
}
