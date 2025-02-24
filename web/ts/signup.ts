import { nodeKinds } from "./common.js"

window.onload = function () {
  // Attach handlers
  document.getElementById("signup-form")!.onsubmit = signup
}

function signup(e: SubmitEvent) {
  const fd = new FormData(e.target as HTMLFormElement)
  let nud = [fd.get("Name"), fd.get("Email"), fd.get("Password")]

  fetch("/signup", {
    method: "post",
    body: JSON.stringify(nud),
    mode: "same-origin",
  })
    .then((response) => {
      if (response.ok) {
        window.location.replace("login.html")
      } else {
        throw "failed"
      }
    })
    .catch((e) => {
      alert(e)
    })
  return false
}
