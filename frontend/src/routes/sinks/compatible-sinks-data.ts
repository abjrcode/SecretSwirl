import { Dashboard_ListCompatibleSinks } from "../../utils/ipc-adapter"

export async function compatibleSinksLoader({ request }: { request: Request }) {
  const url = new URL(request.url);
  const providerCode = url.searchParams.get("providerCode");

  if (!providerCode) {
    return new Response("Missing providerCode", { status: 400 })
  }

  return Dashboard_ListCompatibleSinks(providerCode)
}