import { Dashboard_ListFavorites } from "../../utils/ipc-adapter"

export async function dashboardLoader() {
  return Dashboard_ListFavorites()
}