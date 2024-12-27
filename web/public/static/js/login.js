import { nodeKinds } from "./common.js";
window.onload = function () {
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
};
function login(e) {
    const fd = new FormData(e.target);
    let ucn = {
        Kind: nodeKinds.User,
        Name: fd.get("Name"),
        Parms: {
            Password: btoa(fd.get("Password")),
        },
    };
    fetch("/login", {
        method: "post",
        body: JSON.stringify(ucn),
        mode: "same-origin",
    })
        .then((response) => {
        if (response.ok) {
            window.location.replace("/gui");
        }
        else {
            throw "unauthorized";
        }
    })
        .catch((e) => {
        alert(e);
    });
    return false;
}
