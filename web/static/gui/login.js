"use strict";
window.onload = function () {
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
};
function login(e) {
    const data = new FormData(e.target);
    let user = {
        Email: data.get("Email"),
        Password: data.get("Password"),
    };
    fetch("/login", {
        method: 'post',
        body: JSON.stringify(user),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        }
        else {
            throw 'unauthorized';
        }
    }).then((data) => {
        window.location.replace("/gui" + new URL(location.href).search);
    }).catch((e) => { alert(e); });
    return false;
}
