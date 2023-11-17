import React from "react";
import { createContext } from "react";

export type ToastSpec = {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  duration: number;
  onClose(toastId: string): void;
}

export type ToastOptions = {
  message: string;
  duration: number;
}

export type ToastContainer = (toasts: React.ReactElement[]) => void

type ToastContext = {
  iRenderToasts(fn: ToastContainer): void;
  showInfo(opts: string | ToastOptions): void;
  showSuccess(opts: string | ToastOptions): void;
  showWarning(opts: string | ToastOptions): void;
  showError(opts: string | ToastOptions): void;
}

export const ToastContext = createContext<ToastContext>(null as never)

export function useToaster() {
  return React.useContext(ToastContext)
}