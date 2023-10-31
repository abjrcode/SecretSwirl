import { redirect } from "react-router-dom"
import { Logout } from '../../../wailsjs/go/main/AuthController'
import { ListFavorites } from "../../../wailsjs/go/main/DashboardController"

export async function loader() {
  return ListFavorites()
}

export async function dashboardAction() {
  await Logout()
  return redirect("/")
}