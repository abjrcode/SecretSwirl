export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <div className="h-screen flex flex-col justify-center items-center gap-8">
      {children}
    </div>
  )
}
