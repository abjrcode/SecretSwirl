import "./tailwind.css"
import Alpine from "alpinejs"

window.Alpine = Alpine

Alpine.store("htmlClass", () => {
  if (import.meta.env.DEV) {
    return "debug-screens"
  }
  return ""
})

Alpine.start()

export { }