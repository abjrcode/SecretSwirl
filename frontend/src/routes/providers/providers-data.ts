import { Dashboard_ListProviders } from "../../utils/ipc-adapter";

export async function providersLoader() {
  return Dashboard_ListProviders()
}