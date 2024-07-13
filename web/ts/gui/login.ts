import {nodeKinds} from "../shared/common.js"

window.onload = function() {
    // Attach handlers
    document.getElementById("login-form")!.onsubmit = login;
}

function login(e: SubmitEvent) {
    const data = new FormData(e.target as HTMLFormElement)
    let userCredentials = {
        Kind: nodeKinds.User,
        Name: data.get("Email"),
        Parms: {
            DisplayName: "subidubi",
            Password: btoa(data.get("Password") as string),
        }
    }

    fetch("/login", {
        method: 'post',
        body: JSON.stringify(userCredentials),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            window.location.replace("/static/gui.html")
        } else {
            throw 'unauthorized';
        }
    }).catch((e) => { alert(e) });
    return false;
}

