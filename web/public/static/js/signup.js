window.onload = function () {
    // Attach handlers
    document.getElementById("signup-form").onsubmit = signup;
};
function signup(e) {
    const data = new FormData(e.target);
    let nud = [data.get("Name"), data.get("Email"), btoa(data.get("Password"))];
    fetch("/signup", {
        method: "post",
        body: JSON.stringify(nud),
        mode: "same-origin",
    })
        .then((response) => {
        if (response.ok) {
            window.location.replace("login.html");
        }
        else {
            throw "failed";
        }
    })
        .catch((e) => {
        alert(e);
    });
    return false;
}
export {};
