import { ListProviders } from "../../../wailsjs/go/main/DashboardController"

export async function providersLoader() {
  return ListProviders()
}