import {nodeKinds} from "../shared/common.js"

window.onload = function() {
    // Attach handlers
    document.getElementById("signup-form")!.onsubmit = signup;
}

function signup(e: SubmitEvent) {
    const data = new FormData(<HTMLFormElement>e.target)
    let newUser = {
        Kind: nodeKinds.User,
        Name: data.get("Email"),
        Parms: {
            DisplayName: data.get("Name"),
            Password: data.get("Password"),
        }
    }
 
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(newUser),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json()
        } else {
            throw 'failed'
        }
    }).then((data) => {
        window.location.replace("login")
    }).catch((e) => { alert(e) })
    return false
}