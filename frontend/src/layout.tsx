import { useToaster } from "./toast-provider/toast-context"

export function Layout({ children }: { children: React.ReactNode }) {
  const toaster = useToaster()

  return (
    <>
      <div
        id="toastContainer"
        className="fixed left-5 top-5 w-[512px] z-50">
        {...toaster.toasts}
      </div>
      <div className="h-screen flex flex-col justify-center items-center gap-8">
        {children}
      </div>
    </>
  )
}
