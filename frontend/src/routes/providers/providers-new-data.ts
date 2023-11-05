import { ListProviders } from "../../../wailsjs/go/main/DashboardController"

export async function providersNewLoader() {
  return ListProviders()
}