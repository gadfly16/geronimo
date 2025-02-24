window.onload = function () {
    // Attach handlers
    document.getElementById("login-form").onsubmit = login;
};
function login(e) {
    const fd = new FormData(e.target);
    let aud = [fd.get("Name"), fd.get("Password")];
    fetch("/login", {
        method: "post",
        body: JSON.stringify(aud),
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
export {};
