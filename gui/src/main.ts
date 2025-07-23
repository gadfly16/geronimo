import { $ } from "./utility"
import "./components/gui"
import { APIMsg } from "./api/common"

window.onload = () => {
  const m = APIMsg.create()
  console.log(`GUI started. User_ID="${window.UserID}"`)
  document.body.appendChild($(`<g-gui user_id=${window.UserID}></g-gui>`))
}
