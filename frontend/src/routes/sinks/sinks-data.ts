import { Dashboard_ListSinks } from "../../utils/ipc-adapter"

export async function sinksLoader() {
  return Dashboard_ListSinks()
}