import {nodeKinds} from "../shared/common.js"

window.onload = function() {
    // Attach handlers
    document.getElementById("signup-form")!.onsubmit = signup;
}

function signup(e: SubmitEvent) {
    const data = new FormData(e.target as HTMLFormElement)
    let newUser = {
        Kind: nodeKinds.User,
        Name: data.get("Email"),
        Parms: {
            DisplayName: data.get("Name"),
            Password: btoa(data.get("Password") as string),
        }
    }
 
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(newUser),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            window.location.replace("login.html")
        } else {
            throw 'failed'
        }
    }).catch((e) => { alert(e) })
    return false
}