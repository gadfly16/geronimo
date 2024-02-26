"use strict";
window.onload = function () {
    // Attach handlers
    document.getElementById("signup-form").onsubmit = signup;
};
function signup(e) {
    const data = new FormData(e.target);
    let uws = {
        User: {
            Name: data.get("Name"),
            Email: data.get("Email")
        },
        Secret: {
            Password: data.get("Password")
        }
    };
    // Send the request
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(uws),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        }
        else {
            throw 'failed';
        }
    }).then((data) => {
        window.location.replace("login");
    }).catch((e) => { alert(e); });
    return false;
}
