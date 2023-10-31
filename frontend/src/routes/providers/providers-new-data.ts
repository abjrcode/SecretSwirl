import { ListProviders } from "../../../wailsjs/go/main/DashboardController"

export async function loader() {
  return ListProviders()
}