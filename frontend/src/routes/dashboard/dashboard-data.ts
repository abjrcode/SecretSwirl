import { ListFavorites } from "../../../wailsjs/go/main/DashboardController"
export async function dashboardLoader() {
  return ListFavorites()
}