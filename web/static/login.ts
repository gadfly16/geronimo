window.onload = function() {
    // Attach handlers
    document.getElementById("login-form")!.onsubmit = login;
}

function login() {
    let formData = {
        "Email": (<HTMLInputElement>document.getElementById("login-email")).value,
        "Password": (<HTMLInputElement>document.getElementById("login-password")).value
    }
    // Send the request
    fetch("/login", {
        method: 'post',
        body: JSON.stringify(formData),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        } else {
            throw 'unauthorized';
        }
    }).then((data) => {
        window.location.replace("/gui/user/" + data.userid)
    }).catch((e) => { alert(e) });
    return false;
}

