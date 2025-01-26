import { nodeKinds } from "./common.js"

window.onload = function () {
  // Attach handlers
  document.getElementById("login-form")!.onsubmit = login
}

function login(e: SubmitEvent) {
  const fd = new FormData(e.target as HTMLFormElement)
  let aud = [fd.get("Name"), btoa(fd.get("Password") as string)]

  fetch("/login", {
    method: "post",
    body: JSON.stringify(aud),
    mode: "same-origin",
  })
    .then((response) => {
      if (response.ok) {
        window.location.replace("/gui")
      } else {
        throw "unauthorized"
      }
    })
    .catch((e) => {
      alert(e)
    })
  return false
}
