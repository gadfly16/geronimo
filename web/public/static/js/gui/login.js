import { nodeKinds } from "../shared/common.js";
window.onload = function () {
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
};
function login(e) {
    const data = new FormData(e.target);
    let userCredentials = {
        Kind: nodeKinds.User,
        Name: data.get("Email"),
        Parms: {
            Password: btoa(data.get("Password")),
        }
    };
    fetch("/login", {
        method: 'post',
        body: JSON.stringify(userCredentials),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            window.location.replace("/gui");
        }
        else {
            throw 'unauthorized';
        }
    }).catch((e) => { alert(e); });
    return false;
}
