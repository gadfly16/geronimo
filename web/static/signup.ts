window.onload = function() {
    // Attach handlers
    document.getElementById("signup-form")!.onsubmit = signup;
}

function signup() {
    let formData = {
        "Name": (<HTMLInputElement>document.getElementById("signup-name")).value,
        "Email": (<HTMLInputElement>document.getElementById("signup-email")).value,
        "Password": (<HTMLInputElement>document.getElementById("signup-password")).value
    }
    // Send the request
    fetch("/signup", {
        method: 'post',
        body: JSON.stringify(formData),
        mode: 'same-origin',
    }).then((response) => {
        if (response.ok) {
            return response.json();
        } else {
            throw 'failed';
        }
    }).catch((e) => { alert(e) });
    return false;
}